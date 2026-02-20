package com.mobileproxy.core.network;

/**
 * Socket factory that creates sockets bound to the cellular network.
 * Ensures all proxy outbound traffic goes through mobile IP, not WiFi.
 */
@javax.inject.Singleton()
@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000,\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\b\n\u0002\b\u0003\n\u0002\u0010\u000e\n\u0000\b\u0007\u0018\u00002\u00020\u0001B\u000f\b\u0007\u0012\u0006\u0010\u0002\u001a\u00020\u0003\u00a2\u0006\u0002\u0010\u0004J\b\u0010\u0005\u001a\u00020\u0006H\u0016J\u0018\u0010\u0005\u001a\u00020\u00062\u0006\u0010\u0007\u001a\u00020\b2\u0006\u0010\t\u001a\u00020\nH\u0016J(\u0010\u0005\u001a\u00020\u00062\u0006\u0010\u0007\u001a\u00020\b2\u0006\u0010\t\u001a\u00020\n2\u0006\u0010\u000b\u001a\u00020\b2\u0006\u0010\f\u001a\u00020\nH\u0016J\u0018\u0010\u0005\u001a\u00020\u00062\u0006\u0010\r\u001a\u00020\u000e2\u0006\u0010\t\u001a\u00020\nH\u0016J(\u0010\u0005\u001a\u00020\u00062\u0006\u0010\r\u001a\u00020\u000e2\u0006\u0010\t\u001a\u00020\n2\u0006\u0010\u000b\u001a\u00020\b2\u0006\u0010\f\u001a\u00020\nH\u0016R\u000e\u0010\u0002\u001a\u00020\u0003X\u0082\u0004\u00a2\u0006\u0002\n\u0000\u00a8\u0006\u000f"}, d2 = {"Lcom/mobileproxy/core/network/CellularSocketFactory;", "Ljavax/net/SocketFactory;", "networkManager", "Lcom/mobileproxy/core/network/NetworkManager;", "(Lcom/mobileproxy/core/network/NetworkManager;)V", "createSocket", "Ljava/net/Socket;", "addr", "Ljava/net/InetAddress;", "port", "", "localAddr", "localPort", "host", "", "app_debug"})
public final class CellularSocketFactory extends javax.net.SocketFactory {
    @org.jetbrains.annotations.NotNull()
    private final com.mobileproxy.core.network.NetworkManager networkManager = null;
    
    @javax.inject.Inject()
    public CellularSocketFactory(@org.jetbrains.annotations.NotNull()
    com.mobileproxy.core.network.NetworkManager networkManager) {
        super();
    }
    
    @java.lang.Override()
    @org.jetbrains.annotations.NotNull()
    public java.net.Socket createSocket() {
        return null;
    }
    
    @java.lang.Override()
    @org.jetbrains.annotations.NotNull()
    public java.net.Socket createSocket(@org.jetbrains.annotations.NotNull()
    java.lang.String host, int port) {
        return null;
    }
    
    @java.lang.Override()
    @org.jetbrains.annotations.NotNull()
    public java.net.Socket createSocket(@org.jetbrains.annotations.NotNull()
    java.lang.String host, int port, @org.jetbrains.annotations.NotNull()
    java.net.InetAddress localAddr, int localPort) {
        return null;
    }
    
    @java.lang.Override()
    @org.jetbrains.annotations.NotNull()
    public java.net.Socket createSocket(@org.jetbrains.annotations.NotNull()
    java.net.InetAddress addr, int port) {
        return null;
    }
    
    @java.lang.Override()
    @org.jetbrains.annotations.NotNull()
    public java.net.Socket createSocket(@org.jetbrains.annotations.NotNull()
    java.net.InetAddress addr, int port, @org.jetbrains.annotations.NotNull()
    java.net.InetAddress localAddr, int localPort) {
        return null;
    }
}