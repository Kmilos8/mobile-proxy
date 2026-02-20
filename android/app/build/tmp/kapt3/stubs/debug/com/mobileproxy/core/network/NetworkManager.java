package com.mobileproxy.core.network;

@javax.inject.Singleton()
@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000n\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0003\n\u0002\u0018\u0002\n\u0002\b\u0005\n\u0002\u0010\u0002\n\u0002\b\u0003\n\u0002\u0010\u000b\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\b\n\u0002\b\u0003\n\u0002\u0018\u0002\n\u0002\b\u0006\n\u0002\u0010\u000e\n\u0002\b\u0003\b\u0007\u0018\u0000 02\u00020\u0001:\u00010B\u0011\b\u0007\u0012\b\b\u0001\u0010\u0002\u001a\u00020\u0003\u00a2\u0006\u0002\u0010\u0004J\b\u0010\u0017\u001a\u00020\u0018H\u0002J\u0006\u0010\u0019\u001a\u00020\u0018J\b\u0010\u001a\u001a\u00020\u0018H\u0002J\u000e\u0010\u001b\u001a\u00020\u001c2\u0006\u0010\u001d\u001a\u00020\u001eJ\u0016\u0010\u001f\u001a\u00020\u001e2\u0006\u0010 \u001a\u00020!2\u0006\u0010\"\u001a\u00020#J\n\u0010$\u001a\u0004\u0018\u00010\fH\u0002J\b\u0010%\u001a\u0004\u0018\u00010\fJ\b\u0010&\u001a\u0004\u0018\u00010\'J\n\u0010(\u001a\u0004\u0018\u00010\fH\u0002J\b\u0010)\u001a\u0004\u0018\u00010\fJ\u0006\u0010*\u001a\u00020\u0018J\u0006\u0010+\u001a\u00020\u0018J\u0016\u0010,\u001a\u00020!2\u0006\u0010-\u001a\u00020.H\u0086@\u00a2\u0006\u0002\u0010/R\u0014\u0010\u0005\u001a\b\u0012\u0004\u0012\u00020\u00070\u0006X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u0014\u0010\b\u001a\b\u0012\u0004\u0012\u00020\u00070\u0006X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u0010\u0010\t\u001a\u0004\u0018\u00010\nX\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u0010\u0010\u000b\u001a\u0004\u0018\u00010\fX\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u0017\u0010\r\u001a\b\u0012\u0004\u0012\u00020\u00070\u000e\u00a2\u0006\b\n\u0000\u001a\u0004\b\u000f\u0010\u0010R\u000e\u0010\u0011\u001a\u00020\u0012X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0002\u001a\u00020\u0003X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u0010\u0010\u0013\u001a\u0004\u0018\u00010\nX\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u0010\u0010\u0014\u001a\u0004\u0018\u00010\fX\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u0017\u0010\u0015\u001a\b\u0012\u0004\u0012\u00020\u00070\u000e\u00a2\u0006\b\n\u0000\u001a\u0004\b\u0016\u0010\u0010\u00a8\u00061"}, d2 = {"Lcom/mobileproxy/core/network/NetworkManager;", "", "context", "Landroid/content/Context;", "(Landroid/content/Context;)V", "_cellularState", "Lkotlinx/coroutines/flow/MutableStateFlow;", "Lcom/mobileproxy/core/network/NetworkState;", "_wifiState", "cellularCallback", "Landroid/net/ConnectivityManager$NetworkCallback;", "cellularNetwork", "Landroid/net/Network;", "cellularState", "Lkotlinx/coroutines/flow/StateFlow;", "getCellularState", "()Lkotlinx/coroutines/flow/StateFlow;", "connectivityManager", "Landroid/net/ConnectivityManager;", "wifiCallback", "wifiNetwork", "wifiState", "getWifiState", "acquireCellular", "", "acquireNetworks", "acquireWifi", "bindSocketToCellular", "", "socket", "Ljava/net/Socket;", "createCellularSocket", "address", "Ljava/net/InetAddress;", "port", "", "findCellularNetwork", "getCellularNetwork", "getCellularSocketFactory", "Ljavax/net/SocketFactory;", "getOrFindCellular", "getWifiNetwork", "reconnectCellular", "releaseNetworks", "resolveDnsCellular", "hostname", "", "(Ljava/lang/String;Lkotlin/coroutines/Continuation;)Ljava/lang/Object;", "Companion", "app_debug"})
public final class NetworkManager {
    @org.jetbrains.annotations.NotNull()
    private final android.content.Context context = null;
    @org.jetbrains.annotations.NotNull()
    private static final java.lang.String TAG = "NetworkManager";
    @org.jetbrains.annotations.NotNull()
    private final android.net.ConnectivityManager connectivityManager = null;
    @kotlin.jvm.Volatile()
    @org.jetbrains.annotations.Nullable()
    private volatile android.net.Network cellularNetwork;
    @kotlin.jvm.Volatile()
    @org.jetbrains.annotations.Nullable()
    private volatile android.net.Network wifiNetwork;
    @org.jetbrains.annotations.NotNull()
    private final kotlinx.coroutines.flow.MutableStateFlow<com.mobileproxy.core.network.NetworkState> _cellularState = null;
    @org.jetbrains.annotations.NotNull()
    private final kotlinx.coroutines.flow.StateFlow<com.mobileproxy.core.network.NetworkState> cellularState = null;
    @org.jetbrains.annotations.NotNull()
    private final kotlinx.coroutines.flow.MutableStateFlow<com.mobileproxy.core.network.NetworkState> _wifiState = null;
    @org.jetbrains.annotations.NotNull()
    private final kotlinx.coroutines.flow.StateFlow<com.mobileproxy.core.network.NetworkState> wifiState = null;
    @org.jetbrains.annotations.Nullable()
    private android.net.ConnectivityManager.NetworkCallback cellularCallback;
    @org.jetbrains.annotations.Nullable()
    private android.net.ConnectivityManager.NetworkCallback wifiCallback;
    @org.jetbrains.annotations.NotNull()
    public static final com.mobileproxy.core.network.NetworkManager.Companion Companion = null;
    
