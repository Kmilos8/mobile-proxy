package com.mobileproxy.core.rotation;

@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u00006\n\u0002\u0018\u0002\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\u0002\n\u0000\n\u0002\u0010\u000b\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0002\b\u0005\n\u0002\u0018\u0002\n\u0000\n\u0002\u0010\b\n\u0002\b\u0003\u0018\u0000 \u00152\u00020\u0001:\u0001\u0015B\u0005\u00a2\u0006\u0002\u0010\u0002J\b\u0010\u0005\u001a\u00020\u0006H\u0002J\b\u0010\u0007\u001a\u00020\bH\u0002J\u0012\u0010\t\u001a\u00020\u00062\b\u0010\n\u001a\u0004\u0018\u00010\u000bH\u0016J\b\u0010\f\u001a\u00020\u0006H\u0016J\b\u0010\r\u001a\u00020\u0006H\u0016J\b\u0010\u000e\u001a\u00020\u0006H\u0014J\u0018\u0010\u000f\u001a\u00020\b2\u0006\u0010\u0010\u001a\u00020\u00112\u0006\u0010\u0012\u001a\u00020\u0013H\u0002J\b\u0010\u0014\u001a\u00020\u0006H\u0002R\u000e\u0010\u0003\u001a\u00020\u0004X\u0082\u0004\u00a2\u0006\u0002\n\u0000\u00a8\u0006\u0016"}, d2 = {"Lcom/mobileproxy/core/rotation/AirplaneModeAccessibilityService;", "Landroid/accessibilityservice/AccessibilityService;", "()V", "scope", "Lkotlinx/coroutines/CoroutineScope;", "closeQuickSettings", "", "findAndClickAirplane", "", "onAccessibilityEvent", "event", "Landroid/view/accessibility/AccessibilityEvent;", "onDestroy", "onInterrupt", "onServiceConnected", "searchAndClick", "node", "Landroid/view/accessibility/AccessibilityNodeInfo;", "depth", "", "startToggleSequence", "Companion", "app_debug"})
public final class AirplaneModeAccessibilityService extends android.accessibilityservice.AccessibilityService {
    @org.jetbrains.annotations.NotNull()
    private static final java.lang.String TAG = "AirplaneModeAS";
    @org.jetbrains.annotations.Nullable()
    private static com.mobileproxy.core.rotation.AirplaneModeAccessibilityService instance;
    @kotlin.jvm.Volatile()
    @org.jetbrains.annotations.Nullable()
    private static volatile kotlinx.coroutines.CompletableDeferred<java.lang.Boolean> pendingToggle;
    @org.jetbrains.annotations.NotNull()
    private final kotlinx.coroutines.CoroutineScope scope = null;
    @org.jetbrains.annotations.NotNull()
    public static final com.mobileproxy.core.rotation.AirplaneModeAccessibilityService.Companion Companion = null;
    
    public AirplaneModeAccessibilityService() {
        super();
    }
    
    @java.lang.Override()
    protected void onServiceConnected() {
    }
    
    @java.lang.Override()
    public void onAccessibilityEvent(@org.jetbrains.annotations.Nullable()
    android.view.accessibility.AccessibilityEvent event) {
    }
    
    @java.lang.Override()
    public void onInterrupt() {
    }
    
    @java.lang.Override()
    public void onDestroy() {
    }
    
    private final void startToggleSequence() {
    }
    
    private final void closeQuickSettings() {
    }
    
    private final boolean findAndClickAirplane() {
        return false;
    }
    
    private final boolean searchAndClick(android.view.accessibility.AccessibilityNodeInfo node, int depth) {
        return false;
    }
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000,\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\u000e\n\u0000\n\u0002\u0018\u0002\n\u0000\n\u0002\u0018\u0002\n\u0002\u0010\u000b\n\u0002\b\u0006\n\u0002\u0010\u0002\n\u0002\b\u0003\b\u0086\u0003\u0018\u00002\u00020\u0001B\u0007\b\u0002\u00a2\u0006\u0002\u0010\u0002J\u0006\u0010\u000e\u001a\u00020\tJ\u0006\u0010\u000f\u001a\u00020\u0010J\u000e\u0010\u0011\u001a\u00020\tH\u0086@\u00a2\u0006\u0002\u0010\u0012R\u000e\u0010\u0003\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000R\u0010\u0010\u0005\u001a\u0004\u0018\u00010\u0006X\u0082\u000e\u00a2\u0006\u0002\n\u0000R\"\u0010\u0007\u001a\n\u0012\u0004\u0012\u00020\t\u0018\u00010\bX\u0086\u000e\u00a2\u0006\u000e\n\u0000\u001a\u0004\b\n\u0010\u000b\"\u0004\b\f\u0010\r\u00a8\u0006\u0013"}, d2 = {"Lcom/mobileproxy/core/rotation/AirplaneModeAccessibilityService$Companion;", "", "()V", "TAG", "", "instance", "Lcom/mobileproxy/core/rotation/AirplaneModeAccessibilityService;", "pendingToggle", "Lkotlinx/coroutines/CompletableDeferred;", "", "getPendingToggle", "()Lkotlinx/coroutines/CompletableDeferred;", "setPendingToggle", "(Lkotlinx/coroutines/CompletableDeferred;)V", "isAvailable", "requestToggle", "", "requestToggleAndWait", "(Lkotlin/coroutines/Continuation;)Ljava/lang/Object;", "app_debug"})
    public static final class Companion {
        
        private Companion() {
            super();
        }
        
        @org.jetbrains.annotations.Nullable()
        public final kotlinx.coroutines.CompletableDeferred<java.lang.Boolean> getPendingToggle() {
            return null;
        }
        
        public final void setPendingToggle(@org.jetbrains.annotations.Nullable()
        kotlinx.coroutines.CompletableDeferred<java.lang.Boolean> p0) {
        }
        
        public final void requestToggle() {
        }
        
        @org.jetbrains.annotations.Nullable()
        public final java.lang.Object requestToggleAndWait(@org.jetbrains.annotations.NotNull()
        kotlin.coroutines.Continuation<? super java.lang.Boolean> $completion) {
            return null;
        }
        
        public final boolean isAvailable() {
            return false;
        }
    }
}