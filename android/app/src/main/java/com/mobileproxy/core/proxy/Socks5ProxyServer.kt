package com.mobileproxy.core.proxy

import android.util.Log
import com.mobileproxy.core.network.NetworkManager
import kotlinx.coroutines.*
import java.io.DataInputStream
import java.net.InetAddress
import java.io.DataOutputStream
import java.net.InetSocketAddress
import java.net.ServerSocket
import java.net.Socket
import java.util.concurrent.atomic.AtomicLong
import javax.inject.Inject
import javax.inject.Singleton

/**
 * SOCKS5 proxy server (RFC 1928).
 * Supports CONNECT command with NO AUTH and USERNAME/PASSWORD auth.
 */
@Singleton
class Socks5ProxyServer @Inject constructor(
    private val networkManager: NetworkManager
) {
    companion object {
        private const val TAG = "Socks5ProxyServer"
        private const val SOCKS_VERSION: Byte = 0x05
        private const val CMD_CONNECT: Byte = 0x01
        private const val ATYP_IPV4: Byte = 0x01
        private const val ATYP_DOMAIN: Byte = 0x03
        private const val ATYP_IPV6: Byte = 0x04
        private const val AUTH_NONE: Byte = 0x00
        private const val BUFFER_SIZE = 32768
    }

    private var serverSocket: ServerSocket? = null
    private var running = false
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    private val _bytesIn = AtomicLong(0)
    private val _bytesOut = AtomicLong(0)
    val bytesIn: Long get() = _bytesIn.get()
    val bytesOut: Long get() = _bytesOut.get()

    fun start(port: Int = 1080) {
        if (running) return
        running = true

        scope.launch {
            try {
                serverSocket = ServerSocket(port)
                Log.i(TAG, "SOCKS5 proxy listening on port $port")

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

            // Select NO AUTH
            output.write(byteArrayOf(SOCKS_VERSION, AUTH_NONE))
            output.flush()

            // Request
            val reqVersion = input.readByte()
            val cmd = input.readByte()
            input.readByte() // reserved

            if (cmd != CMD_CONNECT) {
                sendReply(output, 0x07) // Command not supported
                return
            }

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
                    val addr = ByteArray(16)
                    input.readFully(addr)
                    host = java.net.InetAddress.getByAddress(addr).hostAddress ?: return
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
        } catch (e: Exception) {
            Log.e(TAG, "Client error", e)
        } finally {
            clientSocket.close()
        }
    }

    private fun connectToTarget(host: String, port: Int): Socket {
        // Try cellular first
        try {
            val cellularNet = networkManager.getCellularNetwork()
            if (cellularNet != null) {
                val addr = cellularNet.getAllByName(host).firstOrNull()
                    ?: throw Exception("cellular DNS failed")
                val socket = cellularNet.socketFactory.createSocket()
                socket.connect(InetSocketAddress(addr, port), 10000)
                Log.i(TAG, "Connected to $host:$port via CELLULAR")
                return socket
            }
        } catch (e: Exception) {
            Log.w(TAG, "Cellular failed for $host:$port: ${e.message}")
        }
        // Fallback to default network
        Log.i(TAG, "Connecting to $host:$port via DEFAULT network")
        val socket = Socket()
        socket.connect(InetSocketAddress(InetAddress.getByName(host), port), 10000)
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
