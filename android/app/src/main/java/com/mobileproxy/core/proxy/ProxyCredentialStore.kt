package com.mobileproxy.core.proxy

import java.util.concurrent.ConcurrentHashMap
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Thread-safe in-memory store of valid proxy credentials.
 * Updated from heartbeat responses.
 */
@Singleton
class ProxyCredentialStore @Inject constructor() {

    // username -> password (plaintext)
    private val credentials = ConcurrentHashMap<String, String>()

    /**
     * Replace all stored credentials with the given list.
     */
    fun update(creds: List<Pair<String, String>>) {
        credentials.clear()
        for ((username, password) in creds) {
            credentials[username] = password
        }
    }

    /**
     * Validate a username/password pair.
     * Returns true if the credentials match.
     */
    fun validate(username: String, password: String): Boolean {
        val stored = credentials[username] ?: return false
        return stored == password
    }

    /**
     * Check if any credentials are configured.
     * If no credentials are configured, all connections are allowed (backward compat).
     */
    fun hasCredentials(): Boolean = credentials.isNotEmpty()
}
