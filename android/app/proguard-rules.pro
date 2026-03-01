# Keep Gson serialized classes
-keepattributes Signature
-keepattributes *Annotation*

# Keep data classes used with Gson (field names must not be obfuscated)
-keep class com.mobileproxy.core.status.HeartbeatResponse { *; }
-keep class com.mobileproxy.core.status.HeartbeatPayload { *; }
-keep class com.mobileproxy.core.status.ProxyCredentialResponse { *; }
-keep class com.mobileproxy.core.commands.DeviceCommand { *; }

# Keep Retrofit interfaces
-keep,allowobfuscation interface * {
    @retrofit2.http.* <methods>;
}

# Keep Hilt generated classes
-keep class dagger.hilt.** { *; }
-keep class * extends dagger.hilt.android.internal.managers.ViewComponentManager$FragmentContextWrapper { *; }
