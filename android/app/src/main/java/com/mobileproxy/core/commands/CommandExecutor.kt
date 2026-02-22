package com.mobileproxy.core.commands

import android.content.Context
import android.media.AudioAttributes
import android.media.AudioManager
import android.media.MediaPlayer
import android.media.RingtoneManager
import android.os.Build
import android.os.VibrationEffect
import android.os.Vibrator
import android.os.VibratorManager
import android.util.Log
import com.mobileproxy.core.rotation.IPRotationManager
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.delay
import javax.inject.Inject
import javax.inject.Singleton

data class DeviceCommand(
    val id: String,
    val type: String,
    val payload: String
)

@Singleton
class CommandExecutor @Inject constructor(
    @ApplicationContext private val context: Context,
    private val rotationManager: IPRotationManager
) {
    companion object {
        private const val TAG = "CommandExecutor"
    }

    suspend fun execute(command: DeviceCommand): Result<String> {
        Log.i(TAG, "Executing command: ${command.type}")

        return when (command.type) {
            "rotate_ip" -> {
                val success = rotationManager.rotateByCellularReconnect()
                if (success) Result.success("IP rotation initiated")
                else Result.failure(Exception("IP rotation failed"))
            }
            "rotate_ip_airplane" -> {
                rotationManager.requestAirplaneModeToggle()
                Result.success("Airplane mode toggle requested")
            }
            "find_phone" -> {
                playFindPhoneAlert()
                Result.success("Find phone alert playing")
            }
            else -> {
                Result.failure(Exception("Unknown command: ${command.type}"))
            }
        }
    }

    private suspend fun playFindPhoneAlert() {
        // Vibrate (best-effort, may fail without permission)
        try {
            val vibrator = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
                val vm = context.getSystemService(Context.VIBRATOR_MANAGER_SERVICE) as VibratorManager
                vm.defaultVibrator
            } else {
                @Suppress("DEPRECATION")
                context.getSystemService(Context.VIBRATOR_SERVICE) as Vibrator
            }
            val pattern = longArrayOf(0, 500, 200, 500, 200, 500, 200, 500, 200, 500)
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                vibrator.vibrate(VibrationEffect.createWaveform(pattern, -1))
            } else {
                @Suppress("DEPRECATION")
                vibrator.vibrate(pattern, -1)
            }
        } catch (e: Exception) {
            Log.w(TAG, "Vibrate failed: ${e.message}")
        }

        // Play alarm sound at max volume for 10 seconds
        var mediaPlayer: MediaPlayer? = null
        var audioManager: AudioManager? = null
        var originalVolume = 0
        try {
            val alarmUri = RingtoneManager.getDefaultUri(RingtoneManager.TYPE_ALARM)
                ?: RingtoneManager.getDefaultUri(RingtoneManager.TYPE_RINGTONE)

            audioManager = context.getSystemService(Context.AUDIO_SERVICE) as AudioManager
            originalVolume = audioManager.getStreamVolume(AudioManager.STREAM_ALARM)
            val maxVolume = audioManager.getStreamMaxVolume(AudioManager.STREAM_ALARM)
            audioManager.setStreamVolume(AudioManager.STREAM_ALARM, maxVolume, 0)

            mediaPlayer = MediaPlayer().apply {
                setAudioAttributes(
                    AudioAttributes.Builder()
                        .setUsage(AudioAttributes.USAGE_ALARM)
                        .setContentType(AudioAttributes.CONTENT_TYPE_SONIFICATION)
                        .build()
                )
                setDataSource(context, alarmUri)
                prepare()
                start()
            }

            delay(10_000)
            Log.i(TAG, "Find phone alert completed")
        } catch (e: Exception) {
            Log.e(TAG, "Find phone alert error", e)
        } finally {
            try { mediaPlayer?.stop() } catch (_: Exception) {}
            try { mediaPlayer?.release() } catch (_: Exception) {}
            try { audioManager?.setStreamVolume(AudioManager.STREAM_ALARM, originalVolume, 0) } catch (_: Exception) {}
        }
    }
}
