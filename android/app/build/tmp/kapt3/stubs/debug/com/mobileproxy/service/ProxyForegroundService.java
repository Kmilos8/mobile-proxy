package com.mobileproxy.service;

@dagger.hilt.android.AndroidEntryPoint()
@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000j\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0002\b\u0005\n\u0002\u0010\u000b\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0005\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0005\n\u0002\u0018\u0002\n\u0002\b\u0005\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\u0002\n\u0000\n\u0002\u0010\b\n\u0002\b\u0004\n\u0002\u0010\u000e\n\u0002\b\u0005\b\u0007\u0018\u0000 42\u00020\u0001:\u00014B\u0005\u00a2\u0006\u0002\u0010\u0002J\b\u0010\"\u001a\u00020#H\u0002J\u0014\u0010$\u001a\u0004\u0018\u00010%2\b\u0010&\u001a\u0004\u0018\u00010\'H\u0016J\b\u0010(\u001a\u00020)H\u0016J\"\u0010*\u001a\u00020+2\b\u0010&\u001a\u0004\u0018\u00010\'2\u0006\u0010,\u001a\u00020+2\u0006\u0010-\u001a\u00020+H\u0016J \u0010.\u001a\u00020)2\u0006\u0010/\u001a\u0002002\u0006\u00101\u001a\u0002002\u0006\u00102\u001a\u000200H\u0002J\b\u00103\u001a\u00020)H\u0002R\u001e\u0010\u0003\u001a\u00020\u00048\u0006@\u0006X\u0087.\u00a2\u0006\u000e\n\u0000\u001a\u0004\b\u0005\u0010\u0006\"\u0004\b\u0007\u0010\bR\u000e\u0010\t\u001a\u00020\nX\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u001e\u0010\u000b\u001a\u00020\f8\u0006@\u0006X\u0087.\u00a2\u0006\u000e\n\u0000\u001a\u0004\b\r\u0010\u000e\"\u0004\b\u000f\u0010\u0010R\u000e\u0010\u0011\u001a\u00020\u0012X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u001e\u0010\u0013\u001a\u00020\u00148\u0006@\u0006X\u0087.\u00a2\u0006\u000e\n\u0000\u001a\u0004\b\u0015\u0010\u0016\"\u0004\b\u0017\u0010\u0018R\u001e\u0010\u0019\u001a\u00020\u001a8\u0006@\u0006X\u0087.\u00a2\u0006\u000e\n\u0000\u001a\u0004\b\u001b\u0010\u001c\"\u0004\b\u001d\u0010\u001eR\u0014\u0010\u001f\u001a\b\u0018\u00010 R\u00020!X\u0082\u000e\u00a2\u0006\u0002\n\u0000\u00a8\u00065"}, d2 = {"Lcom/mobileproxy/service/ProxyForegroundService;", "Landroid/app/Service;", "()V", "httpProxy", "Lcom/mobileproxy/core/proxy/HttpProxyServer;", "getHttpProxy", "()Lcom/mobileproxy/core/proxy/HttpProxyServer;", "setHttpProxy", "(Lcom/mobileproxy/core/proxy/HttpProxyServer;)V", "isRunning", "", "networkManager", "Lcom/mobileproxy/core/network/NetworkManager;", "getNetworkManager", "()Lcom/mobileproxy/core/network/NetworkManager;", "setNetworkManager", "(Lcom/mobileproxy/core/network/NetworkManager;)V", "scope", "Lkotlinx/coroutines/CoroutineScope;", "socks5Proxy", "Lcom/mobileproxy/core/proxy/Socks5ProxyServer;", "getSocks5Proxy", "()Lcom/mobileproxy/core/proxy/Socks5ProxyServer;", "setSocks5Proxy", "(Lcom/mobileproxy/core/proxy/Socks5ProxyServer;)V", "statusReporter", "Lcom/mobileproxy/core/status/DeviceStatusReporter;", "getStatusReporter", "()Lcom/mobileproxy/core/status/DeviceStatusReporter;", "setStatusReporter", "(Lcom/mobileproxy/core/status/DeviceStatusReporter;)V", "wakeLock", "Landroid/os/PowerManager$WakeLock;", "Landroid/os/PowerManager;", "createNotification", "Landroid/app/Notification;", "onBind", "Landroid/os/IBinder;", "intent", "Landroid/content/Intent;", "onDestroy", "", "onStartCommand", "", "flags", "startId", "startProxy", "serverUrl", "", "deviceId", "authToken", "stopProxy", "Companion", "app_debug"})
public final class ProxyForegroundService extends android.app.Service {
    @org.jetbrains.annotations.NotNull()
    private static final java.lang.String TAG = "ProxyForegroundService";
    @org.jetbrains.annotations.NotNull()
    public static final java.lang.String ACTION_START = "com.mobileproxy.START";
    @org.jetbrains.annotations.NotNull()
    public static final java.lang.String ACTION_STOP = "com.mobileproxy.STOP";
    @org.jetbrains.annotations.NotNull()
    public static final java.lang.String EXTRA_SERVER_URL = "server_url";
    @org.jetbrains.annotations.NotNull()
    public static final java.lang.String EXTRA_DEVICE_ID = "device_id";
    @org.jetbrains.annotations.NotNull()
    public static final java.lang.String EXTRA_AUTH_TOKEN = "auth_token";
    @javax.inject.Inject()
    public com.mobileproxy.core.network.NetworkManager networkManager;
    @javax.inject.Inject()
    public com.mobileproxy.core.proxy.HttpProxyServer httpProxy;
    @javax.inject.Inject()
    public com.mobileproxy.core.proxy.Socks5ProxyServer socks5Proxy;
    @javax.inject.Inject()
    public com.mobileproxy.core.status.DeviceStatusReporter statusReporter;
    @org.jetbrains.annotations.Nullable()
    private android.os.PowerManager.WakeLock wakeLock;
    @org.jetbrains.annotations.NotNull()
    private final kotlinx.coroutines.CoroutineScope scope = null;
    @kotlin.jvm.Volatile()
    private volatile boolean isRunning = false;
    @org.jetbrains.annotations.NotNull()
    public static final com.mobileproxy.service.ProxyForegroundService.Companion Companion = null;
    
