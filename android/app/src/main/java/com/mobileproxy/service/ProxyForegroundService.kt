package com.mobileproxy.service

import android.app.Notification
import android.app.PendingIntent
import android.app.Service
import android.content.Intent
import android.os.IBinder
import android.os.PowerManager
import android.util.Log
import androidx.core.app.NotificationCompat
import com.mobileproxy.MobileProxyApp
import com.mobileproxy.R
import com.mobileproxy.core.network.NetworkManager
import com.mobileproxy.core.proxy.HttpProxyServer
import com.mobileproxy.core.proxy.Socks5ProxyServer
import com.mobileproxy.core.status.DeviceStatusReporter
import com.mobileproxy.ui.MainActivity
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class ProxyForegroundService : Service() {

    companion object {
        private const val TAG = "ProxyForegroundService"
        const val ACTION_START = "com.mobileproxy.START"
        const val ACTION_STOP = "com.mobileproxy.STOP"
        const val EXTRA_SERVER_URL = "server_url"
        const val EXTRA_DEVICE_ID = "device_id"
        const val EXTRA_AUTH_TOKEN = "auth_token"
    }

    @Inject lateinit var networkManager: NetworkManager
    @Inject lateinit var httpProxy: HttpProxyServer
    @Inject lateinit var socks5Proxy: Socks5ProxyServer
    @Inject lateinit var statusReporter: DeviceStatusReporter

    private var wakeLock: PowerManager.WakeLock? = null

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_START -> {
                val serverUrl = intent.getStringExtra(EXTRA_SERVER_URL) ?: ""
                val deviceId = intent.getStringExtra(EXTRA_DEVICE_ID) ?: ""
                val authToken = intent.getStringExtra(EXTRA_AUTH_TOKEN) ?: ""
                startProxy(serverUrl, deviceId, authToken)
            }
            ACTION_STOP -> stopProxy()
        }
        return START_STICKY
    }

    private fun startProxy(serverUrl: String, deviceId: String, authToken: String) {
        Log.i(TAG, "Starting proxy service")

        // Acquire wake lock
        val powerManager = getSystemService(POWER_SERVICE) as PowerManager
        wakeLock = powerManager.newWakeLock(
            PowerManager.PARTIAL_WAKE_LOCK,
            "MobileProxy::ProxyWakeLock"
        ).apply { acquire() }

        // Start foreground notification
        startForeground(MobileProxyApp.NOTIFICATION_ID, createNotification())

        // Acquire both networks (WiFi Split)
        networkManager.acquireNetworks()

        // Start proxy servers
        httpProxy.start(8080)
        socks5Proxy.start(1080)

        // Start heartbeat reporting
        statusReporter.start(serverUrl, deviceId, authToken)
    }

    private fun stopProxy() {
        Log.i(TAG, "Stopping proxy service")
        statusReporter.stop()
        httpProxy.stop()
        socks5Proxy.stop()
        networkManager.releaseNetworks()
        wakeLock?.release()
        stopForeground(STOP_FOREGROUND_REMOVE)
        stopSelf()
    }

    private fun createNotification(): Notification {
        val pendingIntent = PendingIntent.getActivity(
            this, 0,
            Intent(this, MainActivity::class.java),
            PendingIntent.FLAG_IMMUTABLE
        )

        return NotificationCompat.Builder(this, MobileProxyApp.NOTIFICATION_CHANNEL_ID)
            .setContentTitle("MobileProxy Active")
            .setContentText("Proxy server running on cellular network")
            .setSmallIcon(android.R.drawable.ic_menu_share)
            .setContentIntent(pendingIntent)
            .setOngoing(true)
            .build()
    }

    override fun onDestroy() {
        stopProxy()
        super.onDestroy()
    }
}
