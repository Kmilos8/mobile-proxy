package com.mobileproxy.core.network

import android.content.Context
import android.net.ConnectivityManager
import android.net.Network
import android.net.NetworkCapabilities
import android.net.NetworkRequest
import android.util.Log
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.suspendCancellableCoroutine
import java.net.InetAddress
import java.net.Socket
import javax.inject.Inject
import javax.inject.Singleton
import javax.net.SocketFactory
import kotlin.coroutines.resume
import kotlin.coroutines.resumeWithException

@Singleton
class NetworkManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    companion object {
        private const val TAG = "NetworkManager"
    }

    private val connectivityManager = context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager

    private var cellularNetwork: Network? = null
    private var wifiNetwork: Network? = null

    private val _cellularState = MutableStateFlow<NetworkState>(NetworkState.Disconnected)
    val cellularState: StateFlow<NetworkState> = _cellularState

    private val _wifiState = MutableStateFlow<NetworkState>(NetworkState.Disconnected)
    val wifiState: StateFlow<NetworkState> = _wifiState

    private var cellularCallback: ConnectivityManager.NetworkCallback? = null
    private var wifiCallback: ConnectivityManager.NetworkCallback? = null

    /**
     * Request both cellular and WiFi networks simultaneously.
     * This is the core of the WiFi Split mechanism.
     */
    fun acquireNetworks() {
        acquireCellular()
        acquireWifi()
    }

    fun releaseNetworks() {
        cellularCallback?.let { connectivityManager.unregisterNetworkCallback(it) }
        wifiCallback?.let { connectivityManager.unregisterNetworkCallback(it) }
        cellularCallback = null
        wifiCallback = null
        cellularNetwork = null
        wifiNetwork = null
        _cellularState.value = NetworkState.Disconnected
        _wifiState.value = NetworkState.Disconnected
    }

    private fun acquireCellular() {
        val request = NetworkRequest.Builder()
            .addTransportType(NetworkCapabilities.TRANSPORT_CELLULAR)
            .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
            .build()

        cellularCallback = object : ConnectivityManager.NetworkCallback() {
            override fun onAvailable(network: Network) {
                Log.i(TAG, "Cellular network available: $network")
                cellularNetwork = network
                _cellularState.value = NetworkState.Connected(network)
            }

            override fun onLost(network: Network) {
                Log.w(TAG, "Cellular network lost: $network")
                cellularNetwork = null
                _cellularState.value = NetworkState.Disconnected
            }

            override fun onCapabilitiesChanged(network: Network, caps: NetworkCapabilities) {
                cellularNetwork = network
                _cellularState.value = NetworkState.Connected(network)
            }
        }
        connectivityManager.requestNetwork(request, cellularCallback!!)
    }

    private fun acquireWifi() {
        val request = NetworkRequest.Builder()
            .addTransportType(NetworkCapabilities.TRANSPORT_WIFI)
            .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
            .build()

        wifiCallback = object : ConnectivityManager.NetworkCallback() {
            override fun onAvailable(network: Network) {
                Log.i(TAG, "WiFi network available: $network")
                wifiNetwork = network
                _wifiState.value = NetworkState.Connected(network)
            }

            override fun onLost(network: Network) {
                Log.w(TAG, "WiFi network lost: $network")
                wifiNetwork = null
                _wifiState.value = NetworkState.Disconnected
            }

            override fun onCapabilitiesChanged(network: Network, caps: NetworkCapabilities) {
                wifiNetwork = network
                _wifiState.value = NetworkState.Connected(network)
            }
        }
        connectivityManager.requestNetwork(request, wifiCallback!!)
    }

    /**
     * Get a SocketFactory bound to the cellular network.
     * All sockets created by this factory will route through cellular (mobile IP).
     */
    fun getCellularSocketFactory(): SocketFactory? {
        return cellularNetwork?.socketFactory
    }

    /**
     * Get the cellular Network object for binding sockets directly.
     */
    fun getCellularNetwork(): Network? = cellularNetwork

    /**
     * Get the WiFi Network object (used for VPN tunnel).
     */
    fun getWifiNetwork(): Network? = wifiNetwork

    /**
     * Bind a socket to the cellular network explicitly.
     */
    fun bindSocketToCellular(socket: Socket): Boolean {
        val network = cellularNetwork ?: return false
        return try {
            network.bindSocket(socket)
            true
        } catch (e: Exception) {
            Log.e(TAG, "Failed to bind socket to cellular", e)
            false
        }
    }

    /**
     * Resolve DNS through cellular network to prevent DNS leaks.
     */
    suspend fun resolveDnsCellular(hostname: String): InetAddress {
        val network = cellularNetwork
            ?: throw IllegalStateException("Cellular network not available")
        return suspendCancellableCoroutine { cont ->
            try {
                val addresses = network.getAllByName(hostname)
                if (addresses.isNotEmpty()) {
                    cont.resume(addresses[0])
                } else {
                    cont.resumeWithException(Exception("DNS resolution failed for $hostname"))
                }
            } catch (e: Exception) {
                cont.resumeWithException(e)
            }
        }
    }

    /**
     * Disconnect and reconnect cellular to trigger IP change.
     * This is the fallback IP rotation method.
     */
    fun reconnectCellular() {
        Log.i(TAG, "Reconnecting cellular for IP rotation")
        cellularCallback?.let {
            connectivityManager.unregisterNetworkCallback(it)
        }
        cellularNetwork = null
        _cellularState.value = NetworkState.Disconnected

        // Re-acquire after a short delay (handled by caller)
        acquireCellular()
    }
}

sealed class NetworkState {
    data object Disconnected : NetworkState()
    data class Connected(val network: Network) : NetworkState()
}
