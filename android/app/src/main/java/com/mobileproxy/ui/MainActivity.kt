package com.mobileproxy.ui

import android.content.Intent
import android.net.VpnService
import android.os.Bundle
import android.widget.Button
import android.widget.EditText
import android.widget.TextView
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import com.mobileproxy.BuildConfig
import com.mobileproxy.R
import com.mobileproxy.core.network.NetworkManager
import com.mobileproxy.core.network.NetworkState
import com.mobileproxy.service.ProxyForegroundService
import com.mobileproxy.service.ProxyVpnService
import dagger.hilt.android.AndroidEntryPoint
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.collectLatest
import javax.inject.Inject

@AndroidEntryPoint
class MainActivity : AppCompatActivity() {

    @Inject lateinit var networkManager: NetworkManager

    private val scope = CoroutineScope(Dispatchers.Main + SupervisorJob())
    private var pendingStart = false

    private val vpnPermissionLauncher = registerForActivityResult(
        ActivityResultContracts.StartActivityForResult()
    ) { result ->
        if (result.resultCode == RESULT_OK && pendingStart) {
            startProxyService()
        } else {
            findViewById<TextView>(R.id.textStatus).text = "Status: VPN permission denied"
        }
        pendingStart = false
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        val serverUrlEdit = findViewById<EditText>(R.id.editServerUrl)
        if (serverUrlEdit.text.isNullOrEmpty() && BuildConfig.DEFAULT_SERVER_URL.isNotEmpty()) {
            serverUrlEdit.setText(BuildConfig.DEFAULT_SERVER_URL)
        }
        val statusText = findViewById<TextView>(R.id.textStatus)
        val cellularText = findViewById<TextView>(R.id.textCellular)
        val wifiText = findViewById<TextView>(R.id.textWifi)
        val vpnText = findViewById<TextView>(R.id.textVpn)
        val startButton = findViewById<Button>(R.id.buttonStart)
        val stopButton = findViewById<Button>(R.id.buttonStop)

        startButton.setOnClickListener {
            // Check VPN permission first
            val vpnIntent = VpnService.prepare(this)
            if (vpnIntent != null) {
                pendingStart = true
                vpnPermissionLauncher.launch(vpnIntent)
            } else {
                startProxyService()
            }
        }

        stopButton.setOnClickListener {
            val intent = Intent(this, ProxyForegroundService::class.java).apply {
                action = ProxyForegroundService.ACTION_STOP
            }
            startService(intent)
            statusText.text = "Status: Stopped"
        }

        // Observe network states
        scope.launch {
            networkManager.cellularState.collectLatest { state ->
                cellularText.text = when (state) {
                    is NetworkState.Connected -> "Cellular: Connected"
                    is NetworkState.Disconnected -> "Cellular: Disconnected"
                }
            }
        }
        scope.launch {
            networkManager.wifiState.collectLatest { state ->
                wifiText.text = when (state) {
                    is NetworkState.Connected -> "WiFi: Connected"
                    is NetworkState.Disconnected -> "WiFi: Disconnected"
                }
            }
        }

        // Observe VPN state
        scope.launch {
            ProxyVpnService.vpnState.collectLatest { connected ->
                vpnText.text = if (connected) "VPN: Connected" else "VPN: Disconnected"
            }
        }
    }

    private fun startProxyService() {
        val serverUrlEdit = findViewById<EditText>(R.id.editServerUrl)
        val statusText = findViewById<TextView>(R.id.textStatus)

        val serverUrl = serverUrlEdit.text.toString()
        val intent = Intent(this, ProxyForegroundService::class.java).apply {
            action = ProxyForegroundService.ACTION_START
            putExtra(ProxyForegroundService.EXTRA_SERVER_URL, serverUrl)
            putExtra(ProxyForegroundService.EXTRA_DEVICE_ID, BuildConfig.DEVICE_ID)
            putExtra(ProxyForegroundService.EXTRA_AUTH_TOKEN, BuildConfig.DEVICE_AUTH_TOKEN)
        }
        startForegroundService(intent)
        statusText.text = "Status: Running"
    }

    override fun onDestroy() {
        scope.cancel()
        super.onDestroy()
    }
}
