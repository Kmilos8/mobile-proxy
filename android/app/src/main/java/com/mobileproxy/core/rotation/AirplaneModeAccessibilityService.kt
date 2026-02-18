package com.mobileproxy.core.rotation

import android.accessibilityservice.AccessibilityService
import android.util.Log
import android.view.accessibility.AccessibilityEvent
import android.view.accessibility.AccessibilityNodeInfo
import kotlinx.coroutines.*

class AirplaneModeAccessibilityService : AccessibilityService() {

    companion object {
        private const val TAG = "AirplaneModeAS"
        private var instance: AirplaneModeAccessibilityService? = null

        @Volatile
        var pendingToggle: CompletableDeferred<Boolean>? = null

        fun requestToggle() {
            val service = instance
            if (service == null) {
                Log.e(TAG, "Accessibility service not running")
                return
            }
            val deferred = CompletableDeferred<Boolean>()
            pendingToggle = deferred
            service.startToggleSequence()
        }

        suspend fun requestToggleAndWait(): Boolean {
            val service = instance
            if (service == null) {
                Log.e(TAG, "Accessibility service not running")
                return false
            }
            val deferred = CompletableDeferred<Boolean>()
            pendingToggle = deferred
            service.startToggleSequence()
            return try {
                withTimeout(30000) { deferred.await() }
            } catch (e: Exception) {
                Log.e(TAG, "Toggle timed out", e)
                false
            }
        }

        fun isAvailable(): Boolean = instance != null
    }

    private val scope = CoroutineScope(Dispatchers.Main + SupervisorJob())

    override fun onServiceConnected() {
        super.onServiceConnected()
        instance = this
        Log.i(TAG, "Accessibility service connected")
    }

    override fun onAccessibilityEvent(event: AccessibilityEvent?) {
        // Not used — we actively search the node tree instead
    }

    override fun onInterrupt() {
        Log.w(TAG, "Accessibility service interrupted")
    }

    override fun onDestroy() {
        instance = null
        scope.cancel()
        super.onDestroy()
    }

    private fun startToggleSequence() {
        scope.launch {
            try {
                // Step 1: Open Quick Settings
                Log.i(TAG, "Opening Quick Settings")
                performGlobalAction(GLOBAL_ACTION_QUICK_SETTINGS)
                delay(1500)

                // Step 2: Find and click airplane mode tile
                val clicked = findAndClickAirplane()
                if (!clicked) {
                    Log.e(TAG, "Could not find airplane mode tile")
                    closeQuickSettings()
                    pendingToggle?.complete(false)
                    return@launch
                }
                Log.i(TAG, "Airplane mode toggled ON")

                // Step 3: Wait for the airplane mode to engage (7 seconds like manual toggle)
                delay(7000)

                // Step 4: Click airplane mode again to turn it off
                // Quick settings might still be open, or we may need to reopen
                var clickedOff = findAndClickAirplane()
                if (!clickedOff) {
                    // Try reopening quick settings
                    performGlobalAction(GLOBAL_ACTION_QUICK_SETTINGS)
                    delay(1500)
                    clickedOff = findAndClickAirplane()
                }

                if (clickedOff) {
                    Log.i(TAG, "Airplane mode toggled OFF")
                } else {
                    Log.w(TAG, "Could not click airplane OFF, trying to close anyway")
                }

                // Step 5: Close quick settings
                delay(500)
                closeQuickSettings()

                pendingToggle?.complete(clickedOff)
            } catch (e: Exception) {
                Log.e(TAG, "Toggle sequence failed", e)
                closeQuickSettings()
                pendingToggle?.complete(false)
            }
        }
    }

    private fun closeQuickSettings() {
        performGlobalAction(GLOBAL_ACTION_BACK)
        performGlobalAction(GLOBAL_ACTION_BACK)
    }

    private fun findAndClickAirplane(): Boolean {
        val root = rootInActiveWindow ?: return false
        return searchAndClick(root, 0)
    }

    private fun searchAndClick(node: AccessibilityNodeInfo, depth: Int): Boolean {
        if (depth > 15) return false

        val desc = node.contentDescription?.toString()?.lowercase() ?: ""
        val text = node.text?.toString()?.lowercase() ?: ""

        // Match airplane/flight mode tile
        if (desc.contains("airplane") || desc.contains("flight") || desc.contains("aeroplane") ||
            text.contains("airplane") || text.contains("flight") || text.contains("aeroplane")) {

            // Try clicking this node
            if (node.isClickable) {
                node.performAction(AccessibilityNodeInfo.ACTION_CLICK)
                Log.i(TAG, "Clicked: '$desc' / '$text'")
                return true
            }
            // Try parent
            node.parent?.let { parent ->
                if (parent.isClickable) {
                    parent.performAction(AccessibilityNodeInfo.ACTION_CLICK)
                    Log.i(TAG, "Clicked parent of: '$desc' / '$text'")
                    return true
                }
            }
        }

        // Recurse into children
        for (i in 0 until node.childCount) {
            val child = node.getChild(i) ?: continue
            if (searchAndClick(child, depth + 1)) return true
        }

        return false
    }
}
