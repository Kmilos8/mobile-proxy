package com.mobileproxy.core.proxy

import android.util.Log
import com.mobileproxy.core.network.NetworkManager
import kotlinx.coroutines.*
import java.io.BufferedReader
import java.io.InputStreamReader
import java.net.ServerSocket
import java.net.Socket
import java.util.concurrent.atomic.AtomicLong
import javax.inject.Inject
import javax.inject.Singleton

/**
 * HTTP CONNECT proxy server.
 * Accepts connections on the VPN interface and forwards them through cellular.
 */
@Singleton
class HttpProxyServer @Inject constructor(
    private val networkManager: NetworkManager
) {
    companion object {
        private const val TAG = "HttpProxyServer"
        private const val BUFFER_SIZE = 32768
    }

    private var serverSocket: ServerSocket? = null
    private var running = false
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    private val _bytesIn = AtomicLong(0)
    private val _bytesOut = AtomicLong(0)
    val bytesIn: Long get() = _bytesIn.get()
    val bytesOut: Long get() = _bytesOut.get()

    fun start(port: Int = 8080) {
        if (running) return
        running = true

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
            val reader = BufferedReader(InputStreamReader(clientSocket.getInputStream()))
            val requestLine = reader.readLine() ?: return

            if (requestLine.startsWith("CONNECT")) {
                handleConnect(clientSocket, requestLine, reader)
            } else {
                handlePlainHttp(clientSocket, requestLine, reader)
            }
        } catch (e: Exception) {
            Log.e(TAG, "Client error", e)
        } finally {
            clientSocket.close()
        }
    }

    private suspend fun handleConnect(
        clientSocket: Socket,
        requestLine: String,
        reader: BufferedReader
    ) {
        // Parse CONNECT host:port
        val parts = requestLine.split(" ")
        if (parts.size < 2) return
        val hostPort = parts[1].split(":")
        val host = hostPort[0]
        val port = hostPort.getOrNull(1)?.toIntOrNull() ?: 443

        // Consume remaining headers
        while (true) {
            val line = reader.readLine()
            if (line.isNullOrEmpty()) break
        }

        // Connect to target through cellular network using SocketFactory
        val targetSocket: Socket
        try {
            val addr = networkManager.resolveDnsCellular(host)
            targetSocket = networkManager.createCellularSocket(addr, port)
        } catch (e: Exception) {
            Log.e(TAG, "Failed to connect through cellular: ${e.message}")
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
        reader: BufferedReader
    ) {
        // Parse plain HTTP request (GET http://host/path HTTP/1.1)
        val parts = requestLine.split(" ")
        if (parts.size < 3) return

        val url = parts[1]
        val hostMatch = Regex("https?://([^/:]+)(:(\\d+))?").find(url) ?: return
        val host = hostMatch.groupValues[1]
        val port = hostMatch.groupValues[3].toIntOrNull() ?: 80

        // Read remaining headers
        val headers = StringBuilder()
        headers.appendLine(requestLine)
        while (true) {
            val line = reader.readLine()
            if (line.isNullOrEmpty()) break
            headers.appendLine(line)
        }
        headers.appendLine()

        // Connect through cellular using SocketFactory
        val targetSocket: Socket
        try {
            val addr = networkManager.resolveDnsCellular(host)
            targetSocket = networkManager.createCellularSocket(addr, port)
        } catch (e: Exception) {
            Log.e(TAG, "Failed to connect through cellular: ${e.message}")
            clientSocket.getOutputStream().write(
                "HTTP/1.1 502 Bad Gateway\r\n\r\n".toByteArray()
            )
            return
        }

        // Forward request
        targetSocket.getOutputStream().write(headers.toString().toByteArray())

        // Relay response
        relay(clientSocket, targetSocket)
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
