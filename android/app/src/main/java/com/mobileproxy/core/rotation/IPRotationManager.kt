package com.mobileproxy.core.rotation

import android.content.Context
import android.content.Intent
import android.net.ConnectivityManager
import android.provider.Settings
import android.util.Log
import com.mobileproxy.core.network.NetworkManager
import com.mobileproxy.core.status.DeviceStatusReporter
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.Lazy
import kotlinx.coroutines.delay
import javax.inject.Inject
import javax.inject.Singleton

private const val AIRPLANE_MODE_DELAY = 3000L

@Singleton
class IPRotationManager @Inject constructor(
    @ApplicationContext private val context: Context,
    private val networkManager: NetworkManager,
    private val statusReporter: Lazy<DeviceStatusReporter>
) {
    companion object {
        private const val TAG = "IPRotationManager"
        private const val PREFS_NAME = "rotation_prefs"
        private const val KEY_METHOD = "rotation_method"

        const val METHOD_ACCESSIBILITY = 0
        const val METHOD_SETTINGS_GLOBAL = 1
        const val METHOD_CONNECTIVITY_MANAGER = 2
        const val METHOD_CELLULAR_RECONNECT = 3
    }

    fun getSavedMethod(): Int {
        val prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
        return prefs.getInt(KEY_METHOD, METHOD_ACCESSIBILITY)
    }

    fun saveMethod(method: Int) {
        context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
            .edit().putInt(KEY_METHOD, method).apply()
        Log.i(TAG, "Rotation method saved: $method")
    }

    /**
     * Cellular reconnect - no airplane mode, just reconnects cellular data.
     */
    suspend fun rotateByCellularReconnect(): Boolean {
        Log.i(TAG, "Starting IP rotation via cellular reconnect")
        return try {
            networkManager.reconnectCellular()
            delay(5000)
            statusReporter.get().invalidateIpCache()
            true
        } catch (e: Exception) {
            Log.e(TAG, "Cellular reconnect rotation failed", e)
            false
        }
    }

    /**
     * Main rotation entry point — uses the method selected by the user.
     */
    suspend fun requestAirplaneModeToggle() {
        val method = getSavedMethod()
        Log.i(TAG, "Requesting IP rotation, method=$method")

        when (method) {
            METHOD_ACCESSIBILITY -> rotateViaAccessibility()
            METHOD_SETTINGS_GLOBAL -> rotateViaSettingsGlobal()
            METHOD_CONNECTIVITY_MANAGER -> rotateViaConnectivityManager()
            METHOD_CELLULAR_RECONNECT -> rotateByCellularReconnect()
            else -> rotateViaAccessibility()
        }
    }

    /**
     * Method 0: Accessibility Service — opens Quick Settings and taps airplane mode tile.
     * Most reliable on Samsung and modern Android devices.
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
     * Method 1: Settings.Global write + broadcast.
     * Requires WRITE_SECURE_SETTINGS — granted when app is set as Digital Assistant.
     * The broadcast tells Android to actually engage/disengage radios.
     */
    private suspend fun rotateViaSettingsGlobal() {
        Log.i(TAG, "Using Settings.Global + broadcast method (Digital Assistant)")
        try {
            // Turn ON
            Settings.Global.putInt(context.contentResolver, Settings.Global.AIRPLANE_MODE_ON, 1)
            sendAirplaneModeBroadcast(true)
            Log.i(TAG, "Airplane mode ON (setting + broadcast)")

            delay(AIRPLANE_MODE_DELAY)
        } catch (e: SecurityException) {
            Log.e(TAG, "Settings.Global write failed (not Digital Assistant?) — falling back to Accessibility", e)
            rotateViaAccessibility()
            return
        } catch (e: Exception) {
            Log.e(TAG, "Settings.Global ON failed", e)
            return
        }

        // Always turn OFF — even if broadcast failed, the setting was written
        try {
            Settings.Global.putInt(context.contentResolver, Settings.Global.AIRPLANE_MODE_ON, 0)
            sendAirplaneModeBroadcast(false)
            Log.i(TAG, "Airplane mode OFF (setting + broadcast)")

            delay(5000) // wait for cellular to reconnect
            statusReporter.get().invalidateIpCache()
        } catch (e: Exception) {
            Log.e(TAG, "Settings.Global OFF failed", e)
        }
    }

    private fun sendAirplaneModeBroadcast(enabled: Boolean) {
        try {
            val intent = Intent(Intent.ACTION_AIRPLANE_MODE_CHANGED).putExtra("state", enabled)
            context.sendBroadcast(intent)
            Log.i(TAG, "Broadcast sent: airplane_mode=$enabled")
        } catch (e: SecurityException) {
            Log.w(TAG, "Protected broadcast not allowed (expected on some devices)", e)
        }
    }

    /**
     * Method 2: ConnectivityManager.setAirplaneMode() via reflection.
     * Hidden system API used by Quick Settings. Requires NETWORK_AIRPLANE_MODE
     * or similar permission — may not work on all devices.
     */
    private suspend fun rotateViaConnectivityManager() {
        Log.i(TAG, "Using ConnectivityManager reflection method")
        try {
            val cm = context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
            val method = cm.javaClass.getDeclaredMethod("setAirplaneMode", Boolean::class.javaPrimitiveType)

            method.invoke(cm, true)
            Log.i(TAG, "setAirplaneMode(true) success")

            delay(7000)

            method.invoke(cm, false)
            Log.i(TAG, "setAirplaneMode(false) success")

            statusReporter.get().invalidateIpCache()
        } catch (e: Exception) {
            Log.e(TAG, "ConnectivityManager method failed: ${e.cause?.message ?: e.message}")
        }
    }
}
