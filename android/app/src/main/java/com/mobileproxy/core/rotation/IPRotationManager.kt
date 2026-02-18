package com.mobileproxy.core.rotation

import android.content.Context
import android.util.Log
import com.mobileproxy.core.network.NetworkManager
import com.mobileproxy.core.status.DeviceStatusReporter
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.Lazy
import kotlinx.coroutines.delay
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
    }

    /**
     * Primary rotation method: Cellular reconnect.
     */
    suspend fun rotateByCellularReconnect(): Boolean {
        Log.i(TAG, "Starting IP rotation via cellular reconnect")
        return try {
            networkManager.reconnectCellular()
            delay(5000)
            true
        } catch (e: Exception) {
            Log.e(TAG, "Cellular reconnect rotation failed", e)
            false
        }
    }

    /**
     * Toggle airplane mode ON then OFF via Accessibility Service.
     * The service opens Quick Settings, taps the airplane mode tile,
     * waits, then taps it again — identical to manual toggle.
     */
    suspend fun requestAirplaneModeToggle() {
        Log.i(TAG, "Requesting airplane mode toggle via AccessibilityService")

        if (!AirplaneModeAccessibilityService.isAvailable()) {
            Log.e(TAG, "AccessibilityService not available — enable it in Settings > Accessibility")
            return
        }

        val success = AirplaneModeAccessibilityService.requestToggleAndWait()
        Log.i(TAG, "Airplane mode toggle result: $success")

        // Invalidate cached IP so next heartbeat fetches the new one
        statusReporter.get().invalidateIpCache()
    }
}
