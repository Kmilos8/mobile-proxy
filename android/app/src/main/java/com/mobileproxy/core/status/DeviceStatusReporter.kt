package com.mobileproxy.core.status

import android.content.Context
import android.os.BatteryManager
import android.os.Build
import android.telephony.TelephonyManager
import android.util.Log
import com.google.gson.Gson
import com.mobileproxy.core.commands.CommandExecutor
import com.mobileproxy.core.commands.DeviceCommand
import com.mobileproxy.core.network.NetworkManager
import com.mobileproxy.core.proxy.HttpProxyServer
import com.mobileproxy.core.proxy.ProxyCredentialStore
import com.mobileproxy.core.proxy.Socks5ProxyServer
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.*
import java.io.OutputStreamWriter
import java.net.HttpURLConnection
import java.net.URL
import javax.inject.Inject
import javax.inject.Singleton

data class ProxyCredentialResponse(val username: String = "", val password: String = "")
data class HeartbeatResponse(
    val commands: List<DeviceCommand> = emptyList(),
    val credentials: List<ProxyCredentialResponse> = emptyList()
)

data class HeartbeatPayload(
    val cellular_ip: String,
    val wifi_ip: String,
    val carrier: String,
    val network_type: String,
    val battery_level: Int,
    val battery_charging: Boolean,
    val signal_strength: Int,
    val app_version: String,
    val bytes_in: Long,
    val bytes_out: Long
)

