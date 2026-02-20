package com.mobileproxy.core.status;

@javax.inject.Singleton()
@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000`\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0010\t\n\u0000\n\u0002\u0010\u000e\n\u0002\b\u0007\n\u0002\u0018\u0002\n\u0002\b\u0006\n\u0002\u0018\u0002\n\u0002\b\u0004\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0010\u0002\n\u0002\b\u0003\n\u0002\u0018\u0002\n\u0002\b\u0007\b\u0007\u0018\u0000 12\u00020\u0001:\u00011B1\b\u0007\u0012\b\b\u0001\u0010\u0002\u001a\u00020\u0003\u0012\u0006\u0010\u0004\u001a\u00020\u0005\u0012\u0006\u0010\u0006\u001a\u00020\u0007\u0012\u0006\u0010\b\u001a\u00020\t\u0012\u0006\u0010\n\u001a\u00020\u000b\u00a2\u0006\u0002\u0010\fJ\b\u0010!\u001a\u00020\u0010H\u0002J\u0010\u0010\"\u001a\u00020\u00102\u0006\u0010#\u001a\u00020$H\u0002J\b\u0010%\u001a\u00020\u0010H\u0002J\u0006\u0010&\u001a\u00020\'J$\u0010(\u001a\u00020\'2\u0006\u0010)\u001a\u00020\u00102\f\u0010*\u001a\b\u0012\u0004\u0012\u00020\u00100+H\u0082@\u00a2\u0006\u0002\u0010,J\u000e\u0010-\u001a\u00020\'H\u0082@\u00a2\u0006\u0002\u0010.J\u001e\u0010/\u001a\u00020\'2\u0006\u0010 \u001a\u00020\u00102\u0006\u0010\u0016\u001a\u00020\u00102\u0006\u0010\u000f\u001a\u00020\u0010J\u0006\u00100\u001a\u00020\'R\u000e\u0010\r\u001a\u00020\u000eX\u0082D\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u000f\u001a\u00020\u0010X\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u001a\u0010\u0011\u001a\u00020\u0010X\u0086\u000e\u00a2\u0006\u000e\n\u0000\u001a\u0004\b\u0012\u0010\u0013\"\u0004\b\u0014\u0010\u0015R\u000e\u0010\n\u001a\u00020\u000bX\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0002\u001a\u00020\u0003X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0016\u001a\u00020\u0010X\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0017\u001a\u00020\u0018X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0006\u001a\u00020\u0007X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u001a\u0010\u0019\u001a\u00020\u000eX\u0086\u000e\u00a2\u0006\u000e\n\u0000\u001a\u0004\b\u001a\u0010\u001b\"\u0004\b\u001c\u0010\u001dR\u000e\u0010\u0004\u001a\u00020\u0005X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u0010\u0010\u001e\u001a\u0004\u0018\u00010\u001fX\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u000e\u0010 \u001a\u00020\u0010X\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u000e\u0010\b\u001a\u00020\tX\u0082\u0004\u00a2\u0006\u0002\n\u0000\u00a8\u00062"}, d2 = {"Lcom/mobileproxy/core/status/DeviceStatusReporter;", "", "context", "Landroid/content/Context;", "networkManager", "Lcom/mobileproxy/core/network/NetworkManager;", "httpProxy", "Lcom/mobileproxy/core/proxy/HttpProxyServer;", "socks5Proxy", "Lcom/mobileproxy/core/proxy/Socks5ProxyServer;", "commandExecutor", "Lcom/mobileproxy/core/commands/CommandExecutor;", "(Landroid/content/Context;Lcom/mobileproxy/core/network/NetworkManager;Lcom/mobileproxy/core/proxy/HttpProxyServer;Lcom/mobileproxy/core/proxy/Socks5ProxyServer;Lcom/mobileproxy/core/commands/CommandExecutor;)V", "IP_CACHE_DURATION", "", "authToken", "", "cachedPublicIp", "getCachedPublicIp", "()Ljava/lang/String;", "setCachedPublicIp", "(Ljava/lang/String;)V", "deviceId", "gson", "Lcom/google/gson/Gson;", "lastIpLookupTime", "getLastIpLookupTime", "()J", "setLastIpLookupTime", "(J)V", "scope", "Lkotlinx/coroutines/CoroutineScope;", "serverUrl", "getCellularIp", "getNetworkTypeString", "tm", "Landroid/telephony/TelephonyManager;", "getWifiIp", "invalidateIpCache", "", "reportCommandResult", "commandId", "result", "Lkotlin/Result;", "(Ljava/lang/String;Ljava/lang/Object;Lkotlin/coroutines/Continuation;)Ljava/lang/Object;", "sendHeartbeat", "(Lkotlin/coroutines/Continuation;)Ljava/lang/Object;", "start", "stop", "Companion", "app_debug"})
public final class DeviceStatusReporter {
    @org.jetbrains.annotations.NotNull()
    private final android.content.Context context = null;
    @org.jetbrains.annotations.NotNull()
    private final com.mobileproxy.core.network.NetworkManager networkManager = null;
    @org.jetbrains.annotations.NotNull()
    private final com.mobileproxy.core.proxy.HttpProxyServer httpProxy = null;
    @org.jetbrains.annotations.NotNull()
    private final com.mobileproxy.core.proxy.Socks5ProxyServer socks5Proxy = null;
    @org.jetbrains.annotations.NotNull()
    private final com.mobileproxy.core.commands.CommandExecutor commandExecutor = null;
    @org.jetbrains.annotations.NotNull()
    private static final java.lang.String TAG = "StatusReporter";
    private static final long HEARTBEAT_INTERVAL = 30000L;
    @org.jetbrains.annotations.NotNull()
    private final com.google.gson.Gson gson = null;
    @org.jetbrains.annotations.Nullable()
    private kotlinx.coroutines.CoroutineScope scope;
    @org.jetbrains.annotations.NotNull()
    private java.lang.String serverUrl = "";
    @org.jetbrains.annotations.NotNull()
    private java.lang.String deviceId = "";
    @org.jetbrains.annotations.NotNull()
    private java.lang.String authToken = "";
    @kotlin.jvm.Volatile()
    @org.jetbrains.annotations.NotNull()
    private volatile java.lang.String cachedPublicIp = "";
    @kotlin.jvm.Volatile()
    private volatile long lastIpLookupTime = 0L;
    private final long IP_CACHE_DURATION = 60000L;
    @org.jetbrains.annotations.NotNull()
    public static final com.mobileproxy.core.status.DeviceStatusReporter.Companion Companion = null;
    
    @javax.inject.Inject()
    public DeviceStatusReporter(@dagger.hilt.android.qualifiers.ApplicationContext()
    @org.jetbrains.annotations.NotNull()
    android.content.Context context, @org.jetbrains.annotations.NotNull()
    com.mobileproxy.core.network.NetworkManager networkManager, @org.jetbrains.annotations.NotNull()
    com.mobileproxy.core.proxy.HttpProxyServer httpProxy, @org.jetbrains.annotations.NotNull()
    com.mobileproxy.core.proxy.Socks5ProxyServer socks5Proxy, @org.jetbrains.annotations.NotNull()
    com.mobileproxy.core.commands.CommandExecutor commandExecutor) {
        super();
    }
    
    public final void start(@org.jetbrains.annotations.NotNull()
    java.lang.String serverUrl, @org.jetbrains.annotations.NotNull()
    java.lang.String deviceId, @org.jetbrains.annotations.NotNull()
    java.lang.String authToken) {
    }
    
    public final void stop() {
    }
    
    private final java.lang.Object sendHeartbeat(kotlin.coroutines.Continuation<? super kotlin.Unit> $completion) {
        return null;
    }
    
    private final java.lang.Object reportCommandResult(java.lang.String commandId, java.lang.Object result, kotlin.coroutines.Continuation<? super kotlin.Unit> $completion) {
        return null;
    }
    
    @org.jetbrains.annotations.NotNull()
    public final java.lang.String getCachedPublicIp() {
        return null;
    }
    
    public final void setCachedPublicIp(@org.jetbrains.annotations.NotNull()
    java.lang.String p0) {
    }
    
    public final long getLastIpLookupTime() {
        return 0L;
    }
    
    public final void setLastIpLookupTime(long p0) {
    }
    
    public final void invalidateIpCache() {
    }
    
    private final java.lang.String getCellularIp() {
        return null;
    }
    
    private final java.lang.String getWifiIp() {
        return null;
    }
    
    private final java.lang.String getNetworkTypeString(android.telephony.TelephonyManager tm) {
        return null;
    }
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000\u0018\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\t\n\u0000\n\u0002\u0010\u000e\n\u0000\b\u0086\u0003\u0018\u00002\u00020\u0001B\u0007\b\u0002\u00a2\u0006\u0002\u0010\u0002R\u000e\u0010\u0003\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0005\u001a\u00020\u0006X\u0082T\u00a2\u0006\u0002\n\u0000\u00a8\u0006\u0007"}, d2 = {"Lcom/mobileproxy/core/status/DeviceStatusReporter$Companion;", "", "()V", "HEARTBEAT_INTERVAL", "", "TAG", "", "app_debug"})
    public static final class Companion {
        
        private Companion() {
            super();
        }
    }
}