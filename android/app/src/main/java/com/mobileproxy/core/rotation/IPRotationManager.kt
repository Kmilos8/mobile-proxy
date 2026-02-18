package com.mobileproxy.core.rotation

import android.content.Context
import android.provider.Settings
import android.util.Log
import com.mobileproxy.core.network.NetworkManager
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.delay
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class IPRotationManager @Inject constructor(
    @ApplicationContext private val context: Context,
    private val networkManager: NetworkManager
) {
    companion object {
        private const val TAG = "IPRotationManager"
        private const val AIRPLANE_MODE_DELAY_MS = 5000L
    }

    /**
     * Primary rotation method: Cellular reconnect.
     * Unregisters and re-requests the cellular network, causing modem re-registration.
     */
    suspend fun rotateByCellularReconnect(): Boolean {
        Log.i(TAG, "Starting IP rotation via cellular reconnect")
        return try {
            networkManager.reconnectCellular()
            // Wait for network to re-establish
            delay(5000)
            true
        } catch (e: Exception) {
            Log.e(TAG, "Cellular reconnect rotation failed", e)
            false
        }
    }

    /**
     * Toggle airplane mode ON then OFF to force a new IP assignment.
     * Uses Settings.Global write (requires WRITE_SECURE_SETTINGS granted via adb).
     * Android picks up the change via ContentObserver and toggles the radio.
     * Falls back to AirplaneModeAccessibilityService if permission is missing.
     */
    suspend fun requestAirplaneModeToggle() {
        Log.i(TAG, "Requesting airplane mode toggle")
        if (toggleAirplaneModeViaSettings()) {
            Log.i(TAG, "Airplane mode toggled via Settings.Global")
        } else {
            Log.w(TAG, "Settings.Global write failed, falling back to AccessibilityService")
            AirplaneModeAccessibilityService.requestToggle()
        }
    }

    /**
     * Directly write Settings.Global.AIRPLANE_MODE_ON.
     * Android picks up the change via ContentObserver — no broadcast needed.
     */
    private suspend fun toggleAirplaneModeViaSettings(): Boolean {
        return try {
            // Turn airplane mode ON
            Settings.Global.putInt(context.contentResolver, Settings.Global.AIRPLANE_MODE_ON, 1)
            Log.i(TAG, "Airplane mode ON")

            delay(AIRPLANE_MODE_DELAY_MS)

            // Turn airplane mode OFF
            Settings.Global.putInt(context.contentResolver, Settings.Global.AIRPLANE_MODE_ON, 0)
            Log.i(TAG, "Airplane mode OFF")

            true
        } catch (e: SecurityException) {
            Log.e(TAG, "No permission to write AIRPLANE_MODE_ON", e)
            false
        } catch (e: Exception) {
            Log.e(TAG, "Failed to toggle airplane mode via Settings.Global", e)
            false
        }
    }
}
