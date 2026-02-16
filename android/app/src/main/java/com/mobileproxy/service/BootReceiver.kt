package com.mobileproxy.service

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.util.Log

class BootReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action == Intent.ACTION_BOOT_COMPLETED) {
            Log.i("BootReceiver", "Boot completed, starting proxy service")
            // TODO: Read saved config from DataStore and start service
            // For now, the user needs to manually start from the UI
        }
    }
}
