package com.mobileproxy.core.vpn;

@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000Z\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\u000e\n\u0000\n\u0002\u0010\b\n\u0002\b\u0003\n\u0002\u0010\u000b\n\u0002\b\u0003\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0003\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u000b\n\u0002\u0010\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0002\b\u0005\n\u0002\u0010\u0012\n\u0002\b\u0004\u0018\u0000 /2\u00020\u0001:\u0001/B%\u0012\u0006\u0010\u0002\u001a\u00020\u0003\u0012\u0006\u0010\u0004\u001a\u00020\u0005\u0012\u0006\u0010\u0006\u001a\u00020\u0007\u0012\u0006\u0010\b\u001a\u00020\u0005\u00a2\u0006\u0002\u0010\tJ\u0012\u0010\u001b\u001a\u0004\u0018\u00010\u00052\u0006\u0010\u001c\u001a\u00020\u0017H\u0002J\u000e\u0010\u001d\u001a\u00020\u000bH\u0086@\u00a2\u0006\u0002\u0010\u001eJ\u000e\u0010\u001f\u001a\u00020\u000bH\u0082@\u00a2\u0006\u0002\u0010\u001eJ\u0012\u0010 \u001a\u0004\u0018\u00010\u00152\u0006\u0010!\u001a\u00020\u0005H\u0002J\u0006\u0010\"\u001a\u00020#J\u000e\u0010$\u001a\u00020#H\u0082@\u00a2\u0006\u0002\u0010\u001eJ\u000e\u0010%\u001a\u00020\u000b2\u0006\u0010\u001c\u001a\u00020\u0017J\u000e\u0010%\u001a\u00020\u000b2\u0006\u0010\u001c\u001a\u00020&J\u000e\u0010\'\u001a\u00020#H\u0082@\u00a2\u0006\u0002\u0010\u001eJ\b\u0010(\u001a\u00020#H\u0002J\b\u0010)\u001a\u00020#H\u0002J\b\u0010*\u001a\u00020#H\u0002J\u0012\u0010+\u001a\u0004\u0018\u00010,2\u0006\u0010-\u001a\u00020\u0005H\u0002J\u000e\u0010.\u001a\u00020#H\u0082@\u00a2\u0006\u0002\u0010\u001eR\u000e\u0010\b\u001a\u00020\u0005X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u001e\u0010\f\u001a\u00020\u000b2\u0006\u0010\n\u001a\u00020\u000b@BX\u0086\u000e\u00a2\u0006\b\n\u0000\u001a\u0004\b\f\u0010\rR\u000e\u0010\u000e\u001a\u00020\u000fX\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0010\u001a\u00020\u0011X\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0004\u001a\u00020\u0005X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0006\u001a\u00020\u0007X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u001e\u0010\u0012\u001a\u00020\u000b2\u0006\u0010\n\u001a\u00020\u000b@BX\u0086\u000e\u00a2\u0006\b\n\u0000\u001a\u0004\b\u0013\u0010\rR\u0010\u0010\u0014\u001a\u0004\u0018\u00010\u0015X\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u0010\u0010\u0016\u001a\u0004\u0018\u00010\u0017X\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u001e\u0010\u0018\u001a\u00020\u00052\u0006\u0010\n\u001a\u00020\u0005@BX\u0086\u000e\u00a2\u0006\b\n\u0000\u001a\u0004\b\u0019\u0010\u001aR\u000e\u0010\u0002\u001a\u00020\u0003X\u0082\u0004\u00a2\u0006\u0002\n\u0000\u00a8\u00060"}, d2 = {"Lcom/mobileproxy/core/vpn/VpnTunnelManager;", "", "vpnService", "Lcom/mobileproxy/service/ProxyVpnService;", "serverAddress", "", "serverPort", "", "deviceId", "(Lcom/mobileproxy/service/ProxyVpnService;Ljava/lang/String;ILjava/lang/String;)V", "<set-?>", "", "isConnected", "()Z", "lastPongTime", "Ljava/util/concurrent/atomic/AtomicLong;", "scope", "Lkotlinx/coroutines/CoroutineScope;", "shouldRun", "getShouldRun", "tunFd", "Landroid/os/ParcelFileDescriptor;", "udpSocket", "Ljava/net/DatagramSocket;", "vpnIP", "getVpnIP", "()Ljava/lang/String;", "authenticate", "socket", "connect", "(Lkotlin/coroutines/Continuation;)Ljava/lang/Object;", "connectInternal", "createTun", "assignedIP", "disconnect", "", "keepalive", "protectSocket", "Ljava/net/Socket;", "reconnect", "teardownConnection", "tunToUdp", "udpToTun", "uuidToBytes", "", "uuid", "watchdog", "Companion", "app_debug"})
public final class VpnTunnelManager {
    @org.jetbrains.annotations.NotNull()
    private final com.mobileproxy.service.ProxyVpnService vpnService = null;
    @org.jetbrains.annotations.NotNull()
    private final java.lang.String serverAddress = null;
    private final int serverPort = 0;
    @org.jetbrains.annotations.NotNull()
    private final java.lang.String deviceId = null;
    @org.jetbrains.annotations.NotNull()
    private static final java.lang.String TAG = "VpnTunnelManager";
    public static final int MTU = 1400;
    public static final long KEEPALIVE_MS = 25000L;
    private static final int RECV_BUF_SIZE = 2097152;
    private static final int SEND_BUF_SIZE = 2097152;
    private static final long RECONNECT_DELAY_MS = 3000L;
    private static final long MAX_RECONNECT_DELAY_MS = 30000L;
    private static final long PONG_TIMEOUT_MS = 45000L;
    private static final byte TYPE_AUTH = (byte)1;
    private static final byte TYPE_DATA = (byte)2;
    private static final byte TYPE_PING = (byte)3;
    private static final byte TYPE_AUTH_OK = (byte)1;
    private static final byte TYPE_AUTH_FAIL = (byte)3;
    private static final byte TYPE_PONG = (byte)4;
    @org.jetbrains.annotations.Nullable()
    private android.os.ParcelFileDescriptor tunFd;
    @org.jetbrains.annotations.Nullable()
    private java.net.DatagramSocket udpSocket;
    @org.jetbrains.annotations.NotNull()
    private kotlinx.coroutines.CoroutineScope scope;
    @org.jetbrains.annotations.NotNull()
    private final java.util.concurrent.atomic.AtomicLong lastPongTime = null;
    @kotlin.jvm.Volatile()
    @org.jetbrains.annotations.NotNull()
    private volatile java.lang.String vpnIP = "";
    @kotlin.jvm.Volatile()
    private volatile boolean isConnected = false;
    @kotlin.jvm.Volatile()
    private volatile boolean shouldRun = false;
    @org.jetbrains.annotations.NotNull()
    public static final com.mobileproxy.core.vpn.VpnTunnelManager.Companion Companion = null;
    
    public VpnTunnelManager(@org.jetbrains.annotations.NotNull()
    com.mobileproxy.service.ProxyVpnService vpnService, @org.jetbrains.annotations.NotNull()
    java.lang.String serverAddress, int serverPort, @org.jetbrains.annotations.NotNull()
    java.lang.String deviceId) {
        super();
    }
    
    @org.jetbrains.annotations.NotNull()
    public final java.lang.String getVpnIP() {
        return null;
    }
    
    public final boolean isConnected() {
        return false;
    }
    
    public final boolean getShouldRun() {
        return false;
    }
    
    @org.jetbrains.annotations.Nullable()
    public final java.lang.Object connect(@org.jetbrains.annotations.NotNull()
    kotlin.coroutines.Continuation<? super java.lang.Boolean> $completion) {
        return null;
    }
    
    private final java.lang.Object connectInternal(kotlin.coroutines.Continuation<? super java.lang.Boolean> $completion) {
        return null;
    }
    
    public final void disconnect() {
    }
    
    private final void teardownConnection() {
    }
    
    private final java.lang.Object reconnect(kotlin.coroutines.Continuation<? super kotlin.Unit> $completion) {
        return null;
    }
    
    public final boolean protectSocket(@org.jetbrains.annotations.NotNull()
    java.net.Socket socket) {
        return false;
    }
    
    public final boolean protectSocket(@org.jetbrains.annotations.NotNull()
    java.net.DatagramSocket socket) {
        return false;
    }
    
    private final android.os.ParcelFileDescriptor createTun(java.lang.String assignedIP) {
        return null;
    }
    
    private final java.lang.String authenticate(java.net.DatagramSocket socket) {
        return null;
    }
    
    private final void tunToUdp() {
    }
    
    private final void udpToTun() {
    }
    
    private final java.lang.Object keepalive(kotlin.coroutines.Continuation<? super kotlin.Unit> $completion) {
        return null;
    }
    
    /**
     * Watchdog: detects dead tunnels by monitoring PONG responses.
     * If no PONG received within PONG_TIMEOUT_MS, triggers reconnect.
     */
    private final java.lang.Object watchdog(kotlin.coroutines.Continuation<? super kotlin.Unit> $completion) {
        return null;
    }
    
    private final byte[] uuidToBytes(java.lang.String uuid) {
        return null;
    }
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000*\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\t\n\u0002\b\u0002\n\u0002\u0010\b\n\u0002\b\u0005\n\u0002\u0010\u000e\n\u0000\n\u0002\u0010\u0005\n\u0002\b\u0006\b\u0086\u0003\u0018\u00002\u00020\u0001B\u0007\b\u0002\u00a2\u0006\u0002\u0010\u0002R\u000e\u0010\u0003\u001a\u00020\u0004X\u0086T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0005\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0006\u001a\u00020\u0007X\u0086T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\b\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\t\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\n\u001a\u00020\u0007X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u000b\u001a\u00020\u0007X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\f\u001a\u00020\rX\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u000e\u001a\u00020\u000fX\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0010\u001a\u00020\u000fX\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0011\u001a\u00020\u000fX\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0012\u001a\u00020\u000fX\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0013\u001a\u00020\u000fX\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0014\u001a\u00020\u000fX\u0082T\u00a2\u0006\u0002\n\u0000\u00a8\u0006\u0015"}, d2 = {"Lcom/mobileproxy/core/vpn/VpnTunnelManager$Companion;", "", "()V", "KEEPALIVE_MS", "", "MAX_RECONNECT_DELAY_MS", "MTU", "", "PONG_TIMEOUT_MS", "RECONNECT_DELAY_MS", "RECV_BUF_SIZE", "SEND_BUF_SIZE", "TAG", "", "TYPE_AUTH", "", "TYPE_AUTH_FAIL", "TYPE_AUTH_OK", "TYPE_DATA", "TYPE_PING", "TYPE_PONG", "app_debug"})
    public static final class Companion {
        
        private Companion() {
            super();
        }
    }
}