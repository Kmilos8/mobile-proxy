package com.mobileproxy.core.network

import java.io.IOException
import java.net.InetAddress
import java.net.InetSocketAddress
import java.net.Socket
import javax.inject.Inject
import javax.inject.Singleton
import javax.net.SocketFactory

/**
 * Socket factory that creates sockets bound to the cellular network.
 * Ensures all proxy outbound traffic goes through mobile IP, not WiFi.
 */
@Singleton
class CellularSocketFactory @Inject constructor(
    private val networkManager: NetworkManager
) : SocketFactory() {

    override fun createSocket(): Socket {
        val socket = Socket()
        networkManager.bindSocketToCellular(socket)
        return socket
    }

    override fun createSocket(host: String, port: Int): Socket {
        val socket = Socket()
        networkManager.bindSocketToCellular(socket)
        socket.connect(InetSocketAddress(host, port))
        return socket
    }

    override fun createSocket(host: String, port: Int, localAddr: InetAddress, localPort: Int): Socket {
        val socket = Socket()
        networkManager.bindSocketToCellular(socket)
        socket.bind(InetSocketAddress(localAddr, localPort))
        socket.connect(InetSocketAddress(host, port))
        return socket
    }

    override fun createSocket(addr: InetAddress, port: Int): Socket {
        val socket = Socket()
        networkManager.bindSocketToCellular(socket)
        socket.connect(InetSocketAddress(addr, port))
        return socket
    }

    override fun createSocket(addr: InetAddress, port: Int, localAddr: InetAddress, localPort: Int): Socket {
        val socket = Socket()
        networkManager.bindSocketToCellular(socket)
        socket.bind(InetSocketAddress(localAddr, localPort))
        socket.connect(InetSocketAddress(addr, port))
        return socket
    }
}
