package com.mobileproxy.core.commands;

@javax.inject.Singleton()
@kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000$\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0002\n\u0002\u0018\u0002\n\u0002\u0010\u000e\n\u0000\n\u0002\u0018\u0002\n\u0002\b\u0004\b\u0007\u0018\u0000 \f2\u00020\u0001:\u0001\fB\u000f\b\u0007\u0012\u0006\u0010\u0002\u001a\u00020\u0003\u00a2\u0006\u0002\u0010\u0004J$\u0010\u0005\u001a\b\u0012\u0004\u0012\u00020\u00070\u00062\u0006\u0010\b\u001a\u00020\tH\u0086@\u00f8\u0001\u0000\u00f8\u0001\u0001\u00a2\u0006\u0004\b\n\u0010\u000bR\u000e\u0010\u0002\u001a\u00020\u0003X\u0082\u0004\u00a2\u0006\u0002\n\u0000\u0082\u0002\u000b\n\u0002\b!\n\u0005\b\u00a1\u001e0\u0001\u00a8\u0006\r"}, d2 = {"Lcom/mobileproxy/core/commands/CommandExecutor;", "", "rotationManager", "Lcom/mobileproxy/core/rotation/IPRotationManager;", "(Lcom/mobileproxy/core/rotation/IPRotationManager;)V", "execute", "Lkotlin/Result;", "", "command", "Lcom/mobileproxy/core/commands/DeviceCommand;", "execute-gIAlu-s", "(Lcom/mobileproxy/core/commands/DeviceCommand;Lkotlin/coroutines/Continuation;)Ljava/lang/Object;", "Companion", "app_debug"})
public final class CommandExecutor {
    @org.jetbrains.annotations.NotNull()
    private final com.mobileproxy.core.rotation.IPRotationManager rotationManager = null;
    @org.jetbrains.annotations.NotNull()
    private static final java.lang.String TAG = "CommandExecutor";
    @org.jetbrains.annotations.NotNull()
    public static final com.mobileproxy.core.commands.CommandExecutor.Companion Companion = null;
    
    @javax.inject.Inject()
    public CommandExecutor(@org.jetbrains.annotations.NotNull()
    com.mobileproxy.core.rotation.IPRotationManager rotationManager) {
        super();
    }
    
    @kotlin.Metadata(mv = {1, 9, 0}, k = 1, xi = 48, d1 = {"\u0000\u0012\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0002\b\u0002\n\u0002\u0010\u000e\n\u0000\b\u0086\u0003\u0018\u00002\u00020\u0001B\u0007\b\u0002\u00a2\u0006\u0002\u0010\u0002R\u000e\u0010\u0003\u001a\u00020\u0004X\u0082T\u00a2\u0006\u0002\n\u0000\u00a8\u0006\u0005"}, d2 = {"Lcom/mobileproxy/core/commands/CommandExecutor$Companion;", "", "()V", "TAG", "", "app_debug"})
    public static final class Companion {
        
        private Companion() {
            super();
        }
    }
}