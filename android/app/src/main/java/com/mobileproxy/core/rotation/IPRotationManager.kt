package com.mobileproxy.core.rotation

import android.content.Context
import android.net.ConnectivityManager
import android.net.NetworkCapabilities
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
        private const val CELLULAR_DOWN_TIMEOUT_MS = 15000L
        private const val POLL_INTERVAL_MS = 500L
        private const val POST_DOWN_WAIT_MS = 2000L
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
     * Toggle airplane mode ON then OFF to force a new IP assignment.
     */
    suspend fun requestAirplaneModeToggle() {
        Log.i(TAG, "Requesting airplane mode toggle")
        if (toggleAirplaneMode()) {
            Log.i(TAG, "Airplane mode rotation completed")
        } else {
            Log.w(TAG, "Airplane mode toggle failed, falling back to AccessibilityService")
            AirplaneModeAccessibilityService.requestToggle()
        }
    }

    private fun hasCellularNetwork(): Boolean {
        val cm = context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
        for (network in cm.allNetworks) {
            val caps = cm.getNetworkCapabilities(network) ?: continue
            if (caps.hasTransport(NetworkCapabilities.TRANSPORT_CELLULAR)) return true
        }
        return false
    }

    /**
     * Call the hidden ConnectivityManager.setAirplaneMode(boolean) via reflection.
     * This is the same API the Settings app / Quick Settings tile uses.
     * It actually engages/disengages the radios (solid icon, cellular drops).
     * Settings.Global.putInt only changes the DB value without affecting radios.
     */
    private fun setAirplaneMode(enabled: Boolean): Boolean {
        return try {
            val cm = context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
            val method = cm.javaClass.getDeclaredMethod("setAirplaneMode", Boolean::class.javaPrimitiveType)
            method.invoke(cm, enabled)
            Log.i(TAG, "ConnectivityManager.setAirplaneMode($enabled) success")
            true
        } catch (e: Exception) {
            Log.w(TAG, "ConnectivityManager.setAirplaneMode($enabled) failed: ${e.cause?.message ?: e.message}")
            false
        }
    }

    /**
     * Enable airplane mode, wait for cellular to actually go down, then disable.
     * Uses ConnectivityManager.setAirplaneMode() (hidden API via reflection) which
     * is the same code path as the manual quick settings toggle.
     */
    private suspend fun toggleAirplaneMode(): Boolean {
        return try {
            val hadCellular = hasCellularNetwork()
            Log.i(TAG, "Cellular before toggle: $hadCellular")

            // Enable airplane mode via the real system API
            val enabled = setAirplaneMode(true)
            if (!enabled) {
                Log.e(TAG, "setAirplaneMode(true) failed")
                return false
            }

            // Poll for cellular to actually go down
            if (hadCellular) {
                val startTime = System.currentTimeMillis()
                var lastLog = 0L
                while (hasCellularNetwork() && (System.currentTimeMillis() - startTime) < CELLULAR_DOWN_TIMEOUT_MS) {
                    val elapsed = System.currentTimeMillis() - startTime
                    if (elapsed - lastLog >= 2000) {
                        Log.i(TAG, "Waiting for cellular down... (${elapsed}ms)")
                        lastLog = elapsed
                    }
                    delay(POLL_INTERVAL_MS)
                }
                val elapsed = System.currentTimeMillis() - startTime
                val cellularDown = !hasCellularNetwork()
                Log.i(TAG, "Cellular down: $cellularDown (after ${elapsed}ms)")

                if (cellularDown) {
                    delay(POST_DOWN_WAIT_MS)
                } else {
                    Log.w(TAG, "Cellular did NOT go down after ${elapsed}ms, proceeding anyway")
                    delay(POST_DOWN_WAIT_MS)
                }
            } else {
                delay(7000)
            }

            // Disable airplane mode
            setAirplaneMode(false)
            Log.i(TAG, "Airplane mode OFF")

            // Invalidate cached IP so next heartbeat fetches the new one
            statusReporter.get().invalidateIpCache()

            true
        } catch (e: Exception) {
            Log.e(TAG, "Airplane mode toggle failed", e)
            // Safety: ensure airplane mode is off
            try { setAirplaneMode(false) } catch (_: Exception) {}
            false
        }
    }
}
