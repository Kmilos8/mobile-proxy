package com.mobileproxy.ui;

@dagger.hilt.android.AndroidEntryPoint()
@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000N\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0002\b\u0005\n\u0002\u0018\u0002\n\u0002\b\u0005\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0010\u000b\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0005\n\u0002\u0018\u0002\n\u0002\b\u0003\n\u0002\u0010\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0002\b\u0004\b\u0007\u0018\u00002\u00020\u0001B\u0005\u00a2\u0006\u0002\u0010\u0002J\b\u0010\u001f\u001a\u00020 H\u0002J\u0012\u0010!\u001a\u00020 2\b\u0010\"\u001a\u0004\u0018\u00010#H\u0014J\b\u0010$\u001a\u00020 H\u0014J\b\u0010%\u001a\u00020 H\u0002J\b\u0010&\u001a\u00020 H\u0002R\u001e\u0010\u0003\u001a\u00020\u00048\u0006@\u0006X\u0087.\u00a2\u0006\u000e\n\u0000\u001a\u0004\b\u0005\u0010\u0006\"\u0004\b\u0007\u0010\bR\u001e\u0010\t\u001a\u00020\n8\u0006@\u0006X\u0087.\u00a2\u0006\u000e\n\u0000\u001a\u0004\b\u000b\u0010\f\"\u0004\b\r\u0010\u000eR\u001c\u0010\u000f\u001a\u0010\u0012\f\u0012\n \u0012*\u0004\u0018\u00010\u00110\u00110\u0010X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0013\u001a\u00020\u0014X\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u001e\u0010\u0015\u001a\u00020\u00168\u0006@\u0006X\u0087.\u00a2\u0006\u000e\n\u0000\u001a\u0004\b\u0017\u0010\u0018\"\u0004\b\u0019\u0010\u001aR\u000e\u0010\u001b\u001a\u00020\u001cX\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u001c\u0010\u001d\u001a\u0010\u0012\f\u0012\n \u0012*\u0004\u0018\u00010\u00110\u00110\u0010X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u001c\u0010\u001e\u001a\u0010\u0012\f\u0012\n \u0012*\u0004\u0018\u00010\u00110\u00110\u0010X\u0082\u0004\u00a2\u0006\u0002\n\u0000\u00a8\u0006\'"}, d2 = {"Lcom/mobileproxy/ui/MainActivity;", "Landroidx/appcompat/app/AppCompatActivity;", "()V", "credentialManager", "Lcom/mobileproxy/core/config/CredentialManager;", "getCredentialManager", "()Lcom/mobileproxy/core/config/CredentialManager;", "setCredentialManager", "(Lcom/mobileproxy/core/config/CredentialManager;)V", "networkManager", "Lcom/mobileproxy/core/network/NetworkManager;", "getNetworkManager", "()Lcom/mobileproxy/core/network/NetworkManager;", "setNetworkManager", "(Lcom/mobileproxy/core/network/NetworkManager;)V", "pairingLauncher", "Landroidx/activity/result/ActivityResultLauncher;", "Landroid/content/Intent;", "kotlin.jvm.PlatformType", "pendingStart", "", "rotationManager", "Lcom/mobileproxy/core/rotation/IPRotationManager;", "getRotationManager", "()Lcom/mobileproxy/core/rotation/IPRotationManager;", "setRotationManager", "(Lcom/mobileproxy/core/rotation/IPRotationManager;)V", "scope", "Lkotlinx/coroutines/CoroutineScope;", "setupLauncher", "vpnPermissionLauncher", "migrateFromBuildConfig", "", "onCreate", "savedInstanceState", "Landroid/os/Bundle;", "onDestroy", "startProxyService", "updateServerUrlFromCredentials", "app_debug"})
public final class MainActivity extends androidx.appcompat.app.AppCompatActivity {
    @javax.inject.Inject()
    public com.mobileproxy.core.network.NetworkManager networkManager;
    @javax.inject.Inject()
    public com.mobileproxy.core.rotation.IPRotationManager rotationManager;
    @javax.inject.Inject()
    public com.mobileproxy.core.config.CredentialManager credentialManager;
    @org.jetbrains.annotations.NotNull()
    private final kotlinx.coroutines.CoroutineScope scope = null;
    private boolean pendingStart = false;
    @org.jetbrains.annotations.NotNull()
    private final androidx.activity.result.ActivityResultLauncher<android.content.Intent> vpnPermissionLauncher = null;
    @org.jetbrains.annotations.NotNull()
    private final androidx.activity.result.ActivityResultLauncher<android.content.Intent> setupLauncher = null;
    @org.jetbrains.annotations.NotNull()
    private final androidx.activity.result.ActivityResultLauncher<android.content.Intent> pairingLauncher = null;
    
    public MainActivity() {
        super();
    }
    
    @org.jetbrains.annotations.NotNull()
    public final com.mobileproxy.core.network.NetworkManager getNetworkManager() {
        return null;
    }
    
    public final void setNetworkManager(@org.jetbrains.annotations.NotNull()
    com.mobileproxy.core.network.NetworkManager p0) {
    }
    
    @org.jetbrains.annotations.NotNull()
    public final com.mobileproxy.core.rotation.IPRotationManager getRotationManager() {
        return null;
    }
    
    public final void setRotationManager(@org.jetbrains.annotations.NotNull()
    com.mobileproxy.core.rotation.IPRotationManager p0) {
    }
    
    @org.jetbrains.annotations.NotNull()
    public final com.mobileproxy.core.config.CredentialManager getCredentialManager() {
        return null;
    }
    
    public final void setCredentialManager(@org.jetbrains.annotations.NotNull()
    com.mobileproxy.core.config.CredentialManager p0) {
    }
    
    @java.lang.Override()
    protected void onCreate(@org.jetbrains.annotations.Nullable()
    android.os.Bundle savedInstanceState) {
    }
    
    private final void migrateFromBuildConfig() {
    }
    
    private final void updateServerUrlFromCredentials() {
    }
    
    private final void startProxyService() {
    }
    
    @java.lang.Override()
    protected void onDestroy() {
    }
}