package com.mobileproxy.core.rotation

import android.accessibilityservice.AccessibilityService
import android.accessibilityservice.GestureDescription
import android.content.Intent
import android.graphics.Path
import android.os.Handler
import android.os.Looper
import android.provider.Settings
import android.util.Log
import android.view.accessibility.AccessibilityEvent
import android.view.accessibility.AccessibilityNodeInfo

class AirplaneModeAccessibilityService : AccessibilityService() {

    companion object {
        private const val TAG = "AirplaneModeAS"
        private var instance: AirplaneModeAccessibilityService? = null
        private var toggleRequested = false

        fun requestToggle() {
            toggleRequested = true
            instance?.performToggle()
        }

        fun isAvailable(): Boolean = instance != null
    }

    private val handler = Handler(Looper.getMainLooper())

    override fun onServiceConnected() {
        super.onServiceConnected()
        instance = this
        Log.i(TAG, "Accessibility service connected")
    }

    override fun onAccessibilityEvent(event: AccessibilityEvent?) {
        if (!toggleRequested) return
        // Look for airplane mode tile in quick settings
        event?.source?.let { searchForAirplaneToggle(it) }
    }

    override fun onInterrupt() {
        Log.w(TAG, "Accessibility service interrupted")
    }

    override fun onDestroy() {
        instance = null
        super.onDestroy()
    }

    private fun performToggle() {
        // Open Quick Settings panel
        performGlobalAction(GLOBAL_ACTION_QUICK_SETTINGS)

        // Wait for Quick Settings to appear, then look for airplane toggle
        handler.postDelayed({
            // After toggling on, schedule toggle off after delay
            handler.postDelayed({
                performGlobalAction(GLOBAL_ACTION_QUICK_SETTINGS)
                handler.postDelayed({
                    toggleRequested = false
                }, 3000)
            }, 3000)
        }, 1500)
    }

    private fun searchForAirplaneToggle(node: AccessibilityNodeInfo) {
        // Search for airplane mode tile by description or text
        val airplaneNodes = node.findAccessibilityNodeInfosByText("Airplane")
            .plus(node.findAccessibilityNodeInfosByText("Flight"))
            .plus(node.findAccessibilityNodeInfosByText("Aeroplane"))

        for (airplaneNode in airplaneNodes) {
            if (airplaneNode.isClickable) {
                airplaneNode.performAction(AccessibilityNodeInfo.ACTION_CLICK)
                Log.i(TAG, "Clicked airplane mode toggle")
                return
            }
            // Try parent
            airplaneNode.parent?.let {
                if (it.isClickable) {
                    it.performAction(AccessibilityNodeInfo.ACTION_CLICK)
                    Log.i(TAG, "Clicked airplane mode toggle parent")
                    return
                }
            }
        }
    }
}