@Singleton
class DeviceStatusReporter @Inject constructor(
    @ApplicationContext private val context: Context,
    private val networkManager: NetworkManager,
    private val httpProxy: HttpProxyServer,
    private val socks5Proxy: Socks5ProxyServer,
    private val commandExecutor: CommandExecutor,
    private val credentialStore: ProxyCredentialStore
) {
    companion object {
        private const val TAG = "StatusReporter"
        private const val HEARTBEAT_INTERVAL = 30_000L // 30 seconds
    }

    private val gson = Gson()
    private var scope: CoroutineScope? = null
    private var serverUrl: String = ""
    private var deviceId: String = ""
    private var authToken: String = ""

    fun start(serverUrl: String, deviceId: String, authToken: String) {
        this.serverUrl = serverUrl
        this.deviceId = deviceId
        this.authToken = authToken

        scope = CoroutineScope(Dispatchers.IO + SupervisorJob())
        scope?.launch {
            while (isActive) {
                sendHeartbeat()
                delay(HEARTBEAT_INTERVAL)
            }
        }
    }

    fun stop() {
        scope?.cancel()
        scope = null
    }

    private suspend fun sendHeartbeat() {
        try {
            val batteryManager = context.getSystemService(Context.BATTERY_SERVICE) as BatteryManager
            val telephonyManager = context.getSystemService(Context.TELEPHONY_SERVICE) as TelephonyManager

            val payload = HeartbeatPayload(
                cellular_ip = getCellularIp(),
                wifi_ip = getWifiIp(),
                carrier = telephonyManager.networkOperatorName ?: "Unknown",
                network_type = getNetworkTypeString(telephonyManager),
                battery_level = batteryManager.getIntProperty(BatteryManager.BATTERY_PROPERTY_CAPACITY),
                battery_charging = batteryManager.isCharging,
                signal_strength = 0, // Requires PhoneStateListener
                app_version = context.packageManager.getPackageInfo(context.packageName, 0).versionName ?: "1.0.0",
                bytes_in = httpProxy.bytesIn + socks5Proxy.bytesIn,
                bytes_out = httpProxy.bytesOut + socks5Proxy.bytesOut
            )

            val url = URL("$serverUrl/api/devices/$deviceId/heartbeat")

            // Bind to WiFi for API communication (through VPN tunnel)
            val wifiNetwork = networkManager.getWifiNetwork()
            val connection = if (wifiNetwork != null) {
                wifiNetwork.openConnection(url) as HttpURLConnection
            } else {
                url.openConnection() as HttpURLConnection
            }
            connection.apply {
                requestMethod = "POST"
                setRequestProperty("Content-Type", "application/json")
                setRequestProperty("Authorization", "Bearer $authToken")
                doOutput = true
                connectTimeout = 10000
                readTimeout = 10000
            }

            OutputStreamWriter(connection.outputStream).use { writer ->
                writer.write(gson.toJson(payload))
            }

            val responseCode = connection.responseCode
            if (responseCode == 200) {
                // Parse response for pending commands
                val response = connection.inputStream.bufferedReader().readText()
                Log.d(TAG, "Heartbeat sent successfully")

                val heartbeatResponse = gson.fromJson(response, HeartbeatResponse::class.java)
                heartbeatResponse?.commands?.forEach { command ->
                    scope?.launch {
                        val result = commandExecutor.execute(command)
                        reportCommandResult(command.id, result)
                    }
                }

                // Sync proxy credentials
                heartbeatResponse?.credentials?.let { creds ->
                    if (creds.isNotEmpty()) {
                        credentialStore.update(creds.map { it.username to it.password })
                        Log.d(TAG, "Synced ${creds.size} proxy credentials")
                    }
                }
            } else {
                Log.w(TAG, "Heartbeat failed: $responseCode")
            }
        } catch (e: Exception) {
            Log.e(TAG, "Heartbeat error", e)
        }
    }

    private suspend fun reportCommandResult(commandId: String, result: Result<String>) {
        try {
            val status = if (result.isSuccess) "completed" else "failed"
            val message = result.getOrElse { it.message ?: "Unknown error" }
            val body = gson.toJson(mapOf("status" to status, "result" to message))

            val url = URL("$serverUrl/api/devices/$deviceId/commands/$commandId/result")

            val wifiNetwork = networkManager.getWifiNetwork()
            val connection = if (wifiNetwork != null) {
                wifiNetwork.openConnection(url) as HttpURLConnection
            } else {
                url.openConnection() as HttpURLConnection
            }
            connection.apply {
                requestMethod = "POST"
                setRequestProperty("Content-Type", "application/json")
                setRequestProperty("Authorization", "Bearer $authToken")
                doOutput = true
                connectTimeout = 10000
                readTimeout = 10000
            }

            OutputStreamWriter(connection.outputStream).use { writer ->
                writer.write(body)
            }

            val responseCode = connection.responseCode
            if (responseCode == 200) {
                Log.d(TAG, "Command result reported: $commandId -> $status")
            } else {
                Log.w(TAG, "Failed to report command result: $responseCode")
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error reporting command result for $commandId", e)
        }
    }

    /**
     * Handle a command pushed through the VPN tunnel (instant delivery).
     * Parses JSON, executes, and reports result back to server.
     */
    fun handlePushedCommand(commandJson: String) {
        try {
            val command = gson.fromJson(commandJson, DeviceCommand::class.java)
            Log.i(TAG, "Executing pushed command: ${command.type} (${command.id})")
            scope?.launch {
                val result = commandExecutor.execute(command)
                reportCommandResult(command.id, result)
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to handle pushed command", e)
        }
    }

    @Volatile var cachedPublicIp: String = ""
    @Volatile var lastIpLookupTime: Long = 0
    private val IP_CACHE_DURATION = 60_000L // Re-check every 60s

    fun invalidateIpCache() {
        cachedPublicIp = ""
        lastIpLookupTime = 0
    }

    private fun getCellularIp(): String {
        val network = networkManager.getCellularNetwork() ?: return ""

        // Use cached value if fresh enough
        val now = System.currentTimeMillis()
        if (cachedPublicIp.isNotEmpty() && (now - lastIpLookupTime) < IP_CACHE_DURATION) {
            return cachedPublicIp
        }

        // Query external service through the cellular network to get the real public IP
        try {
            val url = URL("https://api.ipify.org")
            val connection = network.openConnection(url) as HttpURLConnection
            connection.connectTimeout = 5000
            connection.readTimeout = 5000
            val ip = connection.inputStream.bufferedReader().readText().trim()
            connection.disconnect()
            if (ip.isNotEmpty()) {
                cachedPublicIp = ip
                lastIpLookupTime = now
                return ip
            }
        } catch (e: Exception) {
            Log.w(TAG, "External IP lookup failed, using interface IP", e)
        }

        // Fallback: local interface IP
        val connectivityManager = context.getSystemService(Context.CONNECTIVITY_SERVICE) as android.net.ConnectivityManager
        val linkProperties = connectivityManager.getLinkProperties(network) ?: return ""
        return linkProperties.linkAddresses
            .firstOrNull { it.address is java.net.Inet4Address }
            ?.address?.hostAddress ?: ""
    }

    private fun getWifiIp(): String {
        val network = networkManager.getWifiNetwork() ?: return ""
        val connectivityManager = context.getSystemService(Context.CONNECTIVITY_SERVICE) as android.net.ConnectivityManager
        val linkProperties = connectivityManager.getLinkProperties(network) ?: return ""
        return linkProperties.linkAddresses
            .firstOrNull { it.address is java.net.Inet4Address }
            ?.address?.hostAddress ?: ""
    }

    private fun getNetworkTypeString(tm: TelephonyManager): String {
        return try {
            when (tm.dataNetworkType) {
                TelephonyManager.NETWORK_TYPE_LTE -> "4G"
                TelephonyManager.NETWORK_TYPE_NR -> "5G"
                TelephonyManager.NETWORK_TYPE_HSPAP -> "3G+"
                TelephonyManager.NETWORK_TYPE_HSPA -> "3G"
                TelephonyManager.NETWORK_TYPE_UMTS -> "3G"
                TelephonyManager.NETWORK_TYPE_EDGE -> "2G"
                else -> "Unknown"
            }
        } catch (e: SecurityException) {
            Log.w(TAG, "Missing READ_PHONE_STATE permission for network type")
            "Unknown"
        }
    }
}
