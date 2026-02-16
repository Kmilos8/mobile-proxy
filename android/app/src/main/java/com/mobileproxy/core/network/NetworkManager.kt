package com.mobileproxy.core.network

import android.content.Context
import android.net.ConnectivityManager
import android.net.Network
import android.net.NetworkCapabilities
import android.net.NetworkRequest
import android.os.Handler
import android.os.Looper
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

    @Volatile private var cellularNetwork: Network? = null
    @Volatile private var wifiNetwork: Network? = null

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
                Log.i(TAG, "Cellular network available via callback: $network")
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
        // Use main looper handler to ensure callbacks fire reliably
        connectivityManager.requestNetwork(request, cellularCallback!!, Handler(Looper.getMainLooper()))
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
        connectivityManager.requestNetwork(request, wifiCallback!!, Handler(Looper.getMainLooper()))
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
     * Find cellular network by scanning all available networks.
     * Fallback when requestNetwork callback hasn't fired.
     */
    private fun findCellularNetwork(): Network? {
        val networks = connectivityManager.allNetworks
        for (network in networks) {
            val caps = connectivityManager.getNetworkCapabilities(network) ?: continue
            if (caps.hasTransport(NetworkCapabilities.TRANSPORT_CELLULAR) &&
                caps.hasCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)) {
                Log.i(TAG, "Found cellular network via scan: $network")
                cellularNetwork = network
                _cellularState.value = NetworkState.Connected(network)
                return network
            }
        }
        return null
    }

    /**
     * Get the cellular network, using scan fallback if callback hasn't fired.
     */
    private fun getOrFindCellular(): Network {
        cellularNetwork?.let { return it }
        findCellularNetwork()?.let { return it }
        throw IllegalStateException("Cellular network not available - is Mobile Data enabled?")
    }

    /**
     * Create a socket connected through the cellular network using SocketFactory.
     * Uses scan fallback if requestNetwork callback hasn't fired yet.
     */
    fun createCellularSocket(address: InetAddress, port: Int): Socket {
        val network = getOrFindCellular()
        Log.d(TAG, "Creating cellular socket to ${address.hostAddress}:$port via network $network")
        return network.socketFactory.createSocket(address, port)
    }

    /**
     * Bind a socket to the cellular network explicitly.
     */
    fun bindSocketToCellular(socket: Socket) {
        val network = getOrFindCellular()
        network.bindSocket(socket)
        Log.d(TAG, "Socket bound to cellular network $network")
    }

    /**
     * Resolve DNS through cellular network to prevent DNS leaks.
     */
    suspend fun resolveDnsCellular(hostname: String): InetAddress {
        val network = getOrFindCellular()
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
