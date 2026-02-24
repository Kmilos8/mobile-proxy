package com.mobileproxy.core.proxy

import android.util.Log
import com.mobileproxy.core.network.NetworkManager
import com.mobileproxy.service.ProxyVpnService
import kotlinx.coroutines.*
import java.io.DataInputStream
import java.io.DataOutputStream
import java.net.*
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.ConcurrentLinkedQueue
import java.util.concurrent.atomic.AtomicLong
import javax.inject.Inject
import javax.inject.Singleton

/**
 * SOCKS5 proxy server (RFC 1928) with username/password auth (RFC 1929).
 * Supports CONNECT and UDP ASSOCIATE commands.
 */
@Singleton
class Socks5ProxyServer @Inject constructor(
    private val networkManager: NetworkManager,
    private val credentialStore: ProxyCredentialStore
) {
    companion object {
        private const val TAG = "Socks5ProxyServer"
        private const val SOCKS_VERSION: Byte = 0x05
        private const val CMD_CONNECT: Byte = 0x01
        private const val CMD_UDP_ASSOCIATE: Byte = 0x03
        private const val ATYP_IPV4: Byte = 0x01
        private const val ATYP_DOMAIN: Byte = 0x03
        private const val ATYP_IPV6: Byte = 0x04
        private const val AUTH_NONE: Byte = 0x00
        private const val AUTH_USERPASS: Byte = 0x02
        private const val AUTH_NO_ACCEPTABLE: Byte = 0xFF.toByte()
        private const val USERPASS_VERSION: Byte = 0x01
        private const val USERPASS_SUCCESS: Byte = 0x00
        private const val USERPASS_FAILURE: Byte = 0x01
        private const val BUFFER_SIZE = 32768
        private const val UDP_BUFFER_SIZE = 65535
    }

    private var serverSocket: ServerSocket? = null
    private var udpRelaySocket: DatagramSocket? = null
    private var running = false
    private var scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    private val _bytesIn = AtomicLong(0)
    private val _bytesOut = AtomicLong(0)
    val bytesIn: Long get() = _bytesIn.get()
    val bytesOut: Long get() = _bytesOut.get()

    // UDP ASSOCIATE session state
    data class UdpSession(
        val clientAddr: InetSocketAddress,
        val targetSocket: DatagramSocket
    )
    // key = "clientIP:clientPort"
    private val udpSessions = ConcurrentHashMap<String, UdpSession>()

    data class PendingUdpAssociation(
        val targetSocket: DatagramSocket,
        val clientAddrDeferred: CompletableDeferred<InetSocketAddress>
    )
    private val pendingAssociations = ConcurrentLinkedQueue<PendingUdpAssociation>()

    fun start(port: Int = 1080) {
        if (running) return
        running = true
        scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

        scope.launch {
            try {
                serverSocket = ServerSocket(port)
                Log.i(TAG, "SOCKS5 proxy listening on port $port (TCP)")

                udpRelaySocket = DatagramSocket(null).apply {
                    reuseAddress = true
                    bind(InetSocketAddress(port))
                }
                Log.i(TAG, "SOCKS5 UDP relay listening on port $port (UDP)")

                launch { udpRelayLoop() }

                while (running) {
                    val client = serverSocket?.accept() ?: break
                    launch { handleClient(client) }
                }
            } catch (e: Exception) {
                if (running) Log.e(TAG, "Server error", e)
            }
        }
    }

    fun stop() {
        running = false
        serverSocket?.close()
        udpRelaySocket?.close()
        udpSessions.values.forEach { session ->
            session.targetSocket.close()
        }
        udpSessions.clear()
        pendingAssociations.clear()
        scope.cancel()
    }

    private suspend fun handleClient(clientSocket: Socket) {
        try {
            val input = DataInputStream(clientSocket.getInputStream())
            val output = DataOutputStream(clientSocket.getOutputStream())

            // Greeting
            val version = input.readByte()
            if (version != SOCKS_VERSION) return

            val nMethods = input.readByte().toInt() and 0xFF
            val methods = ByteArray(nMethods)
            input.readFully(methods)

            if (credentialStore.hasCredentials()) {
                // Require username/password auth (RFC 1929)
                val hasUserPass = methods.any { it == AUTH_USERPASS }
                if (!hasUserPass) {
                    output.write(byteArrayOf(SOCKS_VERSION, AUTH_NO_ACCEPTABLE))
                    output.flush()
                    return
                }

                // Select username/password auth
                output.write(byteArrayOf(SOCKS_VERSION, AUTH_USERPASS))
                output.flush()

                // RFC 1929 subnegotiation
                val authVersion = input.readByte()
                if (authVersion != USERPASS_VERSION) {
                    output.write(byteArrayOf(USERPASS_VERSION, USERPASS_FAILURE))
                    output.flush()
                    return
                }

                val uLen = input.readByte().toInt() and 0xFF
                val usernameBytes = ByteArray(uLen)
                input.readFully(usernameBytes)
                val username = String(usernameBytes)

                val pLen = input.readByte().toInt() and 0xFF
                val passwordBytes = ByteArray(pLen)
                input.readFully(passwordBytes)
                val password = String(passwordBytes)

                if (!credentialStore.validate(username, password)) {
                    output.write(byteArrayOf(USERPASS_VERSION, USERPASS_FAILURE))
                    output.flush()
                    Log.w(TAG, "Auth failed for user: $username")
                    return
                }

                // Auth success
                output.write(byteArrayOf(USERPASS_VERSION, USERPASS_SUCCESS))
                output.flush()
            } else {
                // No credentials configured — allow without auth
                output.write(byteArrayOf(SOCKS_VERSION, AUTH_NONE))
                output.flush()
            }

            // Request
            val reqVersion = input.readByte()
            val cmd = input.readByte()
            input.readByte() // reserved

            when (cmd) {
                CMD_CONNECT -> {
                    val atyp = input.readByte()
                    val host: String
                    when (atyp) {
                        ATYP_IPV4 -> {
                            val addr = ByteArray(4)
                            input.readFully(addr)
                            host = addr.joinToString(".") { (it.toInt() and 0xFF).toString() }
                        }
                        ATYP_DOMAIN -> {
                            val len = input.readByte().toInt() and 0xFF
                            val domain = ByteArray(len)
                            input.readFully(domain)
                            host = String(domain)
                        }
                        ATYP_IPV6 -> {
                            // Cellular networks often lack IPv6 — reject so client retries with IPv4
                            val addr = ByteArray(16)
                            input.readFully(addr)
                            input.readUnsignedShort() // consume port
                            Log.w(TAG, "IPv6 not supported, rejecting ATYP_IPV6 connect")
                            sendReply(output, 0x08) // Address type not supported
                            return
                        }
                        else -> return
                    }
                    val port = input.readUnsignedShort()

                    // Connect through cellular, fall back to default network
                    val targetSocket: Socket
                    try {
                        targetSocket = connectToTarget(host, port)
                    } catch (e: Exception) {
                        Log.e(TAG, "All connect methods failed for $host:$port: ${e.message}")
                        sendReply(output, 0x05) // Connection refused
                        return
                    }

                    // Success reply
                    sendReply(output, 0x00)

                    // Bidirectional relay
                    relay(clientSocket, targetSocket)
                }
                CMD_UDP_ASSOCIATE -> {
                    handleUdpAssociate(clientSocket, input, output)
                }
                else -> {
                    sendReply(output, 0x07) // Command not supported
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "Client error", e)
        } finally {
            clientSocket.close()
        }
    }

    private suspend fun handleUdpAssociate(
        controlSocket: Socket,
        input: DataInputStream,
        output: DataOutputStream
    ) {
        // Read and discard DST.ADDR/DST.PORT from the request
        readAndDiscardAddress(input)

        // Create target DatagramSocket — protected + cellular-bound
        val targetSocket = DatagramSocket(0)
        ProxyVpnService.protectSocket(targetSocket)
        networkManager.getCellularNetwork()?.bindSocket(targetSocket)
        Log.i(TAG, "UDP ASSOCIATE: target socket on port ${targetSocket.localPort}")

        // Reply success with BND.ADDR=0.0.0.0, BND.PORT=0
        // Clients interpret this as "use the same address/port as the TCP connection"
        sendReply(output, 0x00)

        // Register pending association so udpRelayLoop can match the first UDP packet
        val clientAddrDeferred = CompletableDeferred<InetSocketAddress>()
        val pending = PendingUdpAssociation(targetSocket, clientAddrDeferred)
        pendingAssociations.add(pending)

        // Response relay coroutine: target → client via udpRelaySocket
        val relayJob = scope.launch {
            try {
                val clientAddr = clientAddrDeferred.await()
                val buf = ByteArray(UDP_BUFFER_SIZE)
                while (isActive) {
                    val packet = DatagramPacket(buf, buf.size)
                    targetSocket.receive(packet)
                    _bytesIn.addAndGet(packet.length.toLong())

                    val response = buildSocksUdpHeader(
                        packet.address, packet.port,
                        packet.data, packet.length
                    )
                    val relay = udpRelaySocket ?: break
                    relay.send(DatagramPacket(
                        response, response.size,
                        clientAddr.address, clientAddr.port
                    ))
                }
            } catch (e: Exception) {
                if (isActive) Log.e(TAG, "UDP response relay error", e)
            }
        }

        // Block on TCP control connection — session stays alive while TCP is open
        try {
            while (running) {
                if (input.read() == -1) break
            }
        } catch (_: Exception) {}

        // Cleanup
        Log.i(TAG, "UDP ASSOCIATE: control connection closed, cleaning up")
        relayJob.cancel()
        targetSocket.close()
        pendingAssociations.remove(pending)
        udpSessions.entries.removeIf { it.value.targetSocket === targetSocket }
    }

    private fun readAndDiscardAddress(input: DataInputStream) {
        when (input.readByte()) {
            ATYP_IPV4 -> input.readFully(ByteArray(4))
            ATYP_DOMAIN -> {
                val len = input.readByte().toInt() and 0xFF
                input.readFully(ByteArray(len))
            }
            ATYP_IPV6 -> input.readFully(ByteArray(16))
        }
        input.readUnsignedShort() // port
    }

    private fun udpRelayLoop() {
        val buf = ByteArray(UDP_BUFFER_SIZE)
        while (running) {
            try {
                val packet = DatagramPacket(buf, buf.size)
                val relay = udpRelaySocket ?: break
                relay.receive(packet)

                val clientAddr = InetSocketAddress(packet.address, packet.port)
                val key = "${clientAddr.address.hostAddress}:${clientAddr.port}"

                // Look up or register session
                var session = udpSessions[key]
                if (session == null) {
                    val pending = pendingAssociations.poll()
                    if (pending == null) {
                        Log.w(TAG, "UDP from $key with no pending association, dropping")
                        continue
                    }
                    pending.clientAddrDeferred.complete(clientAddr)
                    session = UdpSession(clientAddr, pending.targetSocket)
                    udpSessions[key] = session
                    Log.i(TAG, "UDP ASSOCIATE: registered client $key")
                }

                // Parse SOCKS5 UDP header
                val data = packet.data
                val len = packet.length
                if (len < 4) continue

                // [0-1] RSV, [2] FRAG, [3] ATYP
                val frag = data[2].toInt() and 0xFF
                if (frag != 0) {
                    Log.w(TAG, "UDP fragmentation not supported, dropping")
                    continue
                }

                val atyp = data[3].toInt()
                var offset = 4
                val targetAddr: InetAddress
                when (atyp) {
                    ATYP_IPV4.toInt() -> {
                        if (len < offset + 6) continue
                        targetAddr = InetAddress.getByAddress(data.copyOfRange(offset, offset + 4))
                        offset += 4
                    }
                    ATYP_DOMAIN.toInt() -> {
                        val domainLen = data[offset].toInt() and 0xFF
                        offset++
                        if (len < offset + domainLen + 2) continue
                        val domain = String(data, offset, domainLen)
                        offset += domainLen
                        // Resolve via cellular DNS, prefer IPv4
                        val cellularNet = networkManager.getCellularNetwork()
                        targetAddr = if (cellularNet != null) {
                            val all = cellularNet.getAllByName(domain)
                            all.firstOrNull { it is Inet4Address } ?: all.first()
                        } else {
                            val all = InetAddress.getAllByName(domain)
                            all.firstOrNull { it is Inet4Address } ?: all.first()
                        }
                    }
                    ATYP_IPV6.toInt() -> {
                        if (len < offset + 18) continue
                        targetAddr = InetAddress.getByAddress(data.copyOfRange(offset, offset + 16))
                        offset += 16
                    }
                    else -> continue
                }

                val targetPort = ((data[offset].toInt() and 0xFF) shl 8) or
                        (data[offset + 1].toInt() and 0xFF)
                offset += 2

                val payloadLen = len - offset
                if (payloadLen <= 0) continue

                // Forward raw data to target via session's protected socket
                session.targetSocket.send(
                    DatagramPacket(data, offset, payloadLen, targetAddr, targetPort)
                )
                _bytesOut.addAndGet(payloadLen.toLong())
            } catch (e: Exception) {
                if (running) Log.e(TAG, "UDP relay loop error", e)
            }
        }
    }

    private fun buildSocksUdpHeader(
        srcAddr: InetAddress,
        srcPort: Int,
        data: ByteArray,
        dataLen: Int
    ): ByteArray {
        val addrBytes = srcAddr.address
        // RSV(2) + FRAG(1) + ATYP(1) + ADDR(4 or 16) + PORT(2)
        val headerLen = 4 + addrBytes.size + 2
        val result = ByteArray(headerLen + dataLen)
        // result[0], result[1] = 0x00 (RSV) — already zero
        // result[2] = 0x00 (FRAG) — already zero
        result[3] = if (addrBytes.size == 4) ATYP_IPV4 else ATYP_IPV6
        System.arraycopy(addrBytes, 0, result, 4, addrBytes.size)
        val portOffset = 4 + addrBytes.size
        result[portOffset] = (srcPort shr 8 and 0xFF).toByte()
        result[portOffset + 1] = (srcPort and 0xFF).toByte()
        System.arraycopy(data, 0, result, headerLen, dataLen)
        return result
    }

    private fun connectToTarget(host: String, port: Int): Socket {
        val cellularNet = networkManager.getCellularNetwork()

        // Resolve DNS via cellular if available, else system DNS
        // Prefer IPv4 — cellular networks often lack IPv6 routing
        val addr = if (cellularNet != null) {
            val all = cellularNet.getAllByName(host)
            all.firstOrNull { it is java.net.Inet4Address }
                ?: all.firstOrNull()
                ?: InetAddress.getByName(host)
        } else {
            val all = InetAddress.getAllByName(host)
            all.firstOrNull { it is java.net.Inet4Address }
                ?: all.first()
        }

        val socket = Socket()

        // Protect from VPN routing (so it doesn't go through tunnel)
        ProxyVpnService.protectSocket(socket)

        // Bind to cellular network
        cellularNet?.bindSocket(socket)

        socket.connect(InetSocketAddress(addr, port), 10000)
        Log.i(TAG, "Connected to $host:$port via CELLULAR (protected+bound)")
        return socket
    }

    private fun sendReply(output: DataOutputStream, status: Byte) {
        output.write(byteArrayOf(
            SOCKS_VERSION, status, 0x00, ATYP_IPV4,
            0, 0, 0, 0, // bind addr
            0, 0         // bind port
        ))
        output.flush()
    }

    private suspend fun relay(client: Socket, target: Socket) = coroutineScope {
        val job1 = launch {
            try {
                val buffer = ByteArray(BUFFER_SIZE)
                val input = client.getInputStream()
                val output = target.getOutputStream()
                while (true) {
                    val read = input.read(buffer)
                    if (read == -1) break
                    output.write(buffer, 0, read)
                    output.flush()
                    _bytesOut.addAndGet(read.toLong())
                }
            } catch (_: Exception) {}
            finally { try { target.shutdownOutput() } catch (_: Exception) {} }
        }
        val job2 = launch {
            try {
                val buffer = ByteArray(BUFFER_SIZE)
                val input = target.getInputStream()
                val output = client.getOutputStream()
                while (true) {
                    val read = input.read(buffer)
                    if (read == -1) break
                    output.write(buffer, 0, read)
                    output.flush()
                    _bytesIn.addAndGet(read.toLong())
                }
            } catch (_: Exception) {}
            finally { try { client.shutdownOutput() } catch (_: Exception) {} }
        }
        listOf(job1, job2).joinAll()
    }
}
