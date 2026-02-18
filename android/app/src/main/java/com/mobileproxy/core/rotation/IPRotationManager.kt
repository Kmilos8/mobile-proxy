package com.mobileproxy.core.rotation

import android.content.Context
import android.content.Intent
import android.net.ConnectivityManager
import android.net.NetworkCapabilities
import android.provider.Settings
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

    private fun isCellularConnected(): Boolean {
        val cm = context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
        val activeNetwork = cm.activeNetwork ?: return false
        val caps = cm.getNetworkCapabilities(activeNetwork) ?: return false
        return caps.hasTransport(NetworkCapabilities.TRANSPORT_CELLULAR)
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
     * Enable airplane mode, wait for cellular to actually go down, then disable.
     * Tries `cmd connectivity` first (full system-level toggle), falls back to Settings.Global.
     */
    private suspend fun toggleAirplaneMode(): Boolean {
        return try {
            val hadCellular = hasCellularNetwork()
            Log.i(TAG, "Cellular before toggle: $hadCellular")

            // Try cmd connectivity first (fully engages airplane mode on Samsung)
            val useCmd = tryEnableAirplaneModeCmd()
            if (!useCmd) {
                // Fallback to Settings.Global
                Settings.Global.putInt(context.contentResolver, Settings.Global.AIRPLANE_MODE_ON, 1)
                context.sendBroadcast(Intent(Intent.ACTION_AIRPLANE_MODE_CHANGED).putExtra("state", true))
            }
            Log.i(TAG, "Airplane mode ON (method=${if (useCmd) "cmd" else "settings"})")

            // Wait for cellular to actually go down
            if (hadCellular) {
                val startTime = System.currentTimeMillis()
                while (hasCellularNetwork() && (System.currentTimeMillis() - startTime) < CELLULAR_DOWN_TIMEOUT_MS) {
                    delay(POLL_INTERVAL_MS)
                }
                val elapsed = System.currentTimeMillis() - startTime
                val cellularDown = !hasCellularNetwork()
                Log.i(TAG, "Cellular down: $cellularDown (waited ${elapsed}ms)")

                // Extra wait after cellular drops to ensure carrier releases the IP
                if (cellularDown) {
                    delay(POST_DOWN_WAIT_MS)
                }
            } else {
                // No cellular was detected, just wait a fixed time
                delay(7000)
            }

            // Disable airplane mode
            if (useCmd) {
                tryDisableAirplaneModeCmd()
            } else {
                Settings.Global.putInt(context.contentResolver, Settings.Global.AIRPLANE_MODE_ON, 0)
                context.sendBroadcast(Intent(Intent.ACTION_AIRPLANE_MODE_CHANGED).putExtra("state", false))
            }
            Log.i(TAG, "Airplane mode OFF")

            // Invalidate cached IP so next heartbeat fetches the new one
            statusReporter.get().invalidateIpCache()

            true
        } catch (e: Exception) {
            Log.e(TAG, "Airplane mode toggle failed", e)
            // Make sure airplane mode is off
            try {
                tryDisableAirplaneModeCmd()
                Settings.Global.putInt(context.contentResolver, Settings.Global.AIRPLANE_MODE_ON, 0)
            } catch (_: Exception) {}
            false
        }
    }

    private fun tryEnableAirplaneModeCmd(): Boolean {
        return try {
            val process = Runtime.getRuntime().exec(arrayOf("sh", "-c", "cmd connectivity airplane-mode enable"))
            val exitCode = process.waitFor()
            exitCode == 0
        } catch (e: Exception) {
            Log.w(TAG, "cmd connectivity enable failed", e)
            false
        }
    }

    private fun tryDisableAirplaneModeCmd(): Boolean {
        return try {
            val process = Runtime.getRuntime().exec(arrayOf("sh", "-c", "cmd connectivity airplane-mode disable"))
            val exitCode = process.waitFor()
            exitCode == 0
        } catch (e: Exception) {
            Log.w(TAG, "cmd connectivity disable failed", e)
            false
        }
    }
}
