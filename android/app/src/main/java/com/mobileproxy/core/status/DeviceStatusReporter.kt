package com.mobileproxy.core.status

import android.content.Context
import android.os.BatteryManager
import android.os.Build
import android.telephony.TelephonyManager
import android.util.Log
import com.google.gson.Gson
import com.mobileproxy.core.network.NetworkManager
import com.mobileproxy.core.proxy.HttpProxyServer
import com.mobileproxy.core.proxy.Socks5ProxyServer
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.*
import java.io.OutputStreamWriter
import java.net.HttpURLConnection
import java.net.URL
import javax.inject.Inject
import javax.inject.Singleton

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
    private val socks5Proxy: Socks5ProxyServer
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
            val connection = (url.openConnection() as HttpURLConnection).apply {
                requestMethod = "POST"
                setRequestProperty("Content-Type", "application/json")
                setRequestProperty("Authorization", "Bearer $authToken")
                doOutput = true
                connectTimeout = 10000
                readTimeout = 10000
            }

            // Bind to WiFi for API communication (through VPN tunnel)
            networkManager.getWifiNetwork()?.let { wifi ->
                wifi.bindSocket(java.net.Socket().apply {
                    connect(java.net.InetSocketAddress(url.host, url.port.let { if (it == -1) 443 else it }))
                })
            }

            OutputStreamWriter(connection.outputStream).use { writer ->
                writer.write(gson.toJson(payload))
            }

            val responseCode = connection.responseCode
            if (responseCode == 200) {
                // Parse response for pending commands
                val response = connection.inputStream.bufferedReader().readText()
                Log.d(TAG, "Heartbeat sent successfully")
                // TODO: Process commands from response
            } else {
                Log.w(TAG, "Heartbeat failed: $responseCode")
            }
        } catch (e: Exception) {
            Log.e(TAG, "Heartbeat error", e)
        }
    }

    private fun getCellularIp(): String {
        val network = networkManager.getCellularNetwork() ?: return ""
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
        return when (tm.dataNetworkType) {
            TelephonyManager.NETWORK_TYPE_LTE -> "4G"
            TelephonyManager.NETWORK_TYPE_NR -> "5G"
            TelephonyManager.NETWORK_TYPE_HSPAP -> "3G+"
            TelephonyManager.NETWORK_TYPE_HSPA -> "3G"
            TelephonyManager.NETWORK_TYPE_UMTS -> "3G"
            TelephonyManager.NETWORK_TYPE_EDGE -> "2G"
            else -> "Unknown"
        }
    }
}
