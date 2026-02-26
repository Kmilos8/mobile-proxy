package com.mobileproxy.core.vpn

import android.net.Network
import android.util.Log
import com.mobileproxy.core.network.NetworkManager
import com.mobileproxy.service.ProxyVpnService
import java.io.IOException
import java.net.DatagramPacket
import java.net.DatagramSocket
import java.net.InetAddress
import java.net.InetSocketAddress
import java.net.Socket
import java.nio.ByteBuffer
import java.nio.channels.SelectionKey
import java.nio.channels.Selector
import java.nio.channels.SocketChannel
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.ExecutorService
import java.util.concurrent.Executors
import java.util.concurrent.atomic.AtomicBoolean

/**
 * Userspace IP forwarder for non-rooted Android.
 *
 * When the server sends IP packets through the VPN tunnel that are NOT addressed to this phone
 * (i.e., they are OpenVPN client traffic being NAT-routed through this device), this class
 * forwards them through the cellular network to their real destination.
 *
 * Architecture:
 * - TCP: For each unique connection (srcPort+dstIP+dstPort), open a cellular-bound TCP socket,
 *   relay data bidirectionally. Response data gets packaged as IP packets and sent back.
 * - UDP: For each unique flow, open a cellular-bound DatagramSocket, relay datagrams.
 * - All outbound sockets are protected from VPN routing and bound to cellular.
 */
class IpForwarder(
    private val networkManager: NetworkManager,
    private val onResponse: (ByteArray, Int, Int) -> Unit // callback to send response IP packets back through tunnel
) {
    companion object {
        private const val TAG = "IpForwarder"
        private const val TCP_IDLE_TIMEOUT_MS = 60_000L
        private const val UDP_IDLE_TIMEOUT_MS = 30_000L
        private const val CLEANUP_INTERVAL_MS = 10_000L
        private const val TCP_BUFFER_SIZE = 65536
    }

    private val running = AtomicBoolean(false)
    private val tcpSessions = ConcurrentHashMap<String, TcpSession>()
    private val udpSessions = ConcurrentHashMap<String, UdpSession>()
    private val executor: ExecutorService = Executors.newCachedThreadPool { r ->
        Thread(r, "ipfwd-worker").apply { isDaemon = true }
    }

    fun start() {
        if (running.getAndSet(true)) return
        Log.i(TAG, "IP forwarder started")
        // Start cleanup thread
        executor.submit { cleanupLoop() }
    }

    fun stop() {
        if (!running.getAndSet(false)) return
        Log.i(TAG, "IP forwarder stopping")
        for (session in tcpSessions.values) session.close()
        for (session in udpSessions.values) session.close()
        tcpSessions.clear()
        udpSessions.clear()
        executor.shutdownNow()
    }

    /**
     * Forward an IP packet that arrived from the tunnel but is not addressed to this phone.
     * Called from the tunnel read loop.
     * [buf] contains the raw IP packet starting at [off] with length [len].
     */
    fun forward(buf: ByteArray, off: Int, len: Int) {
        if (!running.get()) return
        if (len < 20) return // too short for IPv4

        val version = (buf[off].toInt() shr 4) and 0x0F
        if (version != 4) return // only IPv4

        val proto = IpPacketUtils.protocol(buf, off)
        when (proto) {
            IpPacketUtils.PROTO_TCP -> handleTcp(buf, off, len)
            IpPacketUtils.PROTO_UDP -> handleUdp(buf, off, len)
            // ICMP and other protocols — drop silently
        }
    }

    private fun handleTcp(buf: ByteArray, off: Int, len: Int) {
        val ihl = IpPacketUtils.ihl(buf, off)
        if (len < ihl + 20) return // need at least IP header + TCP header

        val srcIp = IpPacketUtils.srcIP(buf, off)
        val dstIp = IpPacketUtils.dstIP(buf, off)
        val srcPort = IpPacketUtils.srcPort(buf, off)
        val dstPort = IpPacketUtils.dstPort(buf, off)

        val tcpOff = off + ihl
        val flags = buf[tcpOff + 13].toInt() and 0xFF
        val syn = (flags and 0x02) != 0
        val fin = (flags and 0x01) != 0
        val rst = (flags and 0x04) != 0
        val ack = (flags and 0x10) != 0

        val key = "${srcPort}-${ipToString(dstIp)}-${dstPort}"

        if (syn && !ack) {
            // New connection — SYN
            val existing = tcpSessions.remove(key)
            existing?.close()

            val session = TcpSession(key, srcIp, dstIp, srcPort, dstPort)
            tcpSessions[key] = session
            // Copy the SYN packet so we can process it asynchronously
            val pktCopy = buf.copyOfRange(off, off + len)
            executor.submit { session.connect(pktCopy) }
            return
        }

        val session = tcpSessions[key] ?: return

        if (rst) {
            session.close()
            tcpSessions.remove(key)
            return
        }

        if (fin) {
            // Copy data portion if any, then signal FIN
            val tcpHeaderLen = ((buf[tcpOff + 12].toInt() shr 4) and 0x0F) * 4
            val dataOff = tcpOff + tcpHeaderLen
            val dataLen = len - ihl - tcpHeaderLen
            if (dataLen > 0) {
                val data = buf.copyOfRange(dataOff, dataOff + dataLen)
                session.sendData(data)
            }
            session.sendFin()
            return
        }

        // Data packet (ACK with payload)
        val tcpHeaderLen = ((buf[tcpOff + 12].toInt() shr 4) and 0x0F) * 4
        val dataOff = tcpOff + tcpHeaderLen
        val dataLen = len - ihl - tcpHeaderLen
        if (dataLen > 0) {
            val data = buf.copyOfRange(dataOff, dataOff + dataLen)
            session.sendData(data)
        }
        session.touch()
    }

    private fun handleUdp(buf: ByteArray, off: Int, len: Int) {
        val ihl = IpPacketUtils.ihl(buf, off)
        if (len < ihl + 8) return // need IP header + UDP header

        val srcIp = IpPacketUtils.srcIP(buf, off)
        val dstIp = IpPacketUtils.dstIP(buf, off)
        val srcPort = IpPacketUtils.srcPort(buf, off)
        val dstPort = IpPacketUtils.dstPort(buf, off)

        val udpOff = off + ihl
        val udpDataOff = udpOff + 8
        val udpDataLen = len - ihl - 8
        if (udpDataLen <= 0) return

        val key = "${srcPort}-${ipToString(dstIp)}-${dstPort}"

        val session = udpSessions.getOrPut(key) {
            val s = UdpSession(key, srcIp, dstIp, srcPort, dstPort)
            executor.submit { s.startReceiver() }
            s
        }

        val payload = buf.copyOfRange(udpDataOff, udpDataOff + udpDataLen)
        session.send(payload)
        session.touch()
    }

    private fun cleanupLoop() {
        while (running.get()) {
            try {
                Thread.sleep(CLEANUP_INTERVAL_MS)
            } catch (_: InterruptedException) {
                return
            }

            val now = System.currentTimeMillis()

            val tcpExpired = tcpSessions.entries.filter { now - it.value.lastActive > TCP_IDLE_TIMEOUT_MS }
            for (entry in tcpExpired) {
                entry.value.close()
                tcpSessions.remove(entry.key)
                Log.d(TAG, "TCP session expired: ${entry.key}")
            }

            val udpExpired = udpSessions.entries.filter { now - it.value.lastActive > UDP_IDLE_TIMEOUT_MS }
            for (entry in udpExpired) {
                entry.value.close()
                udpSessions.remove(entry.key)
                Log.d(TAG, "UDP session expired: ${entry.key}")
            }
        }
    }

    /**
     * TCP session: maintains a cellular-bound socket to the destination.
     * Reads response data and sends it back as IP packets through the tunnel.
     */
    private inner class TcpSession(
        val key: String,
        val clientIp: ByteArray,   // original source IP (OpenVPN client, e.g. 10.9.0.x)
        val remoteIp: ByteArray,   // destination IP (internet host)
        val clientPort: Int,
        val remotePort: Int
    ) {
        @Volatile var lastActive = System.currentTimeMillis()
        private var socket: Socket? = null
        private val closed = AtomicBoolean(false)
        private var seqNum = 1000L // our sequence number for responses
        private var ackNum = 0L   // we track client's sequence number

        fun touch() { lastActive = System.currentTimeMillis() }

        fun connect(synPacket: ByteArray) {
            try {
                // Parse ISN from the SYN packet
                val ihl = IpPacketUtils.ihl(synPacket, 0)
                val tcpOff = ihl
                val clientISN = ((synPacket[tcpOff + 4].toLong() and 0xFF) shl 24) or
                        ((synPacket[tcpOff + 5].toLong() and 0xFF) shl 16) or
                        ((synPacket[tcpOff + 6].toLong() and 0xFF) shl 8) or
                        (synPacket[tcpOff + 7].toLong() and 0xFF)
                ackNum = clientISN + 1 // SYN consumes 1 sequence number

                val dstAddr = InetAddress.getByAddress(remoteIp)
                val sock = Socket()

                // Protect from VPN routing
                ProxyVpnService.protectSocket(sock)

                // Bind to cellular network
                val cellularNet = networkManager.getCellularNetwork()
                cellularNet?.bindSocket(sock)

                sock.connect(InetSocketAddress(dstAddr, remotePort), 10_000)
                sock.tcpNoDelay = true
                socket = sock
                touch()

                // Send SYN-ACK back to client
                sendTcpControl(0x12) // SYN+ACK
                seqNum++ // SYN-ACK consumes 1 seq

                Log.d(TAG, "TCP connected: $key -> ${dstAddr.hostAddress}:$remotePort")

                // Start reading responses from the remote socket
                val inputStream = sock.getInputStream()
                val readBuf = ByteArray(TCP_BUFFER_SIZE)
                while (!closed.get()) {
                    val n = inputStream.read(readBuf)
                    if (n <= 0) break
                    touch()
                    // Send data back as IP packet(s)
                    sendTcpData(readBuf, 0, n)
                }

                // Remote closed — send FIN to client
                if (!closed.get()) {
                    sendTcpControl(0x11) // FIN+ACK
                    seqNum++
                }
            } catch (e: IOException) {
                if (!closed.get()) {
                    Log.d(TAG, "TCP session $key error: ${e.message}")
                    // Send RST to client
                    try { sendTcpControl(0x14) } catch (_: Exception) {} // RST+ACK
                }
            } finally {
                close()
                tcpSessions.remove(key)
            }
        }

        fun sendData(data: ByteArray) {
            touch()
            ackNum += data.size
            try {
                socket?.getOutputStream()?.write(data)
                socket?.getOutputStream()?.flush()
                // Send ACK back to client
                sendTcpControl(0x10) // ACK
            } catch (e: IOException) {
                Log.d(TAG, "TCP write error for $key: ${e.message}")
                close()
                tcpSessions.remove(key)
            }
        }

        fun sendFin() {
            ackNum++ // FIN consumes 1 seq
            try {
                sendTcpControl(0x10) // ACK the FIN
                socket?.shutdownOutput()
            } catch (_: Exception) {}
        }

        /**
         * Build and send a TCP control packet (SYN-ACK, ACK, FIN, RST) back through the tunnel.
         */
        private fun sendTcpControl(flags: Int) {
            val ipLen = 20 + 20 // IP header + TCP header (no options)
            val pkt = ByteArray(ipLen)

            // IP header
            buildIpHeader(pkt, 0, ipLen, IpPacketUtils.PROTO_TCP, remoteIp, clientIp)

            // TCP header
            val tcpOff = 20
            // Source port (remote)
            pkt[tcpOff] = ((remotePort shr 8) and 0xFF).toByte()
            pkt[tcpOff + 1] = (remotePort and 0xFF).toByte()
            // Dest port (client)
            pkt[tcpOff + 2] = ((clientPort shr 8) and 0xFF).toByte()
            pkt[tcpOff + 3] = (clientPort and 0xFF).toByte()
            // Seq number
            pkt[tcpOff + 4] = ((seqNum shr 24) and 0xFF).toByte()
            pkt[tcpOff + 5] = ((seqNum shr 16) and 0xFF).toByte()
            pkt[tcpOff + 6] = ((seqNum shr 8) and 0xFF).toByte()
            pkt[tcpOff + 7] = (seqNum and 0xFF).toByte()
            // Ack number
            pkt[tcpOff + 8] = ((ackNum shr 24) and 0xFF).toByte()
            pkt[tcpOff + 9] = ((ackNum shr 16) and 0xFF).toByte()
            pkt[tcpOff + 10] = ((ackNum shr 8) and 0xFF).toByte()
            pkt[tcpOff + 11] = (ackNum and 0xFF).toByte()
            // Data offset (5 words = 20 bytes) + reserved
            pkt[tcpOff + 12] = 0x50.toByte()
            // Flags
            pkt[tcpOff + 13] = flags.toByte()
            // Window size (65535)
            pkt[tcpOff + 14] = 0xFF.toByte()
            pkt[tcpOff + 15] = 0xFF.toByte()

            IpPacketUtils.recomputeTcpChecksum(pkt, 0)
            onResponse(pkt, 0, ipLen)
        }

        /**
         * Build and send TCP data packet(s) back through the tunnel.
         */
        private fun sendTcpData(data: ByteArray, dataOff: Int, dataLen: Int) {
            // Fragment if needed (max ~1360 bytes payload per packet to fit in MTU 1400)
            val maxPayload = 1360
            var sent = 0
            while (sent < dataLen) {
                val chunkLen = minOf(maxPayload, dataLen - sent)
                val ipLen = 20 + 20 + chunkLen
                val pkt = ByteArray(ipLen)

                buildIpHeader(pkt, 0, ipLen, IpPacketUtils.PROTO_TCP, remoteIp, clientIp)

                val tcpOff = 20
                pkt[tcpOff] = ((remotePort shr 8) and 0xFF).toByte()
                pkt[tcpOff + 1] = (remotePort and 0xFF).toByte()
                pkt[tcpOff + 2] = ((clientPort shr 8) and 0xFF).toByte()
                pkt[tcpOff + 3] = (clientPort and 0xFF).toByte()
                // Seq
                pkt[tcpOff + 4] = ((seqNum shr 24) and 0xFF).toByte()
                pkt[tcpOff + 5] = ((seqNum shr 16) and 0xFF).toByte()
                pkt[tcpOff + 6] = ((seqNum shr 8) and 0xFF).toByte()
                pkt[tcpOff + 7] = (seqNum and 0xFF).toByte()
                // Ack
                pkt[tcpOff + 8] = ((ackNum shr 24) and 0xFF).toByte()
                pkt[tcpOff + 9] = ((ackNum shr 16) and 0xFF).toByte()
                pkt[tcpOff + 10] = ((ackNum shr 8) and 0xFF).toByte()
                pkt[tcpOff + 11] = (ackNum and 0xFF).toByte()
                pkt[tcpOff + 12] = 0x50.toByte() // data offset 5 words
                pkt[tcpOff + 13] = 0x18.toByte() // PSH+ACK
                pkt[tcpOff + 14] = 0xFF.toByte() // window
                pkt[tcpOff + 15] = 0xFF.toByte()

                System.arraycopy(data, dataOff + sent, pkt, 40, chunkLen)
                IpPacketUtils.recomputeTcpChecksum(pkt, 0)

                onResponse(pkt, 0, ipLen)
                seqNum += chunkLen
                sent += chunkLen
            }
        }

        fun close() {
            if (closed.getAndSet(true)) return
            try { socket?.close() } catch (_: Exception) {}
        }
    }

    /**
     * UDP session: maintains a cellular-bound DatagramSocket for a specific flow.
     */
    private inner class UdpSession(
        val key: String,
        val clientIp: ByteArray,
        val remoteIp: ByteArray,
        val clientPort: Int,
        val remotePort: Int
    ) {
        @Volatile var lastActive = System.currentTimeMillis()
        private val socket: DatagramSocket = DatagramSocket()
        private val closed = AtomicBoolean(false)

        init {
            ProxyVpnService.protectSocket(socket)
            val cellularNet = networkManager.getCellularNetwork()
            cellularNet?.bindSocket(socket)
            socket.soTimeout = 5000 // 5s receive timeout for the loop
        }

        fun touch() { lastActive = System.currentTimeMillis() }

        fun send(payload: ByteArray) {
            if (closed.get()) return
            try {
                val dstAddr = InetAddress.getByAddress(remoteIp)
                val pkt = DatagramPacket(payload, payload.size, dstAddr, remotePort)
                socket.send(pkt)
                touch()
            } catch (e: IOException) {
                Log.d(TAG, "UDP send error for $key: ${e.message}")
            }
        }

        fun startReceiver() {
            val recvBuf = ByteArray(65536)
            val recvPkt = DatagramPacket(recvBuf, recvBuf.size)

            while (!closed.get()) {
                try {
                    socket.receive(recvPkt)
                    touch()
                    val payloadLen = recvPkt.length
                    if (payloadLen > 0) {
                        sendUdpResponse(recvBuf, 0, payloadLen)
                    }
                } catch (_: java.net.SocketTimeoutException) {
                    // Normal — just loop and check if we should stop
                } catch (e: IOException) {
                    if (!closed.get()) {
                        Log.d(TAG, "UDP recv error for $key: ${e.message}")
                    }
                    break
                }
            }
        }

        private fun sendUdpResponse(data: ByteArray, dataOff: Int, dataLen: Int) {
            val udpLen = 8 + dataLen
            val ipLen = 20 + udpLen
            val pkt = ByteArray(ipLen)

            buildIpHeader(pkt, 0, ipLen, IpPacketUtils.PROTO_UDP, remoteIp, clientIp)

            val udpOff = 20
            // Source port (remote)
            pkt[udpOff] = ((remotePort shr 8) and 0xFF).toByte()
            pkt[udpOff + 1] = (remotePort and 0xFF).toByte()
            // Dest port (client)
            pkt[udpOff + 2] = ((clientPort shr 8) and 0xFF).toByte()
            pkt[udpOff + 3] = (clientPort and 0xFF).toByte()
            // UDP length
            pkt[udpOff + 4] = ((udpLen shr 8) and 0xFF).toByte()
            pkt[udpOff + 5] = (udpLen and 0xFF).toByte()
            // Checksum = 0 (optional for IPv4 UDP)
            pkt[udpOff + 6] = 0
            pkt[udpOff + 7] = 0

            System.arraycopy(data, dataOff, pkt, 28, dataLen)
            onResponse(pkt, 0, ipLen)
        }

        fun close() {
            if (closed.getAndSet(true)) return
            try { socket.close() } catch (_: Exception) {}
        }
    }

    /** Build a minimal 20-byte IPv4 header. */
    private fun buildIpHeader(pkt: ByteArray, off: Int, totalLen: Int, proto: Int, srcIp: ByteArray, dstIp: ByteArray) {
        pkt[off] = 0x45.toByte() // version 4, IHL 5
        pkt[off + 1] = 0 // DSCP/ECN
        pkt[off + 2] = ((totalLen shr 8) and 0xFF).toByte()
        pkt[off + 3] = (totalLen and 0xFF).toByte()
        // Identification (0)
        pkt[off + 4] = 0; pkt[off + 5] = 0
        // Flags + Fragment offset: DF bit set
        pkt[off + 6] = 0x40.toByte(); pkt[off + 7] = 0
        // TTL
        pkt[off + 8] = 64.toByte()
        // Protocol
        pkt[off + 9] = proto.toByte()
        // Checksum (will compute below)
        pkt[off + 10] = 0; pkt[off + 11] = 0
        // Source IP
        System.arraycopy(srcIp, 0, pkt, off + 12, 4)
        // Dest IP
        System.arraycopy(dstIp, 0, pkt, off + 16, 4)
        // Compute IP checksum
        val cksum = IpPacketUtils.checksum(pkt, off, 20)
        pkt[off + 10] = ((cksum shr 8) and 0xFF).toByte()
        pkt[off + 11] = (cksum and 0xFF).toByte()
    }

    private fun ipToString(ip: ByteArray): String =
        "${ip[0].toInt() and 0xFF}.${ip[1].toInt() and 0xFF}.${ip[2].toInt() and 0xFF}.${ip[3].toInt() and 0xFF}"
}
