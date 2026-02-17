package com.mobileproxy.core.rotation

import android.service.voice.VoiceInteractionService
import android.util.Log

/**
 * Minimal VoiceInteractionService stub.
 * Registering as the device's digital assistant grants WRITE_SECURE_SETTINGS,
 * which allows direct airplane mode toggling via Settings.Global.
 */
class DigitalAssistantService : VoiceInteractionService() {

    companion object {
        private const val TAG = "DigitalAssistantSvc"
    }

    override fun onReady() {
        super.onReady()
        Log.i(TAG, "Digital assistant service ready")
    }
}