    @javax.inject.Inject()
    public NetworkManager(@dagger.hilt.android.qualifiers.ApplicationContext()
    @org.jetbrains.annotations.NotNull()
    android.content.Context context) {
        super();
    }
    
    @org.jetbrains.annotations.NotNull()
    public final kotlinx.coroutines.flow.StateFlow<com.mobileproxy.core.network.NetworkState> getCellularState() {
        return null;
    }
    
    @org.jetbrains.annotations.NotNull()
    public final kotlinx.coroutines.flow.StateFlow<com.mobileproxy.core.network.NetworkState> getWifiState() {
        return null;
    }
    
    /**
     * Request both cellular and WiFi networks simultaneously.
     * This is the core of the WiFi Split mechanism.
     */
    public final void acquireNetworks() {
    }
    
    public final void releaseNetworks() {
    }
    
    private final void acquireCellular() {
    }
    
    private final void acquireWifi() {
    }
    
    /**
     * Get a SocketFactory bound to the cellular network.
     * All sockets created by this factory will route through cellular (mobile IP).
     */
    @org.jetbrains.annotations.Nullable()
    public final javax.net.SocketFactory getCellularSocketFactory() {
        return null;
    }
    
    /**
     * Get the cellular Network object, using scan fallback if callback hasn't fired.
     */
    @org.jetbrains.annotations.Nullable()
    public final android.net.Network getCellularNetwork() {
        return null;
    }
    
    /**
     * Get the WiFi Network object (used for VPN tunnel).
     */
    @org.jetbrains.annotations.Nullable()
    public final android.net.Network getWifiNetwork() {
        return null;
    }
    
    /**
     * Find cellular network by scanning all available networks.
     * Fallback when requestNetwork callback hasn't fired.
     */
    private final android.net.Network findCellularNetwork() {
        return null;
    }
    
    /**
     * Get the cellular network, using scan fallback if callback hasn't fired.
     * Returns null if cellular is genuinely unavailable.
     */
    private final android.net.Network getOrFindCellular() {
        return null;
    }
    
    /**
     * Create a socket connected through the cellular network using SocketFactory.
     * Falls back to default network if cellular is unavailable.
     */
    @org.jetbrains.annotations.NotNull()
    public final java.net.Socket createCellularSocket(@org.jetbrains.annotations.NotNull()
    java.net.InetAddress address, int port) {
        return null;
    }
    
    /**
     * Bind a socket to the cellular network explicitly.
     * Returns true if bound to cellular, false if falling back to default.
     */
    public final boolean bindSocketToCellular(@org.jetbrains.annotations.NotNull()
    java.net.Socket socket) {
        return false;
    }
    
    /**
     * Resolve DNS, preferring cellular network. Falls back to system DNS.
     */
    @org.jetbrains.annotations.Nullable()
    public final java.lang.Object resolveDnsCellular(@org.jetbrains.annotations.NotNull()
    java.lang.String hostname, @org.jetbrains.annotations.NotNull()
    kotlin.coroutines.Continuation<? super java.net.InetAddress> $completion) {
        return null;
    }
    
    /**
     * Disconnect and reconnect cellular to trigger IP change.
     * This is the fallback IP rotation method.
     */
    public final void reconnectCellular() {
    }
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000\u0012\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\u000e\n\u0000\b\u0086\u0003\u0018\u00002\u00020\u0001B\u0007\b\u0002\u00a2\u0006\u0002\u0010\u0002R\u000e\u0010\u0003\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000\u00a8\u0006\u0005"}, d2 = {"Lcom/mobileproxy/core/network/NetworkManager$Companion;", "", "()V", "TAG", "", "app_debug"})
    public static final class Companion {
        
        private Companion() {
            super();
        }
    }
}