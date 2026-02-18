package com.mobileproxy.ui

import android.Manifest
import android.app.role.RoleManager
import android.content.Context
import android.content.Intent
import android.net.Uri
import android.net.VpnService
import android.os.Build
import android.os.Bundle
import android.os.PowerManager
import android.provider.Settings
import android.widget.Button
import android.widget.ProgressBar
import android.widget.TextView
import android.widget.ViewFlipper
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import androidx.core.app.NotificationManagerCompat
import com.mobileproxy.R
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class SetupActivity : AppCompatActivity() {

    companion object {
        private const val PREFS_NAME = "setup_prefs"
        private const val KEY_SETUP_COMPLETE = "setup_complete"

        fun isSetupComplete(context: Context): Boolean {
            return context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
                .getBoolean(KEY_SETUP_COMPLETE, false)
        }

        fun resetSetup(context: Context) {
            context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
                .edit().putBoolean(KEY_SETUP_COMPLETE, false).apply()
        }
    }

    private lateinit var viewFlipper: ViewFlipper
    private lateinit var stepIndicator: TextView
    private lateinit var progressBar: ProgressBar

    private val notificationPermissionLauncher = registerForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { _ ->
        // onResume will re-check and advance
    }

    private val vpnPermissionLauncher = registerForActivityResult(
        ActivityResultContracts.StartActivityForResult()
    ) { _ ->
        // onResume will re-check and advance
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_setup)

        viewFlipper = findViewById(R.id.viewFlipper)
        stepIndicator = findViewById(R.id.textStepIndicator)
        progressBar = findViewById(R.id.progressSetup)

        // Step 1: Welcome
        findViewById<Button>(R.id.buttonStartSetup).setOnClickListener {
            goToStep(1)
        }

        // Step 2: Notifications
        findViewById<Button>(R.id.buttonAllowNotifications).setOnClickListener {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                notificationPermissionLauncher.launch(Manifest.permission.POST_NOTIFICATIONS)
            } else {
                // Pre-13: notifications are on by default, advance
                goToStep(2)
            }
        }
        findViewById<Button>(R.id.buttonSkipNotifications).setOnClickListener {
            goToStep(2)
        }

        // Step 3: VPN
        findViewById<Button>(R.id.buttonAllowVpn).setOnClickListener {
            val vpnIntent = VpnService.prepare(this)
            if (vpnIntent != null) {
                vpnPermissionLauncher.launch(vpnIntent)
            } else {
                // Already granted
                goToStep(3)
            }
        }

        // Step 4: Digital Assistant
        findViewById<Button>(R.id.buttonOpenAssistantSettings).setOnClickListener {
            openDigitalAssistantSettings()
        }
        findViewById<Button>(R.id.buttonContinueAssistant).setOnClickListener {
            goToStep(4)
        }

        // Step 5: Battery Optimization
        findViewById<Button>(R.id.buttonDisableBattery).setOnClickListener {
            val intent = Intent(Settings.ACTION_REQUEST_IGNORE_BATTERY_OPTIMIZATIONS).apply {
                data = Uri.parse("package:$packageName")
            }
            startActivity(intent)
        }
        findViewById<Button>(R.id.buttonContinueBattery).setOnClickListener {
            goToStep(5)
        }

        // Step 6: Done
        findViewById<Button>(R.id.buttonFinish).setOnClickListener {
            markSetupComplete()
            setResult(RESULT_OK)
            finish()
        }
    }

    override fun onResume() {
        super.onResume()
        updateCurrentStepStatus()
    }

    private fun goToStep(step: Int) {
        viewFlipper.displayedChild = step
        stepIndicator.text = "Step ${step + 1} of 6"
        progressBar.progress = step + 1
        updateCurrentStepStatus()
    }

    private fun updateCurrentStepStatus() {
        val currentStep = viewFlipper.displayedChild

        when (currentStep) {
            1 -> { // Notifications
                val granted = NotificationManagerCompat.from(this).areNotificationsEnabled()
                findViewById<TextView>(R.id.textNotificationStatus).text =
                    if (granted) "Status: Granted" else "Status: Not granted"
                if (granted) goToStep(2)
            }
            2 -> { // VPN
                val granted = VpnService.prepare(this) == null
                findViewById<TextView>(R.id.textVpnStatus).text =
                    if (granted) "Status: Granted" else "Status: Not granted"
                if (granted) goToStep(3)
            }
            3 -> { // Digital Assistant
                val isAssistant = isDigitalAssistant()
                findViewById<TextView>(R.id.textAssistantStatus).text =
                    if (isAssistant) "Status: MobileProxy is active" else "Status: Not set"
                if (isAssistant) goToStep(4)
            }
            4 -> { // Battery
                val pm = getSystemService(Context.POWER_SERVICE) as PowerManager
                val ignored = pm.isIgnoringBatteryOptimizations(packageName)
                findViewById<TextView>(R.id.textBatteryStatus).text =
                    if (ignored) "Status: Unrestricted" else "Status: Optimized"
                if (ignored) goToStep(5)
            }
        }
    }

    private fun isDigitalAssistant(): Boolean {
        // Check via RoleManager (used by Samsung and Android 10+)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            try {
                val roleManager = getSystemService(Context.ROLE_SERVICE) as RoleManager
                if (roleManager.isRoleHeld(RoleManager.ROLE_ASSISTANT)) {
                    return true
                }
            } catch (_: Exception) {}
        }
        // Fallback: check voice_interaction_service setting
        return try {
            val voiceService = Settings.Secure.getString(
                contentResolver, "voice_interaction_service"
            )
            voiceService?.contains(packageName) == true
        } catch (_: Exception) {
            false
        }
    }

    private fun openDigitalAssistantSettings() {
        // Samsung maps ACTION_VOICE_INPUT_SETTINGS to the Digital Assistant picker
        val intents = listOf(
            Intent(Settings.ACTION_VOICE_INPUT_SETTINGS),
            Intent(Settings.ACTION_MANAGE_DEFAULT_APPS_SETTINGS),
            Intent(Settings.ACTION_SETTINGS)
        )
        for (intent in intents) {
            if (intent.resolveActivity(packageManager) != null) {
                startActivity(intent)
                return
            }
        }
        // Last resort
        startActivity(Intent(Settings.ACTION_SETTINGS))
    }

    private fun markSetupComplete() {
        getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
            .edit().putBoolean(KEY_SETUP_COMPLETE, true).apply()
    }
}
