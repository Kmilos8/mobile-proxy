package com.mobileproxy.core.commands;

import com.mobileproxy.core.rotation.IPRotationManager;
import dagger.internal.DaggerGenerated;
import dagger.internal.Factory;
import dagger.internal.QualifierMetadata;
import dagger.internal.ScopeMetadata;
import javax.annotation.processing.Generated;
import javax.inject.Provider;

@ScopeMetadata("javax.inject.Singleton")
@QualifierMetadata
@DaggerGenerated
@Generated(
    value = "dagger.internal.codegen.ComponentProcessor",
    comments = "https://dagger.dev"
)
@SuppressWarnings({
    "unchecked",
    "rawtypes",
    "KotlinInternal",
    "KotlinInternalInJava"
})
public final class CommandExecutor_Factory implements Factory<CommandExecutor> {
  private final Provider<IPRotationManager> rotationManagerProvider;

  public CommandExecutor_Factory(Provider<IPRotationManager> rotationManagerProvider) {
    this.rotationManagerProvider = rotationManagerProvider;
  }

  @Override
  public CommandExecutor get() {
    return newInstance(rotationManagerProvider.get());
  }

  public static CommandExecutor_Factory create(
      Provider<IPRotationManager> rotationManagerProvider) {
    return new CommandExecutor_Factory(rotationManagerProvider);
  }

  public static CommandExecutor newInstance(IPRotationManager rotationManager) {
    return new CommandExecutor(rotationManager);
  }
}
