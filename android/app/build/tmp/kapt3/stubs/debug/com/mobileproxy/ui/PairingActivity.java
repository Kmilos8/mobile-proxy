package com.mobileproxy.ui;

@dagger.hilt.android.AndroidEntryPoint()
@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000B\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0002\u0010\u000e\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0005\n\u0002\u0010\u000b\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\u0002\n\u0002\b\u0003\n\u0002\u0018\u0002\n\u0002\b\u000b\b\u0007\u0018\u0000 \"2\u00020\u0001:\u0001\"B\u0005\u00a2\u0006\u0002\u0010\u0002J\u0010\u0010\u0013\u001a\u00020\u00142\u0006\u0010\u0015\u001a\u00020\u0005H\u0002J\u0012\u0010\u0016\u001a\u00020\u00142\b\u0010\u0017\u001a\u0004\u0018\u00010\u0018H\u0014J\b\u0010\u0019\u001a\u00020\u0014H\u0014J\u0018\u0010\u001a\u001a\u00020\u00142\u0006\u0010\u001b\u001a\u00020\u00052\u0006\u0010\u001c\u001a\u00020\u0005H\u0002J\u0010\u0010\u001d\u001a\u00020\u00142\u0006\u0010\u001e\u001a\u00020\u0005H\u0002J\b\u0010\u001f\u001a\u00020\u0014H\u0002J\b\u0010 \u001a\u00020\u0014H\u0002J\b\u0010!\u001a\u00020\u0014H\u0003R\u001c\u0010\u0003\u001a\u0010\u0012\f\u0012\n \u0006*\u0004\u0018\u00010\u00050\u00050\u0004X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u0010\u0010\u0007\u001a\u0004\u0018\u00010\bX\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u001e\u0010\t\u001a\u00020\n8\u0006@\u0006X\u0087.\u00a2\u0006\u000e\n\u0000\u001a\u0004\b\u000b\u0010\f\"\u0004\b\r\u0010\u000eR\u000e\u0010\u000f\u001a\u00020\u0010X\u0082\u000e\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0011\u001a\u00020\u0012X\u0082\u0004\u00a2\u0006\u0002\n\u0000\u00a8\u0006#"}, d2 = {"Lcom/mobileproxy/ui/PairingActivity;", "Landroidx/appcompat/app/AppCompatActivity;", "()V", "cameraPermissionLauncher", "Landroidx/activity/result/ActivityResultLauncher;", "", "kotlin.jvm.PlatformType", "cameraProvider", "Landroidx/camera/lifecycle/ProcessCameraProvider;", "credentialManager", "Lcom/mobileproxy/core/config/CredentialManager;", "getCredentialManager", "()Lcom/mobileproxy/core/config/CredentialManager;", "setCredentialManager", "(Lcom/mobileproxy/core/config/CredentialManager;)V", "isPairing", "", "scope", "Lkotlinx/coroutines/CoroutineScope;", "handleQrCode", "", "qrValue", "onCreate", "savedInstanceState", "Landroid/os/Bundle;", "onDestroy", "pair", "serverUrl", "code", "showError", "message", "showManualEntry", "showScannerView", "startCamera", "Companion", "app_debug"})
public final class PairingActivity extends androidx.appcompat.app.AppCompatActivity {
    @org.jetbrains.annotations.NotNull()
    private static final java.lang.String TAG = "PairingActivity";
    @javax.inject.Inject()
    public com.mobileproxy.core.config.CredentialManager credentialManager;
    @org.jetbrains.annotations.NotNull()
    private final kotlinx.coroutines.CoroutineScope scope = null;
    private boolean isPairing = false;
    @org.jetbrains.annotations.Nullable()
    private androidx.camera.lifecycle.ProcessCameraProvider cameraProvider;
    @org.jetbrains.annotations.NotNull()
    private final androidx.activity.result.ActivityResultLauncher<java.lang.String> cameraPermissionLauncher = null;
    @org.jetbrains.annotations.NotNull()
    public static final com.mobileproxy.ui.PairingActivity.Companion Companion = null;
    
    public PairingActivity() {
        super();
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
    
    private final void showScannerView() {
    }
    
    private final void showManualEntry() {
    }
    
    @androidx.annotation.OptIn(markerClass = {androidx.camera.core.ExperimentalGetImage.class})
    private final void startCamera() {
    }
    
    private final void handleQrCode(java.lang.String qrValue) {
    }
    
    private final void pair(java.lang.String serverUrl, java.lang.String code) {
    }
    
    private final void showError(java.lang.String message) {
    }
    
    @java.lang.Override()
    protected void onDestroy() {
    }
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000\u0012\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\u000e\n\u0000\b\u0086\u0003\u0018\u00002\u00020\u0001B\u0007\b\u0002\u00a2\u0006\u0002\u0010\u0002R\u000e\u0010\u0003\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000\u00a8\u0006\u0005"}, d2 = {"Lcom/mobileproxy/ui/PairingActivity$Companion;", "", "()V", "TAG", "", "app_debug"})
    public static final class Companion {
        
        private Companion() {
            super();
        }
    }
}