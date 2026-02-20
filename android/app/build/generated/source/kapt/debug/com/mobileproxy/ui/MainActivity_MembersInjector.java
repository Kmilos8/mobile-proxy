package com.mobileproxy.ui;

import com.mobileproxy.core.config.CredentialManager;
import com.mobileproxy.core.network.NetworkManager;
import com.mobileproxy.core.rotation.IPRotationManager;
import dagger.MembersInjector;
import dagger.internal.DaggerGenerated;
import dagger.internal.InjectedFieldSignature;
import dagger.internal.QualifierMetadata;
import javax.annotation.processing.Generated;
import javax.inject.Provider;

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
public final class MainActivity_MembersInjector implements MembersInjector<MainActivity> {
  private final Provider<NetworkManager> networkManagerProvider;

  private final Provider<IPRotationManager> rotationManagerProvider;

  private final Provider<CredentialManager> credentialManagerProvider;

  public MainActivity_MembersInjector(Provider<NetworkManager> networkManagerProvider,
      Provider<IPRotationManager> rotationManagerProvider,
      Provider<CredentialManager> credentialManagerProvider) {
    this.networkManagerProvider = networkManagerProvider;
    this.rotationManagerProvider = rotationManagerProvider;
    this.credentialManagerProvider = credentialManagerProvider;
  }

  public static MembersInjector<MainActivity> create(
      Provider<NetworkManager> networkManagerProvider,
      Provider<IPRotationManager> rotationManagerProvider,
      Provider<CredentialManager> credentialManagerProvider) {
    return new MainActivity_MembersInjector(networkManagerProvider, rotationManagerProvider, credentialManagerProvider);
  }

  @Override
  public void injectMembers(MainActivity instance) {
    injectNetworkManager(instance, networkManagerProvider.get());
    injectRotationManager(instance, rotationManagerProvider.get());
    injectCredentialManager(instance, credentialManagerProvider.get());
  }

  @InjectedFieldSignature("com.mobileproxy.ui.MainActivity.networkManager")
  public static void injectNetworkManager(MainActivity instance, NetworkManager networkManager) {
    instance.networkManager = networkManager;
  }

  @InjectedFieldSignature("com.mobileproxy.ui.MainActivity.rotationManager")
  public static void injectRotationManager(MainActivity instance,
      IPRotationManager rotationManager) {
    instance.rotationManager = rotationManager;
  }

  @InjectedFieldSignature("com.mobileproxy.ui.MainActivity.credentialManager")
  public static void injectCredentialManager(MainActivity instance,
      CredentialManager credentialManager) {
    instance.credentialManager = credentialManager;
  }
}
