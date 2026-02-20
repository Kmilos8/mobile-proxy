package com.mobileproxy.core.rotation;

/**
 * Custom VoiceInteractionSession that handles airplane mode toggle commands.
 *
 * When a SwitchAirplaneMode command is received via the session Bundle,
 * it calls startVoiceActivity() with Samsung's VOICE_CONTROL_AIRPLANE_MODE intent.
 * Samsung's Settings app trusts this because it comes from the active voice assistant session.
 */
@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u00000\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\u0002\n\u0002\b\u0003\n\u0002\u0018\u0002\n\u0002\b\u0004\n\u0002\u0010\b\n\u0002\b\u0003\u0018\u0000 \u00132\u00020\u0001:\u0001\u0013B\r\u0012\u0006\u0010\u0002\u001a\u00020\u0003\u00a2\u0006\u0002\u0010\u0004J\u0010\u0010\u0007\u001a\u00020\b2\u0006\u0010\t\u001a\u00020\u0006H\u0002J\u0014\u0010\n\u001a\u0004\u0018\u00010\u00062\b\u0010\u000b\u001a\u0004\u0018\u00010\fH\u0002J\b\u0010\r\u001a\u00020\bH\u0016J\b\u0010\u000e\u001a\u00020\bH\u0016J\u001a\u0010\u000f\u001a\u00020\b2\b\u0010\u000b\u001a\u0004\u0018\u00010\f2\u0006\u0010\u0010\u001a\u00020\u0011H\u0016J\u001a\u0010\u0012\u001a\u00020\b2\b\u0010\u000b\u001a\u0004\u0018\u00010\f2\u0006\u0010\u0010\u001a\u00020\u0011H\u0016R\u0010\u0010\u0005\u001a\u0004\u0018\u00010\u0006X\u0082\u000e\u00a2\u0006\u0002\n\u0000\u00a8\u0006\u0014"}, d2 = {"Lcom/mobileproxy/core/rotation/AirplaneModeSession;", "Landroid/service/voice/VoiceInteractionSession;", "context", "Landroid/service/voice/VoiceInteractionSessionService;", "(Landroid/service/voice/VoiceInteractionSessionService;)V", "lastCommand", "Lcom/mobileproxy/core/rotation/AirplaneModeCommand;", "executeCommand", "", "command", "extractCommand", "args", "Landroid/os/Bundle;", "onCreate", "onDestroy", "onPrepareShow", "showFlags", "", "onShow", "Companion", "app_debug"})
public final class AirplaneModeSession extends android.service.voice.VoiceInteractionSession {
    @org.jetbrains.annotations.NotNull()
    private static final java.lang.String TAG = "AirplaneModeSession";
    @org.jetbrains.annotations.NotNull()
    private static final java.lang.String SAMSUNG_AIRPLANE_ACTION = "android.settings.VOICE_CONTROL_AIRPLANE_MODE";
    @org.jetbrains.annotations.NotNull()
    private static final java.lang.String EXTRA_AIRPLANE_ENABLED = "airplane_mode_enabled";
    @org.jetbrains.annotations.Nullable()
    private com.mobileproxy.core.rotation.AirplaneModeCommand lastCommand;
    @org.jetbrains.annotations.NotNull()
    public static final com.mobileproxy.core.rotation.AirplaneModeSession.Companion Companion = null;
    
    public AirplaneModeSession(@org.jetbrains.annotations.NotNull()
    android.service.voice.VoiceInteractionSessionService context) {
        super(null);
    }
    
    @java.lang.Override()
    public void onCreate() {
    }
    
    @java.lang.Override()
    public void onPrepareShow(@org.jetbrains.annotations.Nullable()
    android.os.Bundle args, int showFlags) {
    }
    
    @java.lang.Override()
    public void onShow(@org.jetbrains.annotations.Nullable()
    android.os.Bundle args, int showFlags) {
    }
    
    @java.lang.Override()
    public void onDestroy() {
    }
    
    private final com.mobileproxy.core.rotation.AirplaneModeCommand extractCommand(android.os.Bundle args) {
        return null;
    }
    
    private final void executeCommand(com.mobileproxy.core.rotation.AirplaneModeCommand command) {
    }
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000\u0014\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\u000e\n\u0002\b\u0003\b\u0086\u0003\u0018\u00002\u00020\u0001B\u0007\b\u0002\u00a2\u0006\u0002\u0010\u0002R\u000e\u0010\u0003\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0005\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0006\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000\u00a8\u0006\u0007"}, d2 = {"Lcom/mobileproxy/core/rotation/AirplaneModeSession$Companion;", "", "()V", "EXTRA_AIRPLANE_ENABLED", "", "SAMSUNG_AIRPLANE_ACTION", "TAG", "app_debug"})
    public static final class Companion {
        
        private Companion() {
            super();
        }
    }
}