package com.mobileproxy.core.rotation

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.os.Build
import android.provider.Settings
import android.util.Log
import com.mobileproxy.core.network.NetworkManager
import com.mobileproxy.core.status.DeviceStatusReporter
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.Lazy
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.delay
import kotlinx.coroutines.withTimeoutOrNull
import java.util.concurrent.atomic.AtomicInteger
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class IPRotationManager @Inject constructor(
    @ApplicationContext private val context: Context,
    private val networkManager: NetworkManager,
    private val statusReporter: Lazy<DeviceStatusReporter>
) {
    companion object {
        private const val TAG = "IPRotationManager"
        private const val AIRPLANE_MODE_TOGGLE_TIMEOUT_MS = 7000L
        private const val AIRPLANE_MODE_OFF_DELAY_MS = 6000L
        private const val CELLULAR_RECONNECT_DELAY_MS = 5000L
    }

    private val commandIdCounter = AtomicInteger(0)

    /**
     * Cellular reconnect - no airplane mode, just reconnects cellular data.
     */
    suspend fun rotateByCellularReconnect(): Boolean {
        Log.i(TAG, "Starting IP rotation via cellular reconnect")
        return try {
            networkManager.reconnectCellular()
            delay(CELLULAR_RECONNECT_DELAY_MS)
            statusReporter.get().invalidateIpCache()
            true
        } catch (e: Exception) {
            Log.e(TAG, "Cellular reconnect rotation failed", e)
            false
        }
    }

    /**
     * Main rotation entry point — uses Voice Assistant session to toggle airplane mode
     * via Samsung's VOICE_CONTROL_AIRPLANE_MODE intent. No WRITE_SECURE_SETTINGS needed.
     *
     * Falls back to Settings.Global (requires ADB grant) and then Accessibility Service.
     */
    suspend fun requestAirplaneModeToggle() {
        Log.i(TAG, "Requesting IP rotation")

        // Method 1: Voice Assistant (Samsung VOICE_CONTROL_AIRPLANE_MODE)
        if (DigitalAssistantService.isReady) {
            val success = rotateViaVoiceAssistant()
            if (success) {
                statusReporter.get().invalidateIpCache()
                return
            }
            Log.w(TAG, "Voice Assistant method failed, trying fallbacks")
        } else {
            Log.w(TAG, "Digital Assistant service not ready, trying fallbacks")
        }

        // Method 2: Settings.Global (needs WRITE_SECURE_SETTINGS via ADB)
        val settingsSuccess = rotateViaSettingsGlobal()
        if (settingsSuccess) {
            return
        }

        // Method 3: Accessibility Service (UI automation fallback)
        Log.i(TAG, "Settings.Global failed, falling back to Accessibility")
        rotateViaAccessibility()
    }

    /**
     * Toggle airplane mode ON then OFF via Samsung's Voice Assistant intent.
     * This works without WRITE_SECURE_SETTINGS because Samsung trusts the active
     * voice assistant's VoiceInteractionSession.startVoiceActivity() calls.
     */
    private suspend fun rotateViaVoiceAssistant(): Boolean {
        Log.i(TAG, "Using Voice Assistant method (Samsung VOICE_CONTROL_AIRPLANE_MODE)")

        // Step 1: Turn airplane mode ON
        val onSuccess = toggleAirplaneModeViaVoice(enable = true)
        if (!onSuccess) {
            Log.e(TAG, "Voice Assistant: failed to enable airplane mode")
            return false
        }
        Log.i(TAG, "Voice Assistant: airplane mode ON confirmed")

        // Wait for radios to fully disengage
        delay(AIRPLANE_MODE_OFF_DELAY_MS)

        // Step 2: Turn airplane mode OFF
        val offSuccess = toggleAirplaneModeViaVoice(enable = false)
        if (!offSuccess) {
            Log.e(TAG, "Voice Assistant: failed to disable airplane mode")
            // Try to ensure airplane mode gets turned off
            delay(2000)
            toggleAirplaneModeViaVoice(enable = false)
            return false
        }
        Log.i(TAG, "Voice Assistant: airplane mode OFF confirmed")

        // Wait for cellular to reconnect
        delay(CELLULAR_RECONNECT_DELAY_MS)
        Log.i(TAG, "Voice Assistant: IP rotation complete")
        return true
    }

    /**
     * Send a command to the DigitalAssistantService and wait for the airplane mode
     * broadcast to confirm the toggle happened.
     */
    private suspend fun toggleAirplaneModeViaVoice(enable: Boolean): Boolean {
        val commandId = commandIdCounter.getAndIncrement()
        val command = AirplaneModeCommand(commandId, enable)

        // Register a receiver to detect when airplane mode actually changes
        val toggleConfirmed = CompletableDeferred<Boolean>()
        val receiver = object : BroadcastReceiver() {
            override fun onReceive(ctx: Context?, intent: Intent?) {
                if (intent?.action == Intent.ACTION_AIRPLANE_MODE_CHANGED) {
                    val state = intent.getBooleanExtra("state", false)
                    Log.d(TAG, "Airplane mode broadcast received: state=$state, expected=$enable")
                    if (state == enable) {
                        toggleConfirmed.complete(true)
                    }
                }
            }
        }

        val filter = IntentFilter(Intent.ACTION_AIRPLANE_MODE_CHANGED)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            context.registerReceiver(receiver, filter, Context.RECEIVER_EXPORTED)
        } else {
            context.registerReceiver(receiver, filter)
        }

        try {
            // Send the command to DigitalAssistantService
            val intent = Intent(DigitalAssistantService.ACTION_VOICE_COMMAND).apply {
                putExtra(AirplaneModeCommand.KEY, command)
                setPackage(context.packageName)
            }
            context.sendBroadcast(intent)
            Log.d(TAG, "Voice command sent: $command")

            // Wait for confirmation
            val result = withTimeoutOrNull(AIRPLANE_MODE_TOGGLE_TIMEOUT_MS) {
                toggleConfirmed.await()
            }

            if (result == null) {
                Log.w(TAG, "Airplane mode toggle timed out after ${AIRPLANE_MODE_TOGGLE_TIMEOUT_MS}ms (enable=$enable)")
                // Check if it actually changed despite no broadcast
                val currentState = Settings.Global.getInt(
                    context.contentResolver,
                    Settings.Global.AIRPLANE_MODE_ON, 0
                ) != 0
                if (currentState == enable) {
                    Log.i(TAG, "Airplane mode state matches expected (enable=$enable) despite timeout")
                    return true
                }
                return false
            }

            return true
        } finally {
            try {
                context.unregisterReceiver(receiver)
            } catch (_: Exception) {}
        }
    }

    /**
     * Accessibility Service — opens Quick Settings and taps airplane mode tile.
     */
    private suspend fun rotateViaAccessibility() {
        Log.i(TAG, "Using Accessibility Service method")

        if (!AirplaneModeAccessibilityService.isAvailable()) {
            Log.e(TAG, "AccessibilityService not available — enable it in Settings > Accessibility")
            return
        }

        val success = AirplaneModeAccessibilityService.requestToggleAndWait()
        Log.i(TAG, "Accessibility toggle result: $success")
        statusReporter.get().invalidateIpCache()
    }

    /**
     * Settings.Global write (legacy fallback).
     * Requires WRITE_SECURE_SETTINGS (granted via ADB pm grant).
     */
    private suspend fun rotateViaSettingsGlobal(): Boolean {
        Log.i(TAG, "Using Settings.Global method (requires WRITE_SECURE_SETTINGS)")

        try {
            Settings.Global.putInt(context.contentResolver, Settings.Global.AIRPLANE_MODE_ON, 1)
            Log.i(TAG, "Settings write: airplane_mode_on=1")
        } catch (e: SecurityException) {
            Log.e(TAG, "Settings.Global write failed (no WRITE_SECURE_SETTINGS)", e)
            return false
        }

        delay(AIRPLANE_MODE_OFF_DELAY_MS)

        try {
            Settings.Global.putInt(context.contentResolver, Settings.Global.AIRPLANE_MODE_ON, 0)
            Log.i(TAG, "Settings write: airplane_mode_on=0")
        } catch (e: Exception) {
            Log.e(TAG, "Settings.Global OFF write failed", e)
        }

        delay(CELLULAR_RECONNECT_DELAY_MS)
        statusReporter.get().invalidateIpCache()
        return true
    }
}
