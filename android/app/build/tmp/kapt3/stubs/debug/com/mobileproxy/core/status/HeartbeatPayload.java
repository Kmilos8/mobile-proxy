package com.mobileproxy.core.status;

@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000(\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0000\n\u0002\u0010\u000e\n\u0002\b\u0004\n\u0002\u0010\b\n\u0000\n\u0002\u0010\u000b\n\u0002\b\u0003\n\u0002\u0010\t\n\u0002\b \b\u0086\b\u0018\u00002\u00020\u0001BU\u0012\u0006\u0010\u0002\u001a\u00020\u0003\u0012\u0006\u0010\u0004\u001a\u00020\u0003\u0012\u0006\u0010\u0005\u001a\u00020\u0003\u0012\u0006\u0010\u0006\u001a\u00020\u0003\u0012\u0006\u0010\u0007\u001a\u00020\b\u0012\u0006\u0010\t\u001a\u00020\n\u0012\u0006\u0010\u000b\u001a\u00020\b\u0012\u0006\u0010\f\u001a\u00020\u0003\u0012\u0006\u0010\r\u001a\u00020\u000e\u0012\u0006\u0010\u000f\u001a\u00020\u000e\u00a2\u0006\u0002\u0010\u0010J\t\u0010\u001f\u001a\u00020\u0003H\u00c6\u0003J\t\u0010 \u001a\u00020\u000eH\u00c6\u0003J\t\u0010!\u001a\u00020\u0003H\u00c6\u0003J\t\u0010\"\u001a\u00020\u0003H\u00c6\u0003J\t\u0010#\u001a\u00020\u0003H\u00c6\u0003J\t\u0010$\u001a\u00020\bH\u00c6\u0003J\t\u0010%\u001a\u00020\nH\u00c6\u0003J\t\u0010&\u001a\u00020\bH\u00c6\u0003J\t\u0010\'\u001a\u00020\u0003H\u00c6\u0003J\t\u0010(\u001a\u00020\u000eH\u00c6\u0003Jm\u0010)\u001a\u00020\u00002\b\b\u0002\u0010\u0002\u001a\u00020\u00032\b\b\u0002\u0010\u0004\u001a\u00020\u00032\b\b\u0002\u0010\u0005\u001a\u00020\u00032\b\b\u0002\u0010\u0006\u001a\u00020\u00032\b\b\u0002\u0010\u0007\u001a\u00020\b2\b\b\u0002\u0010\t\u001a\u00020\n2\b\b\u0002\u0010\u000b\u001a\u00020\b2\b\b\u0002\u0010\f\u001a\u00020\u00032\b\b\u0002\u0010\r\u001a\u00020\u000e2\b\b\u0002\u0010\u000f\u001a\u00020\u000eH\u00c6\u0001J\u0013\u0010*\u001a\u00020\n2\b\u0010+\u001a\u0004\u0018\u00010\u0001H\u00d6\u0003J\t\u0010,\u001a\u00020\bH\u00d6\u0001J\t\u0010-\u001a\u00020\u0003H\u00d6\u0001R\u0011\u0010\f\u001a\u00020\u0003\u00a2\u0006\b\n\u0000\u001a\u0004\b\u0011\u0010\u0012R\u0011\u0010\t\u001a\u00020\n\u00a2\u0006\b\n\u0000\u001a\u0004\b\u0013\u0010\u0014R\u0011\u0010\u0007\u001a\u00020\b\u00a2\u0006\b\n\u0000\u001a\u0004\b\u0015\u0010\u0016R\u0011\u0010\r\u001a\u00020\u000e\u00a2\u0006\b\n\u0000\u001a\u0004\b\u0017\u0010\u0018R\u0011\u0010\u000f\u001a\u00020\u000e\u00a2\u0006\b\n\u0000\u001a\u0004\b\u0019\u0010\u0018R\u0011\u0010\u0005\u001a\u00020\u0003\u00a2\u0006\b\n\u0000\u001a\u0004\b\u001a\u0010\u0012R\u0011\u0010\u0002\u001a\u00020\u0003\u00a2\u0006\b\n\u0000\u001a\u0004\b\u001b\u0010\u0012R\u0011\u0010\u0006\u001a\u00020\u0003\u00a2\u0006\b\n\u0000\u001a\u0004\b\u001c\u0010\u0012R\u0011\u0010\u000b\u001a\u00020\b\u00a2\u0006\b\n\u0000\u001a\u0004\b\u001d\u0010\u0016R\u0011\u0010\u0004\u001a\u00020\u0003\u00a2\u0006\b\n\u0000\u001a\u0004\b\u001e\u0010\u0012\u00a8\u0006."}, d2 = {"Lcom/mobileproxy/core/status/HeartbeatPayload;", "", "cellular_ip", "", "wifi_ip", "carrier", "network_type", "battery_level", "", "battery_charging", "", "signal_strength", "app_version", "bytes_in", "", "bytes_out", "(Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;IZILjava/lang/String;JJ)V", "getApp_version", "()Ljava/lang/String;", "getBattery_charging", "()Z", "getBattery_level", "()I", "getBytes_in", "()J", "getBytes_out", "getCarrier", "getCellular_ip", "getNetwork_type", "getSignal_strength", "getWifi_ip", "component1", "component10", "component2", "component3", "component4", "component5", "component6", "component7", "component8", "component9", "copy", "equals", "other", "hashCode", "toString", "app_debug"})
public final class HeartbeatPayload {
    @org.jetbrains.annotations.NotNull()
    private final java.lang.String cellular_ip = null;
    @org.jetbrains.annotations.NotNull()
    private final java.lang.String wifi_ip = null;
    @org.jetbrains.annotations.NotNull()
    private final java.lang.String carrier = null;
    @org.jetbrains.annotations.NotNull()
    private final java.lang.String network_type = null;
    private final int battery_level = 0;
    private final boolean battery_charging = false;
    private final int signal_strength = 0;
    @org.jetbrains.annotations.NotNull()
    private final java.lang.String app_version = null;
    private final long bytes_in = 0L;
    private final long bytes_out = 0L;
    
