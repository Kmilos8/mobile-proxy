package com.mobileproxy.core.vpn

import android.os.ParcelFileDescriptor
import android.util.Log
import com.mobileproxy.service.ProxyVpnService
import kotlinx.coroutines.*
import java.io.FileInputStream
import java.io.FileOutputStream
import java.net.DatagramPacket
import java.net.DatagramSocket
import java.net.InetSocketAddress
import java.net.Socket
import java.util.concurrent.atomic.AtomicLong

class VpnTunnelManager(
    private val vpnService: ProxyVpnService,
    private val serverAddress: String,
    private val serverPort: Int,
    private val deviceId: String
) {
    companion object {
        private const val TAG = "VpnTunnelManager"
        const val MTU = 1400
        const val KEEPALIVE_MS = 25_000L
        private const val RECV_BUF_SIZE = 2 * 1024 * 1024 // 2MB socket receive buffer
        private const val SEND_BUF_SIZE = 2 * 1024 * 1024 // 2MB socket send buffer
        private const val RECONNECT_DELAY_MS = 3_000L
        private const val MAX_RECONNECT_DELAY_MS = 30_000L
        private const val PONG_TIMEOUT_MS = 45_000L // if no PONG in 45s, reconnect

        // Packet type prefixes
        private const val TYPE_AUTH: Byte = 0x01
        private const val TYPE_DATA: Byte = 0x02
        private const val TYPE_PING: Byte = 0x03
        private const val TYPE_AUTH_OK: Byte = 0x01
        private const val TYPE_AUTH_FAIL: Byte = 0x03
        private const val TYPE_PONG: Byte = 0x04
    }

    private var tunFd: ParcelFileDescriptor? = null
    private var udpSocket: DatagramSocket? = null
    private var scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    // Track last PONG for dead-tunnel detection
    private val lastPongTime = AtomicLong(0L)

    @Volatile
    var vpnIP: String = ""
        private set

    @Volatile
    var isConnected = false
        private set

    // Whether the manager should keep trying to stay connected
    @Volatile
    var shouldRun = false
        private set

    suspend fun connect(): Boolean = withContext(Dispatchers.IO) {
        shouldRun = true
        // Ensure we have a fresh scope (previous disconnect may have cancelled it)
        if (!scope.isActive) {
            scope = CoroutineScope(Dispatchers.IO + SupervisorJob())
        }

        // Retry initial connect with exponential backoff
        var delayMs = RECONNECT_DELAY_MS
        while (shouldRun) {
            if (connectInternal()) return@withContext true
            Log.w(TAG, "Initial connect failed, retrying in ${delayMs}ms...")
            delay(delayMs)
            delayMs = (delayMs * 2).coerceAtMost(MAX_RECONNECT_DELAY_MS)
        }
        false
    }

    private suspend fun connectInternal(): Boolean {
        try {
            // Create UDP socket — do NOT protect() yet.
            // protect() sets SO_MARK which can break UDP receive on Samsung
            // when no VPN is active. We protect after auth, before TUN establish.
            val socket = DatagramSocket()
            socket.receiveBufferSize = RECV_BUF_SIZE
            socket.sendBufferSize = SEND_BUF_SIZE
            socket.connect(InetSocketAddress(serverAddress, serverPort))
            udpSocket = socket
            Log.i(TAG, "UDP socket created, connecting to $serverAddress:$serverPort" +
                " (recv=${socket.receiveBufferSize/1024}KB, send=${socket.sendBufferSize/1024}KB)")

            // Authenticate (no VPN active yet, so normal routing works fine)
            val assignedIP = authenticate(socket) ?: return false

            vpnIP = assignedIP
            Log.i(TAG, "Authenticated, assigned VPN IP: $vpnIP")

            // NOW protect the socket before establishing TUN.
            // Once TUN is active, all traffic routes through VPN —
            // protect() ensures our tunnel socket bypasses VPN routing.
            vpnService.protect(socket)
            Log.i(TAG, "Socket protected from VPN routing")

            // Create TUN interface
            tunFd = createTun(vpnIP)
            if (tunFd == null) {
                Log.e(TAG, "Failed to create TUN interface")
                return false
            }

            isConnected = true
            lastPongTime.set(System.currentTimeMillis())

            // Start packet relay coroutines
            scope.launch { tunToUdp() }
            scope.launch { udpToTun() }
            scope.launch { keepalive() }
            scope.launch { watchdog() }

            Log.i(TAG, "VPN tunnel connected successfully")
            return true
        } catch (e: Exception) {
            Log.e(TAG, "Connection failed: ${e.message}", e)
            teardownConnection()
            return false
        }
    }

    fun disconnect() {
        Log.i(TAG, "Disconnecting VPN tunnel")
        shouldRun = false
        isConnected = false
        try { scope.cancel() } catch (_: Exception) {}
        teardownConnection()
        vpnIP = ""
    }

    private fun teardownConnection() {
        isConnected = false
        try { tunFd?.close() } catch (_: Exception) {}
        try { udpSocket?.close() } catch (_: Exception) {}
        tunFd = null
        udpSocket = null
    }

    private suspend fun reconnect() {
        if (!shouldRun) return
        Log.w(TAG, "Initiating reconnect...")
        teardownConnection()

        var delayMs = RECONNECT_DELAY_MS
        while (shouldRun) {
            Log.i(TAG, "Reconnecting in ${delayMs}ms...")
            delay(delayMs)
            if (!shouldRun) return

            if (connectInternal()) {
                Log.i(TAG, "Reconnected successfully")
                // Notify ProxyVpnService of reconnection
                ProxyVpnService.onReconnected()
                return
            }

            // Exponential backoff
            delayMs = (delayMs * 2).coerceAtMost(MAX_RECONNECT_DELAY_MS)
        }
    }

    fun protectSocket(socket: Socket): Boolean {
        return vpnService.protect(socket)
    }

    fun protectSocket(socket: DatagramSocket): Boolean {
        return vpnService.protect(socket)
    }

    private fun createTun(assignedIP: String): ParcelFileDescriptor? {
        return vpnService.Builder()
            .setSession("MobileProxy")
            .addAddress(assignedIP, 24)
            .addRoute("0.0.0.0", 0)
            .addDnsServer("8.8.8.8")
            .addDnsServer("8.8.4.4")
            .setMtu(MTU)
            .setBlocking(true)
            .establish()
    }

    private fun authenticate(socket: DatagramSocket): String? {
        // Parse device ID UUID to 16 bytes
        val uuidBytes = uuidToBytes(deviceId)
        if (uuidBytes == null) {
            Log.e(TAG, "Invalid device ID UUID: $deviceId")
            return null
        }

        // Send AUTH: [0x01][16-byte device_id]
        val authPacket = ByteArray(17)
        authPacket[0] = TYPE_AUTH
        System.arraycopy(uuidBytes, 0, authPacket, 1, 16)

        socket.send(DatagramPacket(authPacket, authPacket.size))
        Log.i(TAG, "AUTH sent, waiting for response...")

        // Receive response
        val recvBuf = ByteArray(64)
        val recvPacket = DatagramPacket(recvBuf, recvBuf.size)
        socket.soTimeout = 10000 // 10s timeout for auth
        socket.receive(recvPacket)

        if (recvPacket.length < 1) {
            Log.e(TAG, "Empty auth response")
            return null
        }

        return when (recvBuf[0]) {
            TYPE_AUTH_OK -> {
                if (recvPacket.length < 5) {
                    Log.e(TAG, "AUTH_OK too short")
                    return null
                }
                // Parse 4-byte IPv4
                val ip = "${recvBuf[1].toInt() and 0xFF}.${recvBuf[2].toInt() and 0xFF}.${recvBuf[3].toInt() and 0xFF}.${recvBuf[4].toInt() and 0xFF}"
                socket.soTimeout = 0 // Remove timeout for data
                ip
            }
            TYPE_AUTH_FAIL -> {
                Log.e(TAG, "Authentication failed")
                null
            }
            else -> {
                Log.e(TAG, "Unknown auth response type: ${recvBuf[0]}")
                null
            }
        }
    }

    private fun tunToUdp() {
        val fd = tunFd ?: return
        val socket = udpSocket ?: return
        val input = FileInputStream(fd.fileDescriptor)
        val buffer = ByteArray(MTU + 1) // +1 for type prefix

        Log.i(TAG, "tunToUdp started")
        try {
            while (isConnected) {
                val n = input.read(buffer, 1, MTU)
                if (n <= 0) continue

                buffer[0] = TYPE_DATA
                socket.send(DatagramPacket(buffer, 0, n + 1))
            }
        } catch (e: Exception) {
            if (isConnected) {
                Log.e(TAG, "tunToUdp error, triggering reconnect", e)
                scope.launch { reconnect() }
            }
        }
        Log.i(TAG, "tunToUdp stopped")
    }

    private fun udpToTun() {
        val fd = tunFd ?: return
        val socket = udpSocket ?: return
        val output = FileOutputStream(fd.fileDescriptor)
        val buffer = ByteArray(MTU + 100) // extra headroom for encapsulation

        Log.i(TAG, "udpToTun started")
        try {
            while (isConnected) {
                val packet = DatagramPacket(buffer, buffer.size)
                socket.receive(packet)

                if (packet.length < 1) continue

                when (buffer[0]) {
                    TYPE_PONG -> {
                        // PONG is 1 byte — handle before DATA size check
                        lastPongTime.set(System.currentTimeMillis())
                    }
                    TYPE_DATA -> {
                        if (packet.length < 2) continue
                        // Write raw IP packet to TUN (skip type byte)
                        output.write(buffer, 1, packet.length - 1)
                    }
                }
            }
        } catch (e: Exception) {
            if (isConnected) {
                Log.e(TAG, "udpToTun error, triggering reconnect", e)
                scope.launch { reconnect() }
            }
        }
        Log.i(TAG, "udpToTun stopped")
    }

    private suspend fun keepalive() {
        val socket = udpSocket ?: return
        val pingPacket = DatagramPacket(byteArrayOf(TYPE_PING), 1)

        Log.i(TAG, "Keepalive started")
        try {
            while (isConnected) {
                delay(KEEPALIVE_MS)
                if (!isConnected) break
                socket.send(pingPacket)
            }
        } catch (e: Exception) {
            if (isConnected) Log.e(TAG, "Keepalive error", e)
        }
        Log.i(TAG, "Keepalive stopped")
    }

    /**
     * Watchdog: detects dead tunnels by monitoring PONG responses.
     * If no PONG received within PONG_TIMEOUT_MS, triggers reconnect.
     */
    private suspend fun watchdog() {
        Log.i(TAG, "Watchdog started")
        try {
            while (isConnected && shouldRun) {
                delay(KEEPALIVE_MS)
                if (!isConnected) break

                val elapsed = System.currentTimeMillis() - lastPongTime.get()
                if (elapsed > PONG_TIMEOUT_MS) {
                    Log.w(TAG, "No PONG received in ${elapsed}ms, tunnel appears dead")
                    reconnect()
                    return
                }
            }
        } catch (e: Exception) {
            if (isConnected) Log.e(TAG, "Watchdog error", e)
        }
        Log.i(TAG, "Watchdog stopped")
    }

    private fun uuidToBytes(uuid: String): ByteArray? {
        return try {
            val hex = uuid.replace("-", "")
            if (hex.length != 32) return null
            val bytes = ByteArray(16)
            for (i in 0 until 16) {
                bytes[i] = hex.substring(i * 2, i * 2 + 2).toInt(16).toByte()
            }
            bytes
        } catch (e: Exception) {
            null
        }
    }
}
