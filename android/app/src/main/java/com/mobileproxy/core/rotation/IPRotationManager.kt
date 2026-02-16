package com.mobileproxy.core.rotation

import android.content.Context
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
     * Trigger airplane mode via Accessibility Service.
     * The AccessibilityService handles the actual toggle.
     */
    fun requestAirplaneModeToggle() {
        Log.i(TAG, "Requesting airplane mode toggle via Accessibility Service")
        AirplaneModeAccessibilityService.requestToggle()
    }
}