    public ProxyForegroundService() {
        super();
    }
    
    @org.jetbrains.annotations.NotNull()
    public final com.mobileproxy.core.network.NetworkManager getNetworkManager() {
        return null;
    }
    
    public final void setNetworkManager(@org.jetbrains.annotations.NotNull()
    com.mobileproxy.core.network.NetworkManager p0) {
    }
    
    @org.jetbrains.annotations.NotNull()
    public final com.mobileproxy.core.proxy.HttpProxyServer getHttpProxy() {
        return null;
    }
    
    public final void setHttpProxy(@org.jetbrains.annotations.NotNull()
    com.mobileproxy.core.proxy.HttpProxyServer p0) {
    }
    
    @org.jetbrains.annotations.NotNull()
    public final com.mobileproxy.core.proxy.Socks5ProxyServer getSocks5Proxy() {
        return null;
    }
    
    public final void setSocks5Proxy(@org.jetbrains.annotations.NotNull()
    com.mobileproxy.core.proxy.Socks5ProxyServer p0) {
    }
    
    @org.jetbrains.annotations.NotNull()
    public final com.mobileproxy.core.status.DeviceStatusReporter getStatusReporter() {
        return null;
    }
    
    public final void setStatusReporter(@org.jetbrains.annotations.NotNull()
    com.mobileproxy.core.status.DeviceStatusReporter p0) {
    }
    
    @java.lang.Override()
    @org.jetbrains.annotations.Nullable()
    public android.os.IBinder onBind(@org.jetbrains.annotations.Nullable()
    android.content.Intent intent) {
        return null;
    }
    
    @java.lang.Override()
    public int onStartCommand(@org.jetbrains.annotations.Nullable()
    android.content.Intent intent, int flags, int startId) {
        return 0;
    }
    
    private final void startProxy(java.lang.String serverUrl, java.lang.String deviceId, java.lang.String authToken) {
    }
    
    private final void stopProxy() {
    }
    
    private final android.app.Notification createNotification() {
        return null;
    }
    
    @java.lang.Override()
    public void onDestroy() {
    }
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000\u0014\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\u000e\n\u0002\b\u0006\b\u0086\u0003\u0018\u00002\u00020\u0001B\u0007\b\u0002\u00a2\u0006\u0002\u0010\u0002R\u000e\u0010\u0003\u001a\u00020\u0004X\u0086T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0005\u001a\u00020\u0004X\u0086T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0006\u001a\u00020\u0004X\u0086T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0007\u001a\u00020\u0004X\u0086T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\b\u001a\u00020\u0004X\u0086T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\t\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000\u00a8\u0006\n"}, d2 = {"Lcom/mobileproxy/service/ProxyForegroundService$Companion;", "", "()V", "ACTION_START", "", "ACTION_STOP", "EXTRA_AUTH_TOKEN", "EXTRA_DEVICE_ID", "EXTRA_SERVER_URL", "TAG", "app_debug"})
    public static final class Companion {
        
        private Companion() {
            super();
        }
    }
}