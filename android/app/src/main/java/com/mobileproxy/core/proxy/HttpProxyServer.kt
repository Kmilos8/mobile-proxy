package com.mobileproxy.core.proxy

import android.util.Base64
import android.util.Log
import com.mobileproxy.core.network.NetworkManager
import com.mobileproxy.service.ProxyVpnService
import kotlinx.coroutines.*
import java.io.BufferedReader
import java.io.InputStreamReader
import java.net.InetAddress
import java.net.InetSocketAddress
import java.net.ServerSocket
import java.net.Socket
import java.util.concurrent.atomic.AtomicLong
import javax.inject.Inject
import javax.inject.Singleton

/**
 * HTTP CONNECT proxy server with Proxy-Authorization (Basic) authentication.
 * Accepts connections on the VPN interface and forwards them through cellular.
 */
@Singleton
class HttpProxyServer @Inject constructor(
    private val networkManager: NetworkManager,
    private val credentialStore: ProxyCredentialStore
) {
    companion object {
        private const val TAG = "HttpProxyServer"
        private const val BUFFER_SIZE = 32768
        private const val AUTH_RESPONSE = "HTTP/1.1 407 Proxy Authentication Required\r\n" +
                "Proxy-Authenticate: Basic realm=\"MobileProxy\"\r\n" +
                "Content-Length: 0\r\n" +
                "\r\n"
    }

    private var serverSocket: ServerSocket? = null
    private var running = false
    private var scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    private val _bytesIn = AtomicLong(0)
    private val _bytesOut = AtomicLong(0)
    val bytesIn: Long get() = _bytesIn.get()
    val bytesOut: Long get() = _bytesOut.get()

    fun start(port: Int = 8080) {
        if (running) return
        running = true
        scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

        scope.launch {
            try {
                serverSocket = ServerSocket(port)
                Log.i(TAG, "HTTP proxy listening on port $port")

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
            clientSocket.tcpNoDelay = true
            clientSocket.soTimeout = 120_000 // 120s idle timeout
            val reader = BufferedReader(InputStreamReader(clientSocket.getInputStream()))
            val requestLine = reader.readLine() ?: return

            // Read all headers
            val headers = mutableListOf<String>()
            while (true) {
                val line = reader.readLine()
                if (line.isNullOrEmpty()) break
                headers.add(line)
            }

            // Check Proxy-Authorization if credentials are configured
            if (credentialStore.hasCredentials()) {
                val authHeader = headers.find {
                    it.startsWith("Proxy-Authorization:", ignoreCase = true)
                }

                if (authHeader == null) {
                    clientSocket.getOutputStream().write(AUTH_RESPONSE.toByteArray())
                    return
                }

                val authValue = authHeader.substringAfter(":").trim()
                if (!authValue.startsWith("Basic ", ignoreCase = true)) {
                    clientSocket.getOutputStream().write(AUTH_RESPONSE.toByteArray())
                    return
                }

                val decoded = try {
                    String(Base64.decode(authValue.substring(6), Base64.DEFAULT))
                } catch (e: Exception) {
                    clientSocket.getOutputStream().write(AUTH_RESPONSE.toByteArray())
                    return
                }

                val colonIdx = decoded.indexOf(':')
                if (colonIdx < 0) {
                    clientSocket.getOutputStream().write(AUTH_RESPONSE.toByteArray())
                    return
                }

                val username = decoded.substring(0, colonIdx)
                val password = decoded.substring(colonIdx + 1)

                if (!credentialStore.validate(username, password)) {
                    clientSocket.getOutputStream().write(AUTH_RESPONSE.toByteArray())
                    return
                }
            }

            if (requestLine.startsWith("CONNECT")) {
                handleConnect(clientSocket, requestLine)
            } else {
                handlePlainHttp(clientSocket, requestLine, headers)
            }
        } catch (e: Exception) {
            Log.e(TAG, "Client error", e)
        } finally {
            clientSocket.close()
        }
    }

    /**
     * Connect to target host via cellular network.
     * Socket is protected from VPN routing and bound to cellular.
     */
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
        socket.tcpNoDelay = true
        socket.soTimeout = 120_000 // 120s idle timeout to prevent stuck connections
        Log.i(TAG, "Connected to $host:$port via CELLULAR (protected+bound)")
        return socket
    }

    private suspend fun handleConnect(
        clientSocket: Socket,
        requestLine: String
    ) {
        // Parse CONNECT host:port
        val parts = requestLine.split(" ")
        if (parts.size < 2) return
        val hostPort = parts[1].split(":")
        val host = hostPort[0]
        val port = hostPort.getOrNull(1)?.toIntOrNull() ?: 443

        // Connect to target
        val targetSocket: Socket
        try {
            targetSocket = connectToTarget(host, port)
        } catch (e: Exception) {
            Log.e(TAG, "All connect methods failed for $host:$port: ${e.message}")
            clientSocket.getOutputStream().write(
                "HTTP/1.1 502 Bad Gateway\r\n\r\n".toByteArray()
            )
            return
        }

        // Send 200 Connection Established
        clientSocket.getOutputStream().write(
            "HTTP/1.1 200 Connection Established\r\n\r\n".toByteArray()
        )

        // Bidirectional relay
        relay(clientSocket, targetSocket)
    }

    private suspend fun handlePlainHttp(
        clientSocket: Socket,
        requestLine: String,
        headers: List<String>
    ) {
        // Parse plain HTTP request (GET http://host/path HTTP/1.1)
        val parts = requestLine.split(" ")
        if (parts.size < 3) return

        val url = parts[1]
        val hostMatch = Regex("https?://([^/:]+)(:(\\d+))?").find(url) ?: return
        val host = hostMatch.groupValues[1]
        val port = hostMatch.groupValues[3].toIntOrNull() ?: 80

        // Rebuild request with headers (excluding Proxy-Authorization)
        val requestBuf = StringBuilder()
        requestBuf.appendLine(requestLine)
        for (header in headers) {
            if (!header.startsWith("Proxy-Authorization:", ignoreCase = true)) {
                requestBuf.appendLine(header)
            }
        }
        requestBuf.appendLine()

        // Connect to target
        val targetSocket: Socket
        try {
            targetSocket = connectToTarget(host, port)
        } catch (e: Exception) {
            Log.e(TAG, "All connect methods failed for $host:$port: ${e.message}")
            clientSocket.getOutputStream().write(
                "HTTP/1.1 502 Bad Gateway\r\n\r\n".toByteArray()
            )
            return
        }

        // Forward request
        targetSocket.getOutputStream().write(requestBuf.toString().toByteArray())

        // Relay response
        relay(clientSocket, targetSocket)
    }

    private suspend fun relay(client: Socket, target: Socket) = coroutineScope {
        val job1 = launch(Dispatchers.IO) {
            try {
                val buffer = ByteArray(BUFFER_SIZE)
                val input = client.getInputStream()
                val output = target.getOutputStream()
                while (true) {
                    val read = input.read(buffer)
                    if (read == -1) break
                    output.write(buffer, 0, read)
                    _bytesOut.addAndGet(read.toLong())
                }
            } catch (_: Exception) {}
            finally { try { target.shutdownOutput() } catch (_: Exception) {} }
        }
        val job2 = launch(Dispatchers.IO) {
            try {
                val buffer = ByteArray(BUFFER_SIZE)
                val input = target.getInputStream()
                val output = client.getOutputStream()
                while (true) {
                    val read = input.read(buffer)
                    if (read == -1) break
                    output.write(buffer, 0, read)
                    _bytesIn.addAndGet(read.toLong())
                }
            } catch (_: Exception) {}
            finally { try { client.shutdownOutput() } catch (_: Exception) {} }
        }
        listOf(job1, job2).joinAll()
    }
}
