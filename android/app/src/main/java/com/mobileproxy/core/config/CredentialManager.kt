package com.mobileproxy.core.config

import android.content.Context
import android.content.SharedPreferences
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class CredentialManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    companion object {
        private const val PREFS_NAME = "mobileproxy_credentials"
        private const val KEY_SERVER_URL = "server_url"
        private const val KEY_DEVICE_ID = "device_id"
        private const val KEY_AUTH_TOKEN = "auth_token"
        private const val KEY_VPN_CONFIG = "vpn_config"
        private const val KEY_BASE_PORT = "base_port"
    }

    private val prefs: SharedPreferences =
        context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)

    fun isPaired(): Boolean {
        return getDeviceId().isNotEmpty() && getAuthToken().isNotEmpty() && getServerUrl().isNotEmpty()
    }

    fun saveCredentials(
        serverUrl: String,
        deviceId: String,
        authToken: String,
        vpnConfig: String,
        basePort: Int
    ) {
        prefs.edit()
            .putString(KEY_SERVER_URL, serverUrl)
            .putString(KEY_DEVICE_ID, deviceId)
            .putString(KEY_AUTH_TOKEN, authToken)
            .putString(KEY_VPN_CONFIG, vpnConfig)
            .putInt(KEY_BASE_PORT, basePort)
            .apply()
    }

    fun getServerUrl(): String = prefs.getString(KEY_SERVER_URL, "") ?: ""
    fun getDeviceId(): String = prefs.getString(KEY_DEVICE_ID, "") ?: ""
    fun getAuthToken(): String = prefs.getString(KEY_AUTH_TOKEN, "") ?: ""
    fun getVpnConfig(): String = prefs.getString(KEY_VPN_CONFIG, "") ?: ""
    fun getBasePort(): Int = prefs.getInt(KEY_BASE_PORT, 0)

    fun clear() {
        prefs.edit().clear().apply()
    }
}
