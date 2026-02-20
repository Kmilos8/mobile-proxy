package com.mobileproxy.service

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.util.Log
import com.mobileproxy.core.config.CredentialManager

class BootReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action == Intent.ACTION_BOOT_COMPLETED) {
            Log.i("BootReceiver", "Boot completed, checking for saved credentials")

            val credentialManager = CredentialManager(context)
            if (credentialManager.isPaired()) {
                Log.i("BootReceiver", "Device is paired, starting proxy service")
                val serviceIntent = Intent(context, ProxyForegroundService::class.java).apply {
                    action = ProxyForegroundService.ACTION_START
                    putExtra(ProxyForegroundService.EXTRA_SERVER_URL, credentialManager.getServerUrl())
                    putExtra(ProxyForegroundService.EXTRA_DEVICE_ID, credentialManager.getDeviceId())
                    putExtra(ProxyForegroundService.EXTRA_AUTH_TOKEN, credentialManager.getAuthToken())
                }
                context.startForegroundService(serviceIntent)
            } else {
                Log.i("BootReceiver", "Device not paired, skipping auto-start")
            }
        }
    }
}
