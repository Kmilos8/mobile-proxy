package com.mobileproxy.ui;

import com.mobileproxy.core.config.CredentialManager;
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
public final class PairingActivity_MembersInjector implements MembersInjector<PairingActivity> {
  private final Provider<CredentialManager> credentialManagerProvider;

  public PairingActivity_MembersInjector(Provider<CredentialManager> credentialManagerProvider) {
    this.credentialManagerProvider = credentialManagerProvider;
  }

  public static MembersInjector<PairingActivity> create(
      Provider<CredentialManager> credentialManagerProvider) {
    return new PairingActivity_MembersInjector(credentialManagerProvider);
  }

  @Override
  public void injectMembers(PairingActivity instance) {
    injectCredentialManager(instance, credentialManagerProvider.get());
  }

  @InjectedFieldSignature("com.mobileproxy.ui.PairingActivity.credentialManager")
  public static void injectCredentialManager(PairingActivity instance,
      CredentialManager credentialManager) {
    instance.credentialManager = credentialManager;
  }
}
