package com.mobileproxy.service

import android.content.Intent
import android.net.VpnService
import android.util.Log
import com.mobileproxy.core.network.NetworkManager
import com.mobileproxy.core.vpn.VpnTunnelManager
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import java.net.DatagramSocket
import java.net.Socket

class ProxyVpnService : VpnService() {
    companion object {
        private const val TAG = "ProxyVpnService"
        const val ACTION_START = "com.mobileproxy.VPN_START"
        const val ACTION_STOP = "com.mobileproxy.VPN_STOP"
        const val EXTRA_SERVER_IP = "server_ip"
        const val EXTRA_DEVICE_ID = "device_id"

        @Volatile
        var instance: ProxyVpnService? = null
            private set

        fun protectSocket(socket: Socket): Boolean {
            return instance?.protect(socket) ?: false
        }

        fun protectSocket(socket: DatagramSocket): Boolean {
            return instance?.protect(socket) ?: false
        }

        private val _vpnState = MutableStateFlow(false)
        val vpnState: StateFlow<Boolean> = _vpnState

        /**
         * Called by VpnTunnelManager after a successful reconnect.
         * Updates the VPN state flow so observers know we're back online.
         */
        fun onReconnected() {
            Log.i(TAG, "VPN tunnel reconnected")
            _vpnState.value = true
        }

        // Callback for commands pushed through VPN tunnel (set by ProxyForegroundService)
        var commandCallback: ((String) -> Unit)? = null

        // NetworkManager for IP forwarding (set by ProxyForegroundService before VPN start)
        var networkManagerRef: NetworkManager? = null
    }

    private var tunnelManager: VpnTunnelManager? = null
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_START -> {
                val serverIP = intent.getStringExtra(EXTRA_SERVER_IP) ?: ""
                val deviceId = intent.getStringExtra(EXTRA_DEVICE_ID) ?: ""
                if (serverIP.isNotEmpty() && deviceId.isNotEmpty()) {
                    startVpn(serverIP, deviceId)
                } else {
                    Log.e(TAG, "Missing server_ip or device_id, cannot start VPN")
                }
            }
            ACTION_STOP -> {
                stopVpn()
            }
        }
        return START_STICKY
    }

    private fun startVpn(serverIP: String, deviceId: String) {
        // Prevent double-start: disconnect old tunnel first
        tunnelManager?.let { old ->
            Log.w(TAG, "Disconnecting existing tunnel before starting new one")
            old.disconnect()
        }

        Log.i(TAG, "Starting VPN tunnel to $serverIP for device $deviceId")
        instance = this

        val manager = VpnTunnelManager(
            vpnService = this,
            serverAddress = serverIP,
            serverPort = 1194,
            deviceId = deviceId,
            networkManager = networkManagerRef
        )
        manager.commandListener = { json -> commandCallback?.invoke(json) }
        tunnelManager = manager

        scope.launch {
            val success = manager.connect()
            _vpnState.value = success
            if (success) {
                Log.i(TAG, "VPN tunnel connected, IP: ${manager.vpnIP}")
            } else {
                Log.e(TAG, "VPN tunnel connection failed")
            }
        }
    }

    private fun stopVpn() {
        Log.i(TAG, "Stopping VPN tunnel")
        tunnelManager?.disconnect()
        tunnelManager = null
        _vpnState.value = false
        instance = null
        stopSelf()
    }

    override fun onDestroy() {
        Log.i(TAG, "VPN service destroyed")
        tunnelManager?.disconnect()
        tunnelManager = null
        _vpnState.value = false
        instance = null
        scope.cancel()
        super.onDestroy()
    }

    override fun onRevoke() {
        Log.w(TAG, "VPN permission revoked")
        stopVpn()
        super.onRevoke()
    }
}
