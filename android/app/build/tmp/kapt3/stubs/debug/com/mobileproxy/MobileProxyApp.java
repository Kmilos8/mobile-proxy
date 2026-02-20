package com.mobileproxy;

@dagger.hilt.android.HiltAndroidApp()
@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000\u0014\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0010\u0002\n\u0002\b\u0003\b\u0007\u0018\u0000 \u00062\u00020\u0001:\u0001\u0006B\u0005\u00a2\u0006\u0002\u0010\u0002J\b\u0010\u0003\u001a\u00020\u0004H\u0002J\b\u0010\u0005\u001a\u00020\u0004H\u0016\u00a8\u0006\u0007"}, d2 = {"Lcom/mobileproxy/MobileProxyApp;", "Landroid/app/Application;", "()V", "createNotificationChannel", "", "onCreate", "Companion", "app_debug"})
public final class MobileProxyApp extends android.app.Application {
    @org.jetbrains.annotations.NotNull()
    public static final java.lang.String NOTIFICATION_CHANNEL_ID = "proxy_service";
    public static final int NOTIFICATION_ID = 1;
    @org.jetbrains.annotations.NotNull()
    public static final com.mobileproxy.MobileProxyApp.Companion Companion = null;
    
    public MobileProxyApp() {
        super();
    }
    
    @java.lang.Override()
    public void onCreate() {
    }
    
    private final void createNotificationChannel() {
    }
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000\u0018\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\u000e\n\u0000\n\u0002\u0010\b\n\u0000\b\u0086\u0003\u0018\u00002\u00020\u0001B\u0007\b\u0002\u00a2\u0006\u0002\u0010\u0002R\u000e\u0010\u0003\u001a\u00020\u0004X\u0086T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0005\u001a\u00020\u0006X\u0086T\u00a2\u0006\u0002\n\u0000\u00a8\u0006\u0007"}, d2 = {"Lcom/mobileproxy/MobileProxyApp$Companion;", "", "()V", "NOTIFICATION_CHANNEL_ID", "", "NOTIFICATION_ID", "", "app_debug"})
    public static final class Companion {
        
        private Companion() {
            super();
        }
    }
}