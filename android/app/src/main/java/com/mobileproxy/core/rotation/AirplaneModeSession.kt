package com.mobileproxy.core.rotation

import android.content.Intent
import android.os.Build
import android.os.Bundle
import android.service.voice.VoiceInteractionSession
import android.service.voice.VoiceInteractionSessionService
import android.util.Log

/**
 * Custom VoiceInteractionSession that handles airplane mode toggle commands.
 *
 * When a SwitchAirplaneMode command is received via the session Bundle,
 * it calls startVoiceActivity() with Samsung's VOICE_CONTROL_AIRPLANE_MODE intent.
 * Samsung's Settings app trusts this because it comes from the active voice assistant session.
 */
class AirplaneModeSession(context: VoiceInteractionSessionService) : VoiceInteractionSession(context) {

    companion object {
        private const val TAG = "AirplaneModeSession"
        private const val SAMSUNG_AIRPLANE_ACTION = "android.settings.VOICE_CONTROL_AIRPLANE_MODE"
        private const val EXTRA_AIRPLANE_ENABLED = "airplane_mode_enabled"
    }

    private var lastCommand: AirplaneModeCommand? = null

    override fun onCreate() {
        super.onCreate()
        Log.d(TAG, "Session created")
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            setUiEnabled(false)
        }
    }

    override fun onPrepareShow(args: Bundle?, showFlags: Int) {
        super.onPrepareShow(args, showFlags)
        val command = extractCommand(args)
        Log.d(TAG, "onPrepareShow: command=$command")
        if (command != null) {
            executeCommand(command)
        }
    }

    override fun onShow(args: Bundle?, showFlags: Int) {
        super.onShow(args, showFlags)
        val command = extractCommand(args)
        Log.d(TAG, "onShow: command=$command")
        if (command != null) {
            executeCommand(command)
        }
    }

    override fun onDestroy() {
        Log.d(TAG, "Session destroyed")
        super.onDestroy()
    }

    private fun extractCommand(args: Bundle?): AirplaneModeCommand? {
        if (args == null) return null
        return try {
            args.classLoader = AirplaneModeCommand::class.java.classLoader
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                args.getParcelable(AirplaneModeCommand.KEY, AirplaneModeCommand::class.java)
            } else {
                @Suppress("DEPRECATION")
                args.getParcelable(AirplaneModeCommand.KEY)
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to extract command from bundle", e)
            null
        }
    }

    private fun executeCommand(command: AirplaneModeCommand) {
        // Prevent duplicate execution (onPrepareShow + onShow may both fire)
        if (command == lastCommand) {
            Log.d(TAG, "Skipping duplicate command: $command")
            return
        }
        lastCommand = command

        try {
            Log.i(TAG, "Executing airplane mode toggle: enable=${command.enable}")
            val intent = Intent(SAMSUNG_AIRPLANE_ACTION).apply {
                putExtra(EXTRA_AIRPLANE_ENABLED, command.enable)
            }
            startVoiceActivity(intent)
            Log.i(TAG, "startVoiceActivity() called successfully for enable=${command.enable}")
        } catch (e: Exception) {
            Log.e(TAG, "Failed to execute airplane mode command", e)
        }
    }
}
