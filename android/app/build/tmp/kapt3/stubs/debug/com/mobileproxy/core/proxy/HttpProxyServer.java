package com.mobileproxy.core.proxy;

/**
 * HTTP CONNECT proxy server.
 * Accepts connections on the VPN interface and forwards them through cellular.
 */
@javax.inject.Singleton()
@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000V\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0010\t\n\u0002\b\u0005\n\u0002\u0010\u000b\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\u000e\n\u0000\n\u0002\u0010\b\n\u0000\n\u0002\u0010\u0002\n\u0002\b\u0005\n\u0002\u0018\u0002\n\u0002\b\n\b\u0007\u0018\u0000 *2\u00020\u0001:\u0001*B\u000f\b\u0007\u0012\u0006\u0010\u0002\u001a\u00020\u0003\u00a2\u0006\u0002\u0010\u0004J\u0018\u0010\u0014\u001a\u00020\u00152\u0006\u0010\u0016\u001a\u00020\u00172\u0006\u0010\u0018\u001a\u00020\u0019H\u0002J\u0016\u0010\u001a\u001a\u00020\u001b2\u0006\u0010\u001c\u001a\u00020\u0015H\u0082@\u00a2\u0006\u0002\u0010\u001dJ&\u0010\u001e\u001a\u00020\u001b2\u0006\u0010\u001c\u001a\u00020\u00152\u0006\u0010\u001f\u001a\u00020\u00172\u0006\u0010 \u001a\u00020!H\u0082@\u00a2\u0006\u0002\u0010\"J&\u0010#\u001a\u00020\u001b2\u0006\u0010\u001c\u001a\u00020\u00152\u0006\u0010\u001f\u001a\u00020\u00172\u0006\u0010 \u001a\u00020!H\u0082@\u00a2\u0006\u0002\u0010\"J\u001e\u0010$\u001a\u00020\u001b2\u0006\u0010%\u001a\u00020\u00152\u0006\u0010&\u001a\u00020\u0015H\u0082@\u00a2\u0006\u0002\u0010\'J\u0010\u0010(\u001a\u00020\u001b2\b\b\u0002\u0010\u0018\u001a\u00020\u0019J\u0006\u0010)\u001a\u00020\u001bR\u000e\u0010\u0005\u001a\u00020\u0006X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0007\u001a\u00020\u0006X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u0011\u0010\b\u001a\u00020\t8F\u00a2\u0006\u0006\u001a\u0004\b\n\u0010\u000bR\u0011\u0010\f\u001a\u00020\t8F\u00a2\u0006\u0006\u001a\u0004\b\r\u0010\u000bR\u000e\u0010\u0002\u001a\u00020\u0003X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u000e\u001a\u00020\u000fX\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0010\u001a\u00020\u0011X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u0010\u0010\u0012\u001a\u0004\u0018\u00010\u0013X\u0082\u000e\u00a2\u0006\u0002\n\u0000\u00a8\u0006+"}, d2 = {"Lcom/mobileproxy/core/proxy/HttpProxyServer;", "", "networkManager", "Lcom/mobileproxy/core/network/NetworkManager;", "(Lcom/mobileproxy/core/network/NetworkManager;)V", "_bytesIn", "Ljava/util/concurrent/atomic/AtomicLong;", "_bytesOut", "bytesIn", "", "getBytesIn", "()J", "bytesOut", "getBytesOut", "running", "", "scope", "Lkotlinx/coroutines/CoroutineScope;", "serverSocket", "Ljava/net/ServerSocket;", "connectToTarget", "Ljava/net/Socket;", "host", "", "port", "", "handleClient", "", "clientSocket", "(Ljava/net/Socket;Lkotlin/coroutines/Continuation;)Ljava/lang/Object;", "handleConnect", "requestLine", "reader", "Ljava/io/BufferedReader;", "(Ljava/net/Socket;Ljava/lang/String;Ljava/io/BufferedReader;Lkotlin/coroutines/Continuation;)Ljava/lang/Object;", "handlePlainHttp", "relay", "client", "target", "(Ljava/net/Socket;Ljava/net/Socket;Lkotlin/coroutines/Continuation;)Ljava/lang/Object;", "start", "stop", "Companion", "app_debug"})
public final class HttpProxyServer {
    @org.jetbrains.annotations.NotNull()
    private final com.mobileproxy.core.network.NetworkManager networkManager = null;
    @org.jetbrains.annotations.NotNull()
    private static final java.lang.String TAG = "HttpProxyServer";
    private static final int BUFFER_SIZE = 32768;
    @org.jetbrains.annotations.Nullable()
    private java.net.ServerSocket serverSocket;
    private boolean running = false;
    @org.jetbrains.annotations.NotNull()
    private final kotlinx.coroutines.CoroutineScope scope = null;
    @org.jetbrains.annotations.NotNull()
    private final java.util.concurrent.atomic.AtomicLong _bytesIn = null;
    @org.jetbrains.annotations.NotNull()
    private final java.util.concurrent.atomic.AtomicLong _bytesOut = null;
    @org.jetbrains.annotations.NotNull()
    public static final com.mobileproxy.core.proxy.HttpProxyServer.Companion Companion = null;
    
    @javax.inject.Inject()
    public HttpProxyServer(@org.jetbrains.annotations.NotNull()
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
    
    /**
     * Connect to target host via cellular network.
     * Socket is protected from VPN routing and bound to cellular.
     */
    private final java.net.Socket connectToTarget(java.lang.String host, int port) {
        return null;
    }
    
    private final java.lang.Object handleConnect(java.net.Socket clientSocket, java.lang.String requestLine, java.io.BufferedReader reader, kotlin.coroutines.Continuation<? super kotlin.Unit> $completion) {
        return null;
    }
    
    private final java.lang.Object handlePlainHttp(java.net.Socket clientSocket, java.lang.String requestLine, java.io.BufferedReader reader, kotlin.coroutines.Continuation<? super kotlin.Unit> $completion) {
        return null;
    }
    
    private final java.lang.Object relay(java.net.Socket client, java.net.Socket target, kotlin.coroutines.Continuation<? super kotlin.Unit> $completion) {
        return null;
    }
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000\u0018\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\b\n\u0000\n\u0002\u0010\u000e\n\u0000\b\u0086\u0003\u0018\u00002\u00020\u0001B\u0007\b\u0002\u00a2\u0006\u0002\u0010\u0002R\u000e\u0010\u0003\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0005\u001a\u00020\u0006X\u0082T\u00a2\u0006\u0002\n\u0000\u00a8\u0006\u0007"}, d2 = {"Lcom/mobileproxy/core/proxy/HttpProxyServer$Companion;", "", "()V", "BUFFER_SIZE", "", "TAG", "", "app_debug"})
    public static final class Companion {
        
        private Companion() {
            super();
        }
    }
}