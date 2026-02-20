package com.mobileproxy.core.rotation;

/**
 * Required companion to DigitalAssistantService.
 * Returns our custom AirplaneModeSession which handles airplane mode toggle commands
 * via Samsung's VOICE_CONTROL_AIRPLANE_MODE intent.
 */
@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000\u0018\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\u0018\u00002\u00020\u0001B\u0005\u00a2\u0006\u0002\u0010\u0002J\u0012\u0010\u0003\u001a\u00020\u00042\b\u0010\u0005\u001a\u0004\u0018\u00010\u0006H\u0016\u00a8\u0006\u0007"}, d2 = {"Lcom/mobileproxy/core/rotation/DigitalAssistantSessionService;", "Landroid/service/voice/VoiceInteractionSessionService;", "()V", "onNewSession", "Landroid/service/voice/VoiceInteractionSession;", "args", "Landroid/os/Bundle;", "app_debug"})
public final class DigitalAssistantSessionService extends android.service.voice.VoiceInteractionSessionService {
    
    public DigitalAssistantSessionService() {
        super();
    }
    
    @java.lang.Override()
    @org.jetbrains.annotations.NotNull()
    public android.service.voice.VoiceInteractionSession onNewSession(@org.jetbrains.annotations.Nullable()
    android.os.Bundle args) {
        return null;
    }
}