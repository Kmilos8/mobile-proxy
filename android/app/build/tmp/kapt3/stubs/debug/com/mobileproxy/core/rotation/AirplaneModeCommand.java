package com.mobileproxy.core.rotation;

/**
 * Command sent to the VoiceInteractionSession to toggle airplane mode.
 * Passed via Bundle through showSession() → onShow()/onPrepareShow().
 */
@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u00006\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0010\b\n\u0000\n\u0002\u0010\u000b\n\u0002\b\u000b\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\u000e\n\u0000\n\u0002\u0010\u0002\n\u0002\b\u0004\b\u0086\b\u0018\u0000 \u001c2\u00020\u0001:\u0001\u001cB\u000f\b\u0016\u0012\u0006\u0010\u0002\u001a\u00020\u0003\u00a2\u0006\u0002\u0010\u0004B\u0015\u0012\u0006\u0010\u0005\u001a\u00020\u0006\u0012\u0006\u0010\u0007\u001a\u00020\b\u00a2\u0006\u0002\u0010\tJ\t\u0010\u000e\u001a\u00020\u0006H\u00c6\u0003J\t\u0010\u000f\u001a\u00020\bH\u00c6\u0003J\u001d\u0010\u0010\u001a\u00020\u00002\b\b\u0002\u0010\u0005\u001a\u00020\u00062\b\b\u0002\u0010\u0007\u001a\u00020\bH\u00c6\u0001J\b\u0010\u0011\u001a\u00020\u0006H\u0016J\u0013\u0010\u0012\u001a\u00020\b2\b\u0010\u0013\u001a\u0004\u0018\u00010\u0014H\u00d6\u0003J\t\u0010\u0015\u001a\u00020\u0006H\u00d6\u0001J\t\u0010\u0016\u001a\u00020\u0017H\u00d6\u0001J\u0018\u0010\u0018\u001a\u00020\u00192\u0006\u0010\u001a\u001a\u00020\u00032\u0006\u0010\u001b\u001a\u00020\u0006H\u0016R\u0011\u0010\u0007\u001a\u00020\b\u00a2\u0006\b\n\u0000\u001a\u0004\b\n\u0010\u000bR\u0011\u0010\u0005\u001a\u00020\u0006\u00a2\u0006\b\n\u0000\u001a\u0004\b\f\u0010\r\u00a8\u0006\u001d"}, d2 = {"Lcom/mobileproxy/core/rotation/AirplaneModeCommand;", "Landroid/os/Parcelable;", "parcel", "Landroid/os/Parcel;", "(Landroid/os/Parcel;)V", "id", "", "enable", "", "(IZ)V", "getEnable", "()Z", "getId", "()I", "component1", "component2", "copy", "describeContents", "equals", "other", "", "hashCode", "toString", "", "writeToParcel", "", "dest", "flags", "CREATOR", "app_debug"})
public final class AirplaneModeCommand implements android.os.Parcelable {
    private final int id = 0;
    private final boolean enable = false;
    @org.jetbrains.annotations.NotNull()
    public static final java.lang.String KEY = "command";
    @org.jetbrains.annotations.NotNull()
    public static final com.mobileproxy.core.rotation.AirplaneModeCommand.CREATOR CREATOR = null;
    
    public AirplaneModeCommand(int id, boolean enable) {
        super();
    }
    
    public final int getId() {
        return 0;
    }
    
    public final boolean getEnable() {
        return false;
    }
    
    public AirplaneModeCommand(@org.jetbrains.annotations.NotNull()
    android.os.Parcel parcel) {
        super();
    }
    
    @java.lang.Override()
    public void writeToParcel(@org.jetbrains.annotations.NotNull()
    android.os.Parcel dest, int flags) {
    }
    
    @java.lang.Override()
    public int describeContents() {
        return 0;
    }
    
    public final int component1() {
        return 0;
    }
    
    public final boolean component2() {
        return false;
    }
    
    @org.jetbrains.annotations.NotNull()
    public final com.mobileproxy.core.rotation.AirplaneModeCommand copy(int id, boolean enable) {
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
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000,\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0010\u000e\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\u0011\n\u0000\n\u0002\u0010\b\n\u0002\b\u0002\b\u0086\u0003\u0018\u00002\b\u0012\u0004\u0012\u00020\u00020\u0001B\u0007\b\u0002\u00a2\u0006\u0002\u0010\u0003J\u0010\u0010\u0006\u001a\u00020\u00022\u0006\u0010\u0007\u001a\u00020\bH\u0016J\u001d\u0010\t\u001a\n\u0012\u0006\u0012\u0004\u0018\u00010\u00020\n2\u0006\u0010\u000b\u001a\u00020\fH\u0016\u00a2\u0006\u0002\u0010\rR\u000e\u0010\u0004\u001a\u00020\u0005X\u0086T\u00a2\u0006\u0002\n\u0000\u00a8\u0006\u000e"}, d2 = {"Lcom/mobileproxy/core/rotation/AirplaneModeCommand$CREATOR;", "Landroid/os/Parcelable$Creator;", "Lcom/mobileproxy/core/rotation/AirplaneModeCommand;", "()V", "KEY", "", "createFromParcel", "parcel", "Landroid/os/Parcel;", "newArray", "", "size", "", "(I)[Lcom/mobileproxy/core/rotation/AirplaneModeCommand;", "app_debug"})
    public static final class CREATOR implements android.os.Parcelable.Creator<com.mobileproxy.core.rotation.AirplaneModeCommand> {
        
        private CREATOR() {
            super();
        }
        
        @java.lang.Override()
        @org.jetbrains.annotations.NotNull()
        public com.mobileproxy.core.rotation.AirplaneModeCommand createFromParcel(@org.jetbrains.annotations.NotNull()
        android.os.Parcel parcel) {
            return null;
        }
        
        @java.lang.Override()
        @org.jetbrains.annotations.NotNull()
        public com.mobileproxy.core.rotation.AirplaneModeCommand[] newArray(int size) {
            return null;
        }
    }
}