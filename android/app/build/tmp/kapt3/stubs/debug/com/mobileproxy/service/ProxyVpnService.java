package com.mobileproxy.service;

@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u00006\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\u0002\n\u0002\b\u0002\n\u0002\u0010\b\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0004\n\u0002\u0010\u000e\n\u0002\b\u0004\u0018\u0000 \u00152\u00020\u0001:\u0001\u0015B\u0005\u00a2\u0006\u0002\u0010\u0002J\b\u0010\u0007\u001a\u00020\bH\u0016J\b\u0010\t\u001a\u00020\bH\u0016J\"\u0010\n\u001a\u00020\u000b2\b\u0010\f\u001a\u0004\u0018\u00010\r2\u0006\u0010\u000e\u001a\u00020\u000b2\u0006\u0010\u000f\u001a\u00020\u000bH\u0016J\u0018\u0010\u0010\u001a\u00020\b2\u0006\u0010\u0011\u001a\u00020\u00122\u0006\u0010\u0013\u001a\u00020\u0012H\u0002J\b\u0010\u0014\u001a\u00020\bH\u0002R\u000e\u0010\u0003\u001a\u00020\u0004X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u0010\u0010\u0005\u001a\u0004\u0018\u00010\u0006X\u0082\u000e\u00a2\u0006\u0002\n\u0000\u00a8\u0006\u0016"}, d2 = {"Lcom/mobileproxy/service/ProxyVpnService;", "Landroid/net/VpnService;", "()V", "scope", "Lkotlinx/coroutines/CoroutineScope;", "tunnelManager", "Lcom/mobileproxy/core/vpn/VpnTunnelManager;", "onDestroy", "", "onRevoke", "onStartCommand", "", "intent", "Landroid/content/Intent;", "flags", "startId", "startVpn", "serverIP", "", "deviceId", "stopVpn", "Companion", "app_debug"})
public final class ProxyVpnService extends android.net.VpnService {
    @org.jetbrains.annotations.NotNull()
    private static final java.lang.String TAG = "ProxyVpnService";
    @org.jetbrains.annotations.NotNull()
    public static final java.lang.String ACTION_START = "com.mobileproxy.VPN_START";
    @org.jetbrains.annotations.NotNull()
    public static final java.lang.String ACTION_STOP = "com.mobileproxy.VPN_STOP";
    @org.jetbrains.annotations.NotNull()
    public static final java.lang.String EXTRA_SERVER_IP = "server_ip";
    @org.jetbrains.annotations.NotNull()
    public static final java.lang.String EXTRA_DEVICE_ID = "device_id";
    @kotlin.jvm.Volatile()
    @org.jetbrains.annotations.Nullable()
    private static volatile com.mobileproxy.service.ProxyVpnService instance;
    @org.jetbrains.annotations.NotNull()
    private static final kotlinx.coroutines.flow.MutableStateFlow<java.lang.Boolean> _vpnState = null;
    @org.jetbrains.annotations.NotNull()
    private static final kotlinx.coroutines.flow.StateFlow<java.lang.Boolean> vpnState = null;
    @org.jetbrains.annotations.Nullable()
    private com.mobileproxy.core.vpn.VpnTunnelManager tunnelManager;
    @org.jetbrains.annotations.NotNull()
    private final kotlinx.coroutines.CoroutineScope scope = null;
    @org.jetbrains.annotations.NotNull()
    public static final com.mobileproxy.service.ProxyVpnService.Companion Companion = null;
    
    public ProxyVpnService() {
        super();
    }
    
    @java.lang.Override()
    public int onStartCommand(@org.jetbrains.annotations.Nullable()
    android.content.Intent intent, int flags, int startId) {
        return 0;
    }
    
    private final void startVpn(java.lang.String serverIP, java.lang.String deviceId) {
    }
    
    private final void stopVpn() {
    }
    
    @java.lang.Override()
    public void onDestroy() {
    }
    
    @java.lang.Override()
    public void onRevoke() {
    }
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000@\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\u000e\n\u0002\b\u0005\n\u0002\u0018\u0002\n\u0002\u0010\u000b\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0004\n\u0002\u0018\u0002\n\u0002\b\u0003\n\u0002\u0010\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0000\b\u0086\u0003\u0018\u00002\u00020\u0001B\u0007\b\u0002\u00a2\u0006\u0002\u0010\u0002J\u0006\u0010\u0015\u001a\u00020\u0016J\u000e\u0010\u0017\u001a\u00020\u000b2\u0006\u0010\u0018\u001a\u00020\u0019J\u000e\u0010\u0017\u001a\u00020\u000b2\u0006\u0010\u0018\u001a\u00020\u001aR\u000e\u0010\u0003\u001a\u00020\u0004X\u0086T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0005\u001a\u00020\u0004X\u0086T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0006\u001a\u00020\u0004X\u0086T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0007\u001a\u00020\u0004X\u0086T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\b\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u0014\u0010\t\u001a\b\u0012\u0004\u0012\u00020\u000b0\nX\u0082\u0004\u00a2\u0006\u0002\n\u0000R\"\u0010\u000e\u001a\u0004\u0018\u00010\r2\b\u0010\f\u001a\u0004\u0018\u00010\r@BX\u0086\u000e\u00a2\u0006\b\n\u0000\u001a\u0004\b\u000f\u0010\u0010R\u0017\u0010\u0011\u001a\b\u0012\u0004\u0012\u00020\u000b0\u0012\u00a2\u0006\b\n\u0000\u001a\u0004\b\u0013\u0010\u0014\u00a8\u0006\u001b"}, d2 = {"Lcom/mobileproxy/service/ProxyVpnService$Companion;", "", "()V", "ACTION_START", "", "ACTION_STOP", "EXTRA_DEVICE_ID", "EXTRA_SERVER_IP", "TAG", "_vpnState", "Lkotlinx/coroutines/flow/MutableStateFlow;", "", "<set-?>", "Lcom/mobileproxy/service/ProxyVpnService;", "instance", "getInstance", "()Lcom/mobileproxy/service/ProxyVpnService;", "vpnState", "Lkotlinx/coroutines/flow/StateFlow;", "getVpnState", "()Lkotlinx/coroutines/flow/StateFlow;", "onReconnected", "", "protectSocket", "socket", "Ljava/net/DatagramSocket;", "Ljava/net/Socket;", "app_debug"})
    public static final class Companion {
        
        private Companion() {
            super();
        }
        
        @org.jetbrains.annotations.Nullable()
        public final com.mobileproxy.service.ProxyVpnService getInstance() {
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
        
        @org.jetbrains.annotations.NotNull()
        public final kotlinx.coroutines.flow.StateFlow<java.lang.Boolean> getVpnState() {
            return null;
        }
        
        /**
         * Called by VpnTunnelManager after a successful reconnect.
         * Updates the VPN state flow so observers know we're back online.
         */
        public final void onReconnected() {
        }
    }
}