package com.mobileproxy.core.rotation

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.os.Build
import android.os.Bundle
import android.service.voice.VoiceInteractionService
import android.util.Log

/**
 * VoiceInteractionService that enables airplane mode toggling via Samsung's
 * VOICE_CONTROL_AIRPLANE_MODE intent.
 *
 * Flow:
 * 1. IPRotationManager sends a local broadcast with ACTION_VOICE_COMMAND
 * 2. This service receives it and calls showSession() with the command in the Bundle
 * 3. AirplaneModeSession.onPrepareShow()/onShow() calls startVoiceActivity()
 *    with Samsung's intent, which toggles airplane mode without WRITE_SECURE_SETTINGS
 */
class DigitalAssistantService : VoiceInteractionService() {

    companion object {
        private const val TAG = "DigitalAssistantSvc"
        const val ACTION_VOICE_COMMAND = "com.mobileproxy.intent.action.VOICE_COMMAND"

        @Volatile
        var isReady: Boolean = false
            private set
    }

    private var commandReceiver: BroadcastReceiver? = null

    override fun onReady() {
        super.onReady()
        Log.i(TAG, "Digital assistant service ready")
        isReady = true

        if (commandReceiver == null) {
            val receiver = object : BroadcastReceiver() {
                override fun onReceive(context: Context?, intent: Intent?) {
                    if (intent?.action != ACTION_VOICE_COMMAND) return
                    if (!intent.hasExtra(AirplaneModeCommand.KEY)) return

                    val command: AirplaneModeCommand? = try {
                        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                            intent.getParcelableExtra(AirplaneModeCommand.KEY, AirplaneModeCommand::class.java)
                        } else {
                            @Suppress("DEPRECATION")
                            intent.getParcelableExtra(AirplaneModeCommand.KEY)
                        }
                    } catch (e: Exception) {
                        Log.e(TAG, "Failed to extract command from intent", e)
                        null
                    }

                    if (command == null) {
                        Log.e(TAG, "Received VOICE_COMMAND without valid command")
                        return
                    }

                    Log.i(TAG, "Received voice command: $command, starting session")
                    try {
                        val bundle = Bundle().apply {
                            putParcelable(AirplaneModeCommand.KEY, command)
                        }
                        showSession(bundle, 0)
                    } catch (e: Exception) {
                        Log.e(TAG, "Failed to start voice interaction session", e)
                    }
                }
            }

            val filter = IntentFilter(ACTION_VOICE_COMMAND)
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                registerReceiver(receiver, filter, Context.RECEIVER_NOT_EXPORTED)
            } else {
                registerReceiver(receiver, filter)
            }
            commandReceiver = receiver
            Log.i(TAG, "Command receiver registered")
        }
    }

    override fun onShutdown() {
        Log.i(TAG, "Digital assistant service shutdown")
        unregisterCommandReceiver()
        isReady = false
        super.onShutdown()
    }

    override fun onDestroy() {
        Log.i(TAG, "Digital assistant service destroyed")
        unregisterCommandReceiver()
        isReady = false
        super.onDestroy()
    }

    private fun unregisterCommandReceiver() {
        try {
            commandReceiver?.let { unregisterReceiver(it) }
        } catch (_: Exception) {}
        commandReceiver = null
    }
}