    public HeartbeatPayload(@org.jetbrains.annotations.NotNull()
    java.lang.String cellular_ip, @org.jetbrains.annotations.NotNull()
    java.lang.String wifi_ip, @org.jetbrains.annotations.NotNull()
    java.lang.String carrier, @org.jetbrains.annotations.NotNull()
    java.lang.String network_type, int battery_level, boolean battery_charging, int signal_strength, @org.jetbrains.annotations.NotNull()
    java.lang.String app_version, long bytes_in, long bytes_out) {
        super();
    }
    
    @org.jetbrains.annotations.NotNull()
    public final java.lang.String getCellular_ip() {
        return null;
    }
    
    @org.jetbrains.annotations.NotNull()
    public final java.lang.String getWifi_ip() {
        return null;
    }
    
    @org.jetbrains.annotations.NotNull()
    public final java.lang.String getCarrier() {
        return null;
    }
    
    @org.jetbrains.annotations.NotNull()
    public final java.lang.String getNetwork_type() {
        return null;
    }
    
    public final int getBattery_level() {
        return 0;
    }
    
    public final boolean getBattery_charging() {
        return false;
    }
    
    public final int getSignal_strength() {
        return 0;
    }
    
    @org.jetbrains.annotations.NotNull()
    public final java.lang.String getApp_version() {
        return null;
    }
    
    public final long getBytes_in() {
        return 0L;
    }
    
    public final long getBytes_out() {
        return 0L;
    }
    
    @org.jetbrains.annotations.NotNull()
    public final java.lang.String component1() {
        return null;
    }
    
    public final long component10() {
        return 0L;
    }
    
    @org.jetbrains.annotations.NotNull()
    public final java.lang.String component2() {
        return null;
    }
    
    @org.jetbrains.annotations.NotNull()
    public final java.lang.String component3() {
        return null;
    }
    
    @org.jetbrains.annotations.NotNull()
    public final java.lang.String component4() {
        return null;
    }
    
    public final int component5() {
        return 0;
    }
    
    public final boolean component6() {
        return false;
    }
    
    public final int component7() {
        return 0;
    }
    
    @org.jetbrains.annotations.NotNull()
    public final java.lang.String component8() {
        return null;
    }
    
    public final long component9() {
        return 0L;
    }
    
    @org.jetbrains.annotations.NotNull()
    public final com.mobileproxy.core.status.HeartbeatPayload copy(@org.jetbrains.annotations.NotNull()
    java.lang.String cellular_ip, @org.jetbrains.annotations.NotNull()
    java.lang.String wifi_ip, @org.jetbrains.annotations.NotNull()
    java.lang.String carrier, @org.jetbrains.annotations.NotNull()
    java.lang.String network_type, int battery_level, boolean battery_charging, int signal_strength, @org.jetbrains.annotations.NotNull()
    java.lang.String app_version, long bytes_in, long bytes_out) {
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