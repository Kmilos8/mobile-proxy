package com.mobileproxy.core.proxy;

/**
 * SOCKS5 proxy server (RFC 1928).
 * Supports CONNECT and UDP ASSOCIATE commands with NO AUTH.
 */
@javax.inject.Singleton()
@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000\u008c\u0001\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0010\t\n\u0002\b\u0005\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\u000b\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0002\u0010\u000e\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\u0012\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\b\n\u0002\b\u0003\n\u0002\u0018\u0002\n\u0002\b\u0003\n\u0002\u0010\u0002\n\u0002\b\u0005\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0002\b\b\n\u0002\u0010\u0005\n\u0002\b\u0007\b\u0007\u0018\u0000 ?2\u00020\u0001:\u0003?@AB\u000f\b\u0007\u0012\u0006\u0010\u0002\u001a\u00020\u0003\u00a2\u0006\u0002\u0010\u0004J(\u0010\u001d\u001a\u00020\u001e2\u0006\u0010\u001f\u001a\u00020 2\u0006\u0010!\u001a\u00020\"2\u0006\u0010#\u001a\u00020\u001e2\u0006\u0010$\u001a\u00020\"H\u0002J\u0018\u0010%\u001a\u00020&2\u0006\u0010\'\u001a\u00020\u001b2\u0006\u0010(\u001a\u00020\"H\u0002J\u0016\u0010)\u001a\u00020*2\u0006\u0010+\u001a\u00020&H\u0082@\u00a2\u0006\u0002\u0010,J&\u0010-\u001a\u00020*2\u0006\u0010.\u001a\u00020&2\u0006\u0010/\u001a\u0002002\u0006\u00101\u001a\u000202H\u0082@\u00a2\u0006\u0002\u00103J\u0010\u00104\u001a\u00020*2\u0006\u0010/\u001a\u000200H\u0002J\u001e\u00105\u001a\u00020*2\u0006\u00106\u001a\u00020&2\u0006\u00107\u001a\u00020&H\u0082@\u00a2\u0006\u0002\u00108J\u0018\u00109\u001a\u00020*2\u0006\u00101\u001a\u0002022\u0006\u0010:\u001a\u00020;H\u0002J\u0010\u0010<\u001a\u00020*2\b\b\u0002\u0010(\u001a\u00020\"J\u0006\u0010=\u001a\u00020*J\b\u0010>\u001a\u00020*H\u0002R\u000e\u0010\u0005\u001a\u00020\u0006X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0007\u001a\u00020\u0006X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u0011\u0010\b\u001a\u00020\t8F\u00a2\u0006\u0006\u001a\u0004\b\n\u0010\u000bR\u0011\u0010\f\u001a\u00020\t8F\u00a2\u0006\u0006\u001a\u0004\b\r\u0010\u000bR\u000e\u0010\u0002\u001a\u00020\u0003X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u0014\u0010\u000e\u001a\b\u0012\u0004\u0012\u00020\u00100\u000fX\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0011\u001a\u00020\u0012X\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0013\u001a\u00020\u0014X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u0010\u0010\u0015\u001a\u0004\u0018\u00010\u0016X\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u0010\u0010\u0017\u001a\u0004\u0018\u00010\u0018X\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u001a\u0010\u0019\u001a\u000e\u0012\u0004\u0012\u00020\u001b\u0012\u0004\u0012\u00020\u001c0\u001aX\u0082\u0004\u00a2\u0006\u0002\n\u0000\u00a8\u0006B"}, d2 = {"Lcom/mobileproxy/core/proxy/Socks5ProxyServer;", "", "networkManager", "Lcom/mobileproxy/core/network/NetworkManager;", "(Lcom/mobileproxy/core/network/NetworkManager;)V", "_bytesIn", "Ljava/util/concurrent/atomic/AtomicLong;", "_bytesOut", "bytesIn", "", "getBytesIn", "()J", "bytesOut", "getBytesOut", "pendingAssociations", "Ljava/util/concurrent/ConcurrentLinkedQueue;", "Lcom/mobileproxy/core/proxy/Socks5ProxyServer$PendingUdpAssociation;", "running", "", "scope", "Lkotlinx/coroutines/CoroutineScope;", "serverSocket", "Ljava/net/ServerSocket;", "udpRelaySocket", "Ljava/net/DatagramSocket;", "udpSessions", "Ljava/util/concurrent/ConcurrentHashMap;", "", "Lcom/mobileproxy/core/proxy/Socks5ProxyServer$UdpSession;", "buildSocksUdpHeader", "", "srcAddr", "Ljava/net/InetAddress;", "srcPort", "", "data", "dataLen", "connectToTarget", "Ljava/net/Socket;", "host", "port", "handleClient", "", "clientSocket", "(Ljava/net/Socket;Lkotlin/coroutines/Continuation;)Ljava/lang/Object;", "handleUdpAssociate", "controlSocket", "input", "Ljava/io/DataInputStream;", "output", "Ljava/io/DataOutputStream;", "(Ljava/net/Socket;Ljava/io/DataInputStream;Ljava/io/DataOutputStream;Lkotlin/coroutines/Continuation;)Ljava/lang/Object;", "readAndDiscardAddress", "relay", "client", "target", "(Ljava/net/Socket;Ljava/net/Socket;Lkotlin/coroutines/Continuation;)Ljava/lang/Object;", "sendReply", "status", "", "start", "stop", "udpRelayLoop", "Companion", "PendingUdpAssociation", "UdpSession", "app_debug"})
public final class Socks5ProxyServer {
    @org.jetbrains.annotations.NotNull()
    private final com.mobileproxy.core.network.NetworkManager networkManager = null;
    @org.jetbrains.annotations.NotNull()
    private static final java.lang.String TAG = "Socks5ProxyServer";
    private static final byte SOCKS_VERSION = (byte)5;
    private static final byte CMD_CONNECT = (byte)1;
    private static final byte CMD_UDP_ASSOCIATE = (byte)3;
    private static final byte ATYP_IPV4 = (byte)1;
    private static final byte ATYP_DOMAIN = (byte)3;
    private static final byte ATYP_IPV6 = (byte)4;
    private static final byte AUTH_NONE = (byte)0;
    private static final int BUFFER_SIZE = 32768;
    private static final int UDP_BUFFER_SIZE = 65535;
    @org.jetbrains.annotations.Nullable()
    private java.net.ServerSocket serverSocket;
    @org.jetbrains.annotations.Nullable()
    private java.net.DatagramSocket udpRelaySocket;
    private boolean running = false;
    @org.jetbrains.annotations.NotNull()
    private final kotlinx.coroutines.CoroutineScope scope = null;
    @org.jetbrains.annotations.NotNull()
    private final java.util.concurrent.atomic.AtomicLong _bytesIn = null;
    @org.jetbrains.annotations.NotNull()
    private final java.util.concurrent.atomic.AtomicLong _bytesOut = null;
    @org.jetbrains.annotations.NotNull()
    private final java.util.concurrent.ConcurrentHashMap<java.lang.String, com.mobileproxy.core.proxy.Socks5ProxyServer.UdpSession> udpSessions = null;
    @org.jetbrains.annotations.NotNull()
    private final java.util.concurrent.ConcurrentLinkedQueue<com.mobileproxy.core.proxy.Socks5ProxyServer.PendingUdpAssociation> pendingAssociations = null;
    @org.jetbrains.annotations.NotNull()
    public static final com.mobileproxy.core.proxy.Socks5ProxyServer.Companion Companion = null;
    
    @javax.inject.Inject()
    public Socks5ProxyServer(@org.jetbrains.annotations.NotNull()
    com.mobileproxy.core.network.NetworkManager networkManager) {
        super();
    }
    
    public final long getBytesIn() {
        return 0L;
    }
    
    public final long getBytesOut() {
        return 0L;
    }
    
    public final void start(int port) {
    }
    
    public final void stop() {
    }
    
    private final java.lang.Object handleClient(java.net.Socket clientSocket, kotlin.coroutines.Continuation<? super kotlin.Unit> $completion) {
        return null;
    }
    
    private final java.lang.Object handleUdpAssociate(java.net.Socket controlSocket, java.io.DataInputStream input, java.io.DataOutputStream output, kotlin.coroutines.Continuation<? super kotlin.Unit> $completion) {
        return null;
    }
    
    private final void readAndDiscardAddress(java.io.DataInputStream input) {
    }
    
    private final void udpRelayLoop() {
    }
    
    private final byte[] buildSocksUdpHeader(java.net.InetAddress srcAddr, int srcPort, byte[] data, int dataLen) {
        return null;
    }
    
    private final java.net.Socket connectToTarget(java.lang.String host, int port) {
        return null;
    }
    
    private final void sendReply(java.io.DataOutputStream output, byte status) {
    }
    
    private final java.lang.Object relay(java.net.Socket client, java.net.Socket target, kotlin.coroutines.Continuation<? super kotlin.Unit> $completion) {
        return null;
    }
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000$\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\u0005\n\u0002\b\u0004\n\u0002\u0010\b\n\u0002\b\u0004\n\u0002\u0010\u000e\n\u0002\b\u0002\b\u0086\u0003\u0018\u00002\u00020\u0001B\u0007\b\u0002\u00a2\u0006\u0002\u0010\u0002R\u000e\u0010\u0003\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0005\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0006\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0007\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\b\u001a\u00020\tX\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\n\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u000b\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\f\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\r\u001a\u00020\u000eX\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u000f\u001a\u00020\tX\u0082T\u00a2\u0006\u0002\n\u0000\u00a8\u0006\u0010"}, d2 = {"Lcom/mobileproxy/core/proxy/Socks5ProxyServer$Companion;", "", "()V", "ATYP_DOMAIN", "", "ATYP_IPV4", "ATYP_IPV6", "AUTH_NONE", "BUFFER_SIZE", "", "CMD_CONNECT", "CMD_UDP_ASSOCIATE", "SOCKS_VERSION", "TAG", "", "UDP_BUFFER_SIZE", "app_debug"})
    public static final class Companion {
        
        private Companion() {
            super();
        }
    }
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u00000\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0002\b\t\n\u0002\u0010\u000b\n\u0002\b\u0002\n\u0002\u0010\b\n\u0000\n\u0002\u0010\u000e\n\u0000\b\u0086\b\u0018\u00002\u00020\u0001B\u001b\u0012\u0006\u0010\u0002\u001a\u00020\u0003\u0012\f\u0010\u0004\u001a\b\u0012\u0004\u0012\u00020\u00060\u0005\u00a2\u0006\u0002\u0010\u0007J\t\u0010\f\u001a\u00020\u0003H\u00c6\u0003J\u000f\u0010\r\u001a\b\u0012\u0004\u0012\u00020\u00060\u0005H\u00c6\u0003J#\u0010\u000e\u001a\u00020\u00002\b\b\u0002\u0010\u0002\u001a\u00020\u00032\u000e\b\u0002\u0010\u0004\u001a\b\u0012\u0004\u0012\u00020\u00060\u0005H\u00c6\u0001J\u0013\u0010\u000f\u001a\u00020\u00102\b\u0010\u0011\u001a\u0004\u0018\u00010\u0001H\u00d6\u0003J\t\u0010\u0012\u001a\u00020\u0013H\u00d6\u0001J\t\u0010\u0014\u001a\u00020\u0015H\u00d6\u0001R\u0017\u0010\u0004\u001a\b\u0012\u0004\u0012\u00020\u00060\u0005\u00a2\u0006\b\n\u0000\u001a\u0004\b\b\u0010\tR\u0011\u0010\u0002\u001a\u00020\u0003\u00a2\u0006\b\n\u0000\u001a\u0004\b\n\u0010\u000b\u00a8\u0006\u0016"}, d2 = {"Lcom/mobileproxy/core/proxy/Socks5ProxyServer$PendingUdpAssociation;", "", "targetSocket", "Ljava/net/DatagramSocket;", "clientAddrDeferred", "Lkotlinx/coroutines/CompletableDeferred;", "Ljava/net/InetSocketAddress;", "(Ljava/net/DatagramSocket;Lkotlinx/coroutines/CompletableDeferred;)V", "getClientAddrDeferred", "()Lkotlinx/coroutines/CompletableDeferred;", "getTargetSocket", "()Ljava/net/DatagramSocket;", "component1", "component2", "copy", "equals", "", "other", "hashCode", "", "toString", "", "app_debug"})
    public static final class PendingUdpAssociation {
        @org.jetbrains.annotations.NotNull()
        private final java.net.DatagramSocket targetSocket = null;
        @org.jetbrains.annotations.NotNull()
        private final kotlinx.coroutines.CompletableDeferred<java.net.InetSocketAddress> clientAddrDeferred = null;
        
        public PendingUdpAssociation(@org.jetbrains.annotations.NotNull()
        java.net.DatagramSocket targetSocket, @org.jetbrains.annotations.NotNull()
        kotlinx.coroutines.CompletableDeferred<java.net.InetSocketAddress> clientAddrDeferred) {
            super();
        }
        
        @org.jetbrains.annotations.NotNull()
        public final java.net.DatagramSocket getTargetSocket() {
            return null;
        }
        
        @org.jetbrains.annotations.NotNull()
        public final kotlinx.coroutines.CompletableDeferred<java.net.InetSocketAddress> getClientAddrDeferred() {
            return null;
        }
        
        @org.jetbrains.annotations.NotNull()
        public final java.net.DatagramSocket component1() {
            return null;
        }
        
        @org.jetbrains.annotations.NotNull()
        public final kotlinx.coroutines.CompletableDeferred<java.net.InetSocketAddress> component2() {
            return null;
        }
        
        @org.jetbrains.annotations.NotNull()
        public final com.mobileproxy.core.proxy.Socks5ProxyServer.PendingUdpAssociation copy(@org.jetbrains.annotations.NotNull()
        java.net.DatagramSocket targetSocket, @org.jetbrains.annotations.NotNull()
        kotlinx.coroutines.CompletableDeferred<java.net.InetSocketAddress> clientAddrDeferred) {
            return null;
        }
        
        @java.lang.Override()
        public boolean equals(@org.jetbrains.annotations.Nullable()
        java.lang.Object other) {
            return false;
        }
        
        @java.lang.Override()
        public int hashCode() {
            return 0;
        }
        
        @java.lang.Override()
        @org.jetbrains.annotations.NotNull()
        public java.lang.String toString() {
            return null;
        }
    }
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000,\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0002\b\t\n\u0002\u0010\u000b\n\u0002\b\u0002\n\u0002\u0010\b\n\u0000\n\u0002\u0010\u000e\n\u0000\b\u0086\b\u0018\u00002\u00020\u0001B\u0015\u0012\u0006\u0010\u0002\u001a\u00020\u0003\u0012\u0006\u0010\u0004\u001a\u00020\u0005\u00a2\u0006\u0002\u0010\u0006J\t\u0010\u000b\u001a\u00020\u0003H\u00c6\u0003J\t\u0010\f\u001a\u00020\u0005H\u00c6\u0003J\u001d\u0010\r\u001a\u00020\u00002\b\b\u0002\u0010\u0002\u001a\u00020\u00032\b\b\u0002\u0010\u0004\u001a\u00020\u0005H\u00c6\u0001J\u0013\u0010\u000e\u001a\u00020\u000f2\b\u0010\u0010\u001a\u0004\u0018\u00010\u0001H\u00d6\u0003J\t\u0010\u0011\u001a\u00020\u0012H\u00d6\u0001J\t\u0010\u0013\u001a\u00020\u0014H\u00d6\u0001R\u0011\u0010\u0002\u001a\u00020\u0003\u00a2\u0006\b\n\u0000\u001a\u0004\b\u0007\u0010\bR\u0011\u0010\u0004\u001a\u00020\u0005\u00a2\u0006\b\n\u0000\u001a\u0004\b\t\u0010\n\u00a8\u0006\u0015"}, d2 = {"Lcom/mobileproxy/core/proxy/Socks5ProxyServer$UdpSession;", "", "clientAddr", "Ljava/net/InetSocketAddress;", "targetSocket", "Ljava/net/DatagramSocket;", "(Ljava/net/InetSocketAddress;Ljava/net/DatagramSocket;)V", "getClientAddr", "()Ljava/net/InetSocketAddress;", "getTargetSocket", "()Ljava/net/DatagramSocket;", "component1", "component2", "copy", "equals", "", "other", "hashCode", "", "toString", "", "app_debug"})
    public static final class UdpSession {
        @org.jetbrains.annotations.NotNull()
        private final java.net.InetSocketAddress clientAddr = null;
        @org.jetbrains.annotations.NotNull()
        private final java.net.DatagramSocket targetSocket = null;
        
        public UdpSession(@org.jetbrains.annotations.NotNull()
        java.net.InetSocketAddress clientAddr, @org.jetbrains.annotations.NotNull()
        java.net.DatagramSocket targetSocket) {
            super();
        }
        
        @org.jetbrains.annotations.NotNull()
        public final java.net.InetSocketAddress getClientAddr() {
            return null;
        }
        
        @org.jetbrains.annotations.NotNull()
        public final java.net.DatagramSocket getTargetSocket() {
            return null;
        }
        
        @org.jetbrains.annotations.NotNull()
        public final java.net.InetSocketAddress component1() {
            return null;
        }
        
        @org.jetbrains.annotations.NotNull()
        public final java.net.DatagramSocket component2() {
            return null;
        }
        
        @org.jetbrains.annotations.NotNull()
        public final com.mobileproxy.core.proxy.Socks5ProxyServer.UdpSession copy(@org.jetbrains.annotations.NotNull()
        java.net.InetSocketAddress clientAddr, @org.jetbrains.annotations.NotNull()
        java.net.DatagramSocket targetSocket) {
            return null;
        }
        
        @java.lang.Override()
        public boolean equals(@org.jetbrains.annotations.Nullable()
        java.lang.Object other) {
            return false;
        }
        
        @java.lang.Override()
        public int hashCode() {
            return 0;
        }
        
        @java.lang.Override()
        @org.jetbrains.annotations.NotNull()
        public java.lang.String toString() {
            return null;
        }
    }
}