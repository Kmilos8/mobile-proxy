package com.mobileproxy.core.rotation

import android.os.Bundle
import android.service.voice.VoiceInteractionSession
import android.service.voice.VoiceInteractionSessionService

/**
 * Required companion to DigitalAssistantService.
 * Returns our custom AirplaneModeSession which handles airplane mode toggle commands
 * via Samsung's VOICE_CONTROL_AIRPLANE_MODE intent.
 */
class DigitalAssistantSessionService : VoiceInteractionSessionService() {

    override fun onNewSession(args: Bundle?): VoiceInteractionSession {
        return AirplaneModeSession(this)
    }
}
