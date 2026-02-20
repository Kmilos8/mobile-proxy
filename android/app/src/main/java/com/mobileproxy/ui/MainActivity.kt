package com.mobileproxy.ui

import android.content.Intent
import android.net.VpnService
import android.os.Bundle
import android.widget.*
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import com.mobileproxy.BuildConfig
import com.mobileproxy.R
import com.mobileproxy.core.config.CredentialManager
import com.mobileproxy.core.network.NetworkManager
import com.mobileproxy.core.network.NetworkState
import com.mobileproxy.core.rotation.IPRotationManager
import com.mobileproxy.service.ProxyForegroundService
import com.mobileproxy.service.ProxyVpnService
import dagger.hilt.android.AndroidEntryPoint
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.collectLatest
import javax.inject.Inject

@AndroidEntryPoint
class MainActivity : AppCompatActivity() {

    @Inject lateinit var networkManager: NetworkManager
    @Inject lateinit var rotationManager: IPRotationManager
    @Inject lateinit var credentialManager: CredentialManager

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

    private val setupLauncher = registerForActivityResult(
        ActivityResultContracts.StartActivityForResult()
    ) { result ->
        if (result.resultCode == RESULT_OK) {
            // Setup complete — auto-start proxy
            startProxyService()
        }
    }

    private val pairingLauncher = registerForActivityResult(
        ActivityResultContracts.StartActivityForResult()
    ) { result ->
        if (result.resultCode == RESULT_OK) {
            // Pairing complete
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // Migrate from BuildConfig if paired via old method
        migrateFromBuildConfig()

        // If not paired, launch PairingActivity
        if (!credentialManager.isPaired()) {
            pairingLauncher.launch(Intent(this, PairingActivity::class.java))
        }

        setContentView(R.layout.activity_main)

        val statusText = findViewById<TextView>(R.id.textStatus)
        val cellularText = findViewById<TextView>(R.id.textCellular)
        val wifiText = findViewById<TextView>(R.id.textWifi)
        val vpnText = findViewById<TextView>(R.id.textVpn)
        val startButton = findViewById<Button>(R.id.buttonStart)
        val stopButton = findViewById<Button>(R.id.buttonStop)
        val setupButton = findViewById<Button>(R.id.buttonSetupWizard)

        startButton.setOnClickListener {
            // Must be paired first
            if (!credentialManager.isPaired()) {
                pairingLauncher.launch(Intent(this, PairingActivity::class.java))
                return@setOnClickListener
            }

            // Redirect to setup wizard if not completed
            if (!SetupActivity.isSetupComplete(this)) {
                setupLauncher.launch(Intent(this, SetupActivity::class.java))
                return@setOnClickListener
            }

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

        setupButton.setOnClickListener {
            setupLauncher.launch(Intent(this, SetupActivity::class.java))
        }

        findViewById<Button>(R.id.buttonUnpair).setOnClickListener {
            credentialManager.clear()
            statusText.text = "Status: Unpaired"
            pairingLauncher.launch(Intent(this, PairingActivity::class.java))
        }

        findViewById<Button>(R.id.buttonChangeIp).setOnClickListener {
            statusText.text = "Status: Changing IP..."
            scope.launch {
                try {
                    rotationManager.requestAirplaneModeToggle()
                    statusText.text = "Status: IP changed"
                } catch (e: Exception) {
                    statusText.text = "Status: IP change failed - ${e.message}"
                }
            }
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
        scope.launch {
            ProxyVpnService.vpnState.collectLatest { connected ->
                vpnText.text = if (connected) "VPN: Connected" else "VPN: Disconnected"
            }
        }
    }

    private fun migrateFromBuildConfig() {
        // Only migrate once — skip if already done or user has unpaired
        if (credentialManager.isMigrationDone()) return

        if (!credentialManager.isPaired() &&
            BuildConfig.DEVICE_ID.isNotEmpty() &&
            BuildConfig.DEFAULT_SERVER_URL.isNotEmpty()
        ) {
            credentialManager.saveCredentials(
                serverUrl = BuildConfig.DEFAULT_SERVER_URL,
                deviceId = BuildConfig.DEVICE_ID,
                authToken = BuildConfig.DEVICE_AUTH_TOKEN,
                vpnConfig = "",
                basePort = 0
            )
        }
        credentialManager.setMigrationDone()
    }

    private fun startProxyService() {
        val statusText = findViewById<TextView>(R.id.textStatus)

        // Use CredentialManager values
        val serverUrl = credentialManager.getServerUrl().ifEmpty { BuildConfig.DEFAULT_SERVER_URL }
        val deviceId = credentialManager.getDeviceId().ifEmpty { BuildConfig.DEVICE_ID }
        val authToken = credentialManager.getAuthToken().ifEmpty { BuildConfig.DEVICE_AUTH_TOKEN }

        val intent = Intent(this, ProxyForegroundService::class.java).apply {
            action = ProxyForegroundService.ACTION_START
            putExtra(ProxyForegroundService.EXTRA_SERVER_URL, serverUrl)
            putExtra(ProxyForegroundService.EXTRA_DEVICE_ID, deviceId)
            putExtra(ProxyForegroundService.EXTRA_AUTH_TOKEN, authToken)
        }
        startForegroundService(intent)
        statusText.text = "Status: Running"
    }

    override fun onDestroy() {
        scope.cancel()
        super.onDestroy()
    }
}
