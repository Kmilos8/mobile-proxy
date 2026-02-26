package com.mobileproxy.core.vpn

/**
 * Utility functions for parsing and constructing IPv4 packets.
 * Used by IpForwarder to inspect incoming packets and build response packets.
 */
object IpPacketUtils {

    const val PROTO_TCP: Int = 6
    const val PROTO_UDP: Int = 17

    /** Extract IPv4 header length in bytes (IHL field * 4). */
    fun ihl(buf: ByteArray, off: Int): Int = (buf[off].toInt() and 0x0F) * 4

    /** Extract total length from IPv4 header. */
    fun totalLength(buf: ByteArray, off: Int): Int =
        ((buf[off + 2].toInt() and 0xFF) shl 8) or (buf[off + 3].toInt() and 0xFF)

    /** Extract protocol field from IPv4 header. */
    fun protocol(buf: ByteArray, off: Int): Int = buf[off + 9].toInt() and 0xFF

    /** Extract source IP as a 4-byte array. */
    fun srcIP(buf: ByteArray, off: Int): ByteArray = buf.copyOfRange(off + 12, off + 16)

    /** Extract destination IP as a 4-byte array. */
    fun dstIP(buf: ByteArray, off: Int): ByteArray = buf.copyOfRange(off + 16, off + 20)

    /** Extract destination IP as dotted string. */
    fun dstIPString(buf: ByteArray, off: Int): String =
        "${buf[off + 16].toInt() and 0xFF}.${buf[off + 17].toInt() and 0xFF}.${buf[off + 18].toInt() and 0xFF}.${buf[off + 19].toInt() and 0xFF}"

    /** Extract source port from TCP/UDP header (starts at off + ihl). */
    fun srcPort(buf: ByteArray, off: Int): Int {
        val hdrOff = off + ihl(buf, off)
        return ((buf[hdrOff].toInt() and 0xFF) shl 8) or (buf[hdrOff + 1].toInt() and 0xFF)
    }

    /** Extract destination port from TCP/UDP header. */
    fun dstPort(buf: ByteArray, off: Int): Int {
        val hdrOff = off + ihl(buf, off)
        return ((buf[hdrOff + 2].toInt() and 0xFF) shl 8) or (buf[hdrOff + 3].toInt() and 0xFF)
    }

    /**
     * Compute one's-complement checksum over a range of bytes.
     * Used for IP header checksum and TCP/UDP pseudo-header checksum.
     */
    fun checksum(buf: ByteArray, off: Int, len: Int): Int {
        var sum = 0L
        var i = 0
        while (i < len - 1) {
            sum += ((buf[off + i].toInt() and 0xFF) shl 8) or (buf[off + i + 1].toInt() and 0xFF)
            i += 2
        }
        if (i < len) {
            sum += (buf[off + i].toInt() and 0xFF) shl 8
        }
        while (sum shr 16 != 0L) {
            sum = (sum and 0xFFFF) + (sum shr 16)
        }
        return (sum.toInt().inv()) and 0xFFFF
    }

    /**
     * Swap source and destination IP in an IPv4 packet (in-place),
     * replacing the source with [newSrc] and destination with [newDst].
     * Recomputes the IP header checksum.
     */
    fun rewriteIpAddresses(buf: ByteArray, off: Int, newSrc: ByteArray, newDst: ByteArray) {
        System.arraycopy(newSrc, 0, buf, off + 12, 4)
        System.arraycopy(newDst, 0, buf, off + 16, 4)
        // Zero checksum field and recompute
        buf[off + 10] = 0
        buf[off + 11] = 0
        val cksum = checksum(buf, off, ihl(buf, off))
        buf[off + 10] = ((cksum shr 8) and 0xFF).toByte()
        buf[off + 11] = (cksum and 0xFF).toByte()
    }

    /**
     * Recompute TCP checksum for a packet in [buf] starting at [ipOff].
     * Zeroes the existing checksum, computes pseudo-header + TCP segment checksum.
     */
    fun recomputeTcpChecksum(buf: ByteArray, ipOff: Int) {
        val hdrLen = ihl(buf, ipOff)
        val tcpOff = ipOff + hdrLen
        val ipTotalLen = totalLength(buf, ipOff)
        val tcpLen = ipTotalLen - hdrLen

        // Zero existing TCP checksum
        buf[tcpOff + 16] = 0
        buf[tcpOff + 17] = 0

        // Pseudo-header: srcIP(4) + dstIP(4) + 0 + proto(1) + tcpLen(2) = 12 bytes
        val pseudo = ByteArray(12)
        System.arraycopy(buf, ipOff + 12, pseudo, 0, 4) // src IP
        System.arraycopy(buf, ipOff + 16, pseudo, 4, 4) // dst IP
        pseudo[8] = 0
        pseudo[9] = PROTO_TCP.toByte()
        pseudo[10] = ((tcpLen shr 8) and 0xFF).toByte()
        pseudo[11] = (tcpLen and 0xFF).toByte()

        // Sum pseudo-header
        var sum = 0L
        for (i in 0 until 12 step 2) {
            sum += ((pseudo[i].toInt() and 0xFF) shl 8) or (pseudo[i + 1].toInt() and 0xFF)
        }
        // Sum TCP segment
        var i = 0
        while (i < tcpLen - 1) {
            sum += ((buf[tcpOff + i].toInt() and 0xFF) shl 8) or (buf[tcpOff + i + 1].toInt() and 0xFF)
            i += 2
        }
        if (i < tcpLen) {
            sum += (buf[tcpOff + i].toInt() and 0xFF) shl 8
        }
        while (sum shr 16 != 0L) {
            sum = (sum and 0xFFFF) + (sum shr 16)
        }
        val cksum = (sum.toInt().inv()) and 0xFFFF
        buf[tcpOff + 16] = ((cksum shr 8) and 0xFF).toByte()
        buf[tcpOff + 17] = (cksum and 0xFF).toByte()
    }

    /**
     * Recompute UDP checksum for a packet in [buf] starting at [ipOff].
     * UDP checksum is optional for IPv4 — we set it to 0 (disabled).
     */
    fun clearUdpChecksum(buf: ByteArray, ipOff: Int) {
        val udpOff = ipOff + ihl(buf, ipOff)
        buf[udpOff + 6] = 0
        buf[udpOff + 7] = 0
    }
}
