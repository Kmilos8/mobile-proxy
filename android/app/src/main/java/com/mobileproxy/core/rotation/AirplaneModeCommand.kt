package com.mobileproxy.core.rotation

import android.os.Parcel
import android.os.Parcelable

/**
 * Command sent to the VoiceInteractionSession to toggle airplane mode.
 * Passed via Bundle through showSession() → onShow()/onPrepareShow().
 */
data class AirplaneModeCommand(
    val id: Int,
    val enable: Boolean
) : Parcelable {

    constructor(parcel: Parcel) : this(
        id = parcel.readInt(),
        enable = parcel.readInt() != 0
    )

    override fun writeToParcel(dest: Parcel, flags: Int) {
        dest.writeInt(id)
        dest.writeInt(if (enable) 1 else 0)
    }

    override fun describeContents(): Int = 0

    companion object CREATOR : Parcelable.Creator<AirplaneModeCommand> {
        const val KEY = "command"

        override fun createFromParcel(parcel: Parcel): AirplaneModeCommand {
            return AirplaneModeCommand(parcel)
        }

        override fun newArray(size: Int): Array<AirplaneModeCommand?> {
            return arrayOfNulls(size)
        }
    }
}
