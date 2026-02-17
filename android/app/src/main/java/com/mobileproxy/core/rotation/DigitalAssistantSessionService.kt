package com.mobileproxy.core.rotation

import android.os.Bundle
import android.service.voice.VoiceInteractionSession
import android.service.voice.VoiceInteractionSessionService

/**
 * Required companion to DigitalAssistantService.
 * Android won't accept a VoiceInteractionService without a session service.
 */
class DigitalAssistantSessionService : VoiceInteractionSessionService() {

    override fun onNewSession(args: Bundle?): VoiceInteractionSession {
        return VoiceInteractionSession(this)
    }
}
