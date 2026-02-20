package com.mobileproxy.core.rotation;

/**
 * VoiceInteractionService that enables airplane mode toggling via Samsung's
 * VOICE_CONTROL_AIRPLANE_MODE intent.
 *
 * Flow:
 * 1. IPRotationManager sends a local broadcast with ACTION_VOICE_COMMAND
 * 2. This service receives it and calls showSession() with the command in the Bundle
 * 3. AirplaneModeSession.onPrepareShow()/onShow() calls startVoiceActivity()
 *   with Samsung's intent, which toggles airplane mode without WRITE_SECURE_SETTINGS
 */
@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000\u001a\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\u0002\n\u0002\b\u0005\u0018\u0000 \n2\u00020\u0001:\u0001\nB\u0005\u00a2\u0006\u0002\u0010\u0002J\b\u0010\u0005\u001a\u00020\u0006H\u0016J\b\u0010\u0007\u001a\u00020\u0006H\u0016J\b\u0010\b\u001a\u00020\u0006H\u0016J\b\u0010\t\u001a\u00020\u0006H\u0002R\u0010\u0010\u0003\u001a\u0004\u0018\u00010\u0004X\u0082\u000e\u00a2\u0006\u0002\n\u0000\u00a8\u0006\u000b"}, d2 = {"Lcom/mobileproxy/core/rotation/DigitalAssistantService;", "Landroid/service/voice/VoiceInteractionService;", "()V", "commandReceiver", "Landroid/content/BroadcastReceiver;", "onDestroy", "", "onReady", "onShutdown", "unregisterCommandReceiver", "Companion", "app_debug"})
public final class DigitalAssistantService extends android.service.voice.VoiceInteractionService {
    @org.jetbrains.annotations.NotNull()
    private static final java.lang.String TAG = "DigitalAssistantSvc";
    @org.jetbrains.annotations.NotNull()
    public static final java.lang.String ACTION_VOICE_COMMAND = "com.mobileproxy.intent.action.VOICE_COMMAND";
    @kotlin.jvm.Volatile()
    private static volatile boolean isReady = false;
    @org.jetbrains.annotations.Nullable()
    private android.content.BroadcastReceiver commandReceiver;
    @org.jetbrains.annotations.NotNull()
    public static final com.mobileproxy.core.rotation.DigitalAssistantService.Companion Companion = null;
    
    public DigitalAssistantService() {
        super();
    }
    
    @java.lang.Override()
    public void onReady() {
    }
    
    @java.lang.Override()
    public void onShutdown() {
    }
    
    @java.lang.Override()
    public void onDestroy() {
    }
    
    private final void unregisterCommandReceiver() {
    }
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000\u001c\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\u000e\n\u0002\b\u0002\n\u0002\u0010\u000b\n\u0002\b\u0003\b\u0086\u0003\u0018\u00002\u00020\u0001B\u0007\b\u0002\u00a2\u0006\u0002\u0010\u0002R\u000e\u0010\u0003\u001a\u00020\u0004X\u0086T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0005\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u001e\u0010\b\u001a\u00020\u00072\u0006\u0010\u0006\u001a\u00020\u0007@BX\u0086\u000e\u00a2\u0006\b\n\u0000\u001a\u0004\b\b\u0010\t\u00a8\u0006\n"}, d2 = {"Lcom/mobileproxy/core/rotation/DigitalAssistantService$Companion;", "", "()V", "ACTION_VOICE_COMMAND", "", "TAG", "<set-?>", "", "isReady", "()Z", "app_debug"})
    public static final class Companion {
        
        private Companion() {
            super();
        }
        
        public final boolean isReady() {
            return false;
        }
    }
}