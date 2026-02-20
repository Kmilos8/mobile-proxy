package com.mobileproxy.core.rotation;

@javax.inject.Singleton()
@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u00008\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\u0002\n\u0002\b\u0002\n\u0002\u0010\u000b\n\u0002\b\b\b\u0007\u0018\u0000 \u00172\u00020\u0001:\u0001\u0017B\'\b\u0007\u0012\b\b\u0001\u0010\u0002\u001a\u00020\u0003\u0012\u0006\u0010\u0004\u001a\u00020\u0005\u0012\f\u0010\u0006\u001a\b\u0012\u0004\u0012\u00020\b0\u0007\u00a2\u0006\u0002\u0010\tJ\u000e\u0010\f\u001a\u00020\rH\u0086@\u00a2\u0006\u0002\u0010\u000eJ\u000e\u0010\u000f\u001a\u00020\u0010H\u0086@\u00a2\u0006\u0002\u0010\u000eJ\u000e\u0010\u0011\u001a\u00020\rH\u0082@\u00a2\u0006\u0002\u0010\u000eJ\u000e\u0010\u0012\u001a\u00020\u0010H\u0082@\u00a2\u0006\u0002\u0010\u000eJ\u000e\u0010\u0013\u001a\u00020\u0010H\u0082@\u00a2\u0006\u0002\u0010\u000eJ\u0016\u0010\u0014\u001a\u00020\u00102\u0006\u0010\u0015\u001a\u00020\u0010H\u0082@\u00a2\u0006\u0002\u0010\u0016R\u000e\u0010\n\u001a\u00020\u000bX\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0002\u001a\u00020\u0003X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0004\u001a\u00020\u0005X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u0014\u0010\u0006\u001a\b\u0012\u0004\u0012\u00020\b0\u0007X\u0082\u0004\u00a2\u0006\u0002\n\u0000\u00a8\u0006\u0018"}, d2 = {"Lcom/mobileproxy/core/rotation/IPRotationManager;", "", "context", "Landroid/content/Context;", "networkManager", "Lcom/mobileproxy/core/network/NetworkManager;", "statusReporter", "Ldagger/Lazy;", "Lcom/mobileproxy/core/status/DeviceStatusReporter;", "(Landroid/content/Context;Lcom/mobileproxy/core/network/NetworkManager;Ldagger/Lazy;)V", "commandIdCounter", "Ljava/util/concurrent/atomic/AtomicInteger;", "requestAirplaneModeToggle", "", "(Lkotlin/coroutines/Continuation;)Ljava/lang/Object;", "rotateByCellularReconnect", "", "rotateViaAccessibility", "rotateViaSettingsGlobal", "rotateViaVoiceAssistant", "toggleAirplaneModeViaVoice", "enable", "(ZLkotlin/coroutines/Continuation;)Ljava/lang/Object;", "Companion", "app_debug"})
public final class IPRotationManager {
    @org.jetbrains.annotations.NotNull()
    private final android.content.Context context = null;
    @org.jetbrains.annotations.NotNull()
    private final com.mobileproxy.core.network.NetworkManager networkManager = null;
    @org.jetbrains.annotations.NotNull()
    private final dagger.Lazy<com.mobileproxy.core.status.DeviceStatusReporter> statusReporter = null;
    @org.jetbrains.annotations.NotNull()
    private static final java.lang.String TAG = "IPRotationManager";
    private static final long AIRPLANE_MODE_TOGGLE_TIMEOUT_MS = 7000L;
    private static final long AIRPLANE_MODE_OFF_DELAY_MS = 6000L;
    private static final long CELLULAR_RECONNECT_DELAY_MS = 5000L;
    @org.jetbrains.annotations.NotNull()
    private final java.util.concurrent.atomic.AtomicInteger commandIdCounter = null;
    @org.jetbrains.annotations.NotNull()
    public static final com.mobileproxy.core.rotation.IPRotationManager.Companion Companion = null;
    
    @javax.inject.Inject()
    public IPRotationManager(@dagger.hilt.android.qualifiers.ApplicationContext()
    @org.jetbrains.annotations.NotNull()
    android.content.Context context, @org.jetbrains.annotations.NotNull()
    com.mobileproxy.core.network.NetworkManager networkManager, @org.jetbrains.annotations.NotNull()
    dagger.Lazy<com.mobileproxy.core.status.DeviceStatusReporter> statusReporter) {
        super();
    }
    
    /**
     * Cellular reconnect - no airplane mode, just reconnects cellular data.
     */
    @org.jetbrains.annotations.Nullable()
    public final java.lang.Object rotateByCellularReconnect(@org.jetbrains.annotations.NotNull()
    kotlin.coroutines.Continuation<? super java.lang.Boolean> $completion) {
        return null;
    }
    
    /**
     * Main rotation entry point — uses Voice Assistant session to toggle airplane mode
     * via Samsung's VOICE_CONTROL_AIRPLANE_MODE intent. No WRITE_SECURE_SETTINGS needed.
     *
     * Falls back to Settings.Global (requires ADB grant) and then Accessibility Service.
     */
    @org.jetbrains.annotations.Nullable()
    public final java.lang.Object requestAirplaneModeToggle(@org.jetbrains.annotations.NotNull()
    kotlin.coroutines.Continuation<? super kotlin.Unit> $completion) {
        return null;
    }
    
    /**
     * Toggle airplane mode ON then OFF via Samsung's Voice Assistant intent.
     * This works without WRITE_SECURE_SETTINGS because Samsung trusts the active
     * voice assistant's VoiceInteractionSession.startVoiceActivity() calls.
     */
    private final java.lang.Object rotateViaVoiceAssistant(kotlin.coroutines.Continuation<? super java.lang.Boolean> $completion) {
        return null;
    }
    
    /**
     * Send a command to the DigitalAssistantService and wait for the airplane mode
     * broadcast to confirm the toggle happened.
     */
    private final java.lang.Object toggleAirplaneModeViaVoice(boolean enable, kotlin.coroutines.Continuation<? super java.lang.Boolean> $completion) {
        return null;
    }
    
    /**
     * Accessibility Service — opens Quick Settings and taps airplane mode tile.
     */
    private final java.lang.Object rotateViaAccessibility(kotlin.coroutines.Continuation<? super kotlin.Unit> $completion) {
        return null;
    }
    
    /**
     * Settings.Global write (legacy fallback).
     * Requires WRITE_SECURE_SETTINGS (granted via ADB pm grant).
     */
    private final java.lang.Object rotateViaSettingsGlobal(kotlin.coroutines.Continuation<? super java.lang.Boolean> $completion) {
        return null;
    }
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000\u001a\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\t\n\u0002\b\u0003\n\u0002\u0010\u000e\n\u0000\b\u0086\u0003\u0018\u00002\u00020\u0001B\u0007\b\u0002\u00a2\u0006\u0002\u0010\u0002R\u000e\u0010\u0003\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0005\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0006\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0007\u001a\u00020\bX\u0082T\u00a2\u0006\u0002\n\u0000\u00a8\u0006\t"}, d2 = {"Lcom/mobileproxy/core/rotation/IPRotationManager$Companion;", "", "()V", "AIRPLANE_MODE_OFF_DELAY_MS", "", "AIRPLANE_MODE_TOGGLE_TIMEOUT_MS", "CELLULAR_RECONNECT_DELAY_MS", "TAG", "", "app_debug"})
    public static final class Companion {
        
        private Companion() {
            super();
        }
    }
}