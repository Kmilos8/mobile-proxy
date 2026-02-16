package com.mobileproxy.service

import android.content.Intent
import android.net.VpnService
import android.util.Log

/**
 * Placeholder VPN service for OpenVPN integration.
 * In production, this will be replaced/extended by ics-openvpn's VpnService.
 * The key function is protect() - used to prevent VPN tunnel socket from being
 * routed through itself.
 */
class ProxyVpnService : VpnService() {
    companion object {
        private const val TAG = "ProxyVpnService"
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        Log.i(TAG, "VPN service started")
        return START_STICKY
    }

    override fun onDestroy() {
        Log.i(TAG, "VPN service destroyed")
        super.onDestroy()
    }
}
