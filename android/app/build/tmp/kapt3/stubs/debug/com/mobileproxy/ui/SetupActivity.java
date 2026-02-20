package com.mobileproxy.ui;

@dagger.hilt.android.AndroidEntryPoint()
@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000L\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0002\u0010\u000e\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\u0002\n\u0000\n\u0002\u0010\b\n\u0000\n\u0002\u0010\u000b\n\u0002\b\u0003\n\u0002\u0018\u0002\n\u0002\b\u0005\b\u0007\u0018\u0000 \u001c2\u00020\u0001:\u0001\u001cB\u0005\u00a2\u0006\u0002\u0010\u0002J\u0010\u0010\u000f\u001a\u00020\u00102\u0006\u0010\u0011\u001a\u00020\u0012H\u0002J\b\u0010\u0013\u001a\u00020\u0014H\u0002J\b\u0010\u0015\u001a\u00020\u0010H\u0002J\u0012\u0010\u0016\u001a\u00020\u00102\b\u0010\u0017\u001a\u0004\u0018\u00010\u0018H\u0014J\b\u0010\u0019\u001a\u00020\u0010H\u0014J\b\u0010\u001a\u001a\u00020\u0010H\u0002J\b\u0010\u001b\u001a\u00020\u0010H\u0002R\u001c\u0010\u0003\u001a\u0010\u0012\f\u0012\n \u0006*\u0004\u0018\u00010\u00050\u00050\u0004X\u0082\u0004\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0007\u001a\u00020\bX\u0082.\u00a2\u0006\u0002\n\u0000R\u000e\u0010\t\u001a\u00020\nX\u0082.\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u000b\u001a\u00020\fX\u0082.\u00a2\u0006\u0002\n\u0000R\u001c\u0010\r\u001a\u0010\u0012\f\u0012\n \u0006*\u0004\u0018\u00010\u000e0\u000e0\u0004X\u0082\u0004\u00a2\u0006\u0002\n\u0000\u00a8\u0006\u001d"}, d2 = {"Lcom/mobileproxy/ui/SetupActivity;", "Landroidx/appcompat/app/AppCompatActivity;", "()V", "notificationPermissionLauncher", "Landroidx/activity/result/ActivityResultLauncher;", "", "kotlin.jvm.PlatformType", "progressBar", "Landroid/widget/ProgressBar;", "stepIndicator", "Landroid/widget/TextView;", "viewFlipper", "Landroid/widget/ViewFlipper;", "vpnPermissionLauncher", "Landroid/content/Intent;", "goToStep", "", "step", "", "isDigitalAssistant", "", "markSetupComplete", "onCreate", "savedInstanceState", "Landroid/os/Bundle;", "onResume", "openDigitalAssistantSettings", "updateCurrentStepStatus", "Companion", "app_debug"})
public final class SetupActivity extends androidx.appcompat.app.AppCompatActivity {
    @org.jetbrains.annotations.NotNull()
    private static final java.lang.String PREFS_NAME = "setup_prefs";
    @org.jetbrains.annotations.NotNull()
    private static final java.lang.String KEY_SETUP_COMPLETE = "setup_complete";
    private android.widget.ViewFlipper viewFlipper;
    private android.widget.TextView stepIndicator;
    private android.widget.ProgressBar progressBar;
    @org.jetbrains.annotations.NotNull()
    private final androidx.activity.result.ActivityResultLauncher<java.lang.String> notificationPermissionLauncher = null;
    @org.jetbrains.annotations.NotNull()
    private final androidx.activity.result.ActivityResultLauncher<android.content.Intent> vpnPermissionLauncher = null;
    @org.jetbrains.annotations.NotNull()
    public static final com.mobileproxy.ui.SetupActivity.Companion Companion = null;
    
    public SetupActivity() {
        super();
    }
    
    @java.lang.Override()
    protected void onCreate(@org.jetbrains.annotations.Nullable()
    android.os.Bundle savedInstanceState) {
    }
    
    @java.lang.Override()
    protected void onResume() {
    }
    
    private final void goToStep(int step) {
    }
    
    private final void updateCurrentStepStatus() {
    }
    
    private final boolean isDigitalAssistant() {
        return false;
    }
    
    private final void openDigitalAssistantSettings() {
    }
    
    private final void markSetupComplete() {
    }
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000&\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\u000e\n\u0002\b\u0002\n\u0002\u0010\u000b\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\u0002\n\u0000\b\u0086\u0003\u0018\u00002\u00020\u0001B\u0007\b\u0002\u00a2\u0006\u0002\u0010\u0002J\u000e\u0010\u0006\u001a\u00020\u00072\u0006\u0010\b\u001a\u00020\tJ\u000e\u0010\n\u001a\u00020\u000b2\u0006\u0010\b\u001a\u00020\tR\u000e\u0010\u0003\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u000e\u0010\u0005\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000\u00a8\u0006\f"}, d2 = {"Lcom/mobileproxy/ui/SetupActivity$Companion;", "", "()V", "KEY_SETUP_COMPLETE", "", "PREFS_NAME", "isSetupComplete", "", "context", "Landroid/content/Context;", "resetSetup", "", "app_debug"})
    public static final class Companion {
        
        private Companion() {
            super();
        }
        
        public final boolean isSetupComplete(@org.jetbrains.annotations.NotNull()
        android.content.Context context) {
            return false;
        }
        
        public final void resetSetup(@org.jetbrains.annotations.NotNull()
        android.content.Context context) {
        }
    }
}