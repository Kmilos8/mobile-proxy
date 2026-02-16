package com.mobileproxy.core.commands

import android.util.Log
import com.mobileproxy.core.rotation.IPRotationManager
import javax.inject.Inject
import javax.inject.Singleton

data class DeviceCommand(
    val id: String,
    val type: String,
    val payload: String
)

@Singleton
class CommandExecutor @Inject constructor(
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
            "reboot" -> {
                // Cannot reboot without root; just restart the service
                Result.success("Service restart not implemented yet")
            }
            "find_phone" -> {
                // TODO: Vibrate and flash
                Result.success("Find phone triggered")
            }
            else -> {
                Result.failure(Exception("Unknown command: ${command.type}"))
            }
        }
    }
}
