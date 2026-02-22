package com.mobileproxy.core.commands

import android.content.Context
import android.hardware.camera2.CameraManager
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
        var cameraId: String? = null
        var cameraManager: CameraManager? = null

        // Turn flashlight on
        try {
            cameraManager = context.getSystemService(Context.CAMERA_SERVICE) as CameraManager
            cameraId = cameraManager.cameraIdList.firstOrNull { id ->
                cameraManager.getCameraCharacteristics(id)
                    .get(android.hardware.camera2.CameraCharacteristics.FLASH_INFO_AVAILABLE) == true
            }
            if (cameraId != null) {
                cameraManager.setTorchMode(cameraId, true)
                Log.i(TAG, "Flashlight ON")
            } else {
                Log.w(TAG, "No camera with flash found")
            }
        } catch (e: Exception) {
            Log.w(TAG, "Flashlight failed: ${e.message}")
        }

        // Vibrate for ~10 seconds
        try {
            val vibrator = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
                val vm = context.getSystemService(Context.VIBRATOR_MANAGER_SERVICE) as VibratorManager
                vm.defaultVibrator
            } else {
                @Suppress("DEPRECATION")
                context.getSystemService(Context.VIBRATOR_SERVICE) as Vibrator
            }
            val pattern = longArrayOf(
                0, 500, 200, 500, 200, 500, 200, 500, 200, 500,
                200, 500, 200, 500, 200, 500, 200, 500, 200, 500,
                200, 500, 200, 500, 200, 500, 200, 500, 200, 500,
                200, 500, 200, 500, 200, 500, 200, 500, 200, 500
            )
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                vibrator.vibrate(VibrationEffect.createWaveform(pattern, -1))
            } else {
                @Suppress("DEPRECATION")
                vibrator.vibrate(pattern, -1)
            }
        } catch (e: Exception) {
            Log.w(TAG, "Vibrate failed: ${e.message}")
        }

        // Wait for vibration to finish then turn flashlight off
        delay(20_000)
        try {
            if (cameraId != null && cameraManager != null) {
                cameraManager.setTorchMode(cameraId, false)
                Log.i(TAG, "Flashlight OFF")
            }
        } catch (e: Exception) {
            Log.w(TAG, "Flashlight off failed: ${e.message}")
        }
        Log.i(TAG, "Find phone alert completed")
    }
}
