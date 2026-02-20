package com.mobileproxy.core.config;

import android.content.Context;
import dagger.internal.DaggerGenerated;
import dagger.internal.Factory;
import dagger.internal.QualifierMetadata;
import dagger.internal.ScopeMetadata;
import javax.annotation.processing.Generated;
import javax.inject.Provider;

@ScopeMetadata("javax.inject.Singleton")
@QualifierMetadata("dagger.hilt.android.qualifiers.ApplicationContext")
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
public final class CredentialManager_Factory implements Factory<CredentialManager> {
  private final Provider<Context> contextProvider;

  public CredentialManager_Factory(Provider<Context> contextProvider) {
    this.contextProvider = contextProvider;
  }

  @Override
  public CredentialManager get() {
    return newInstance(contextProvider.get());
  }

  public static CredentialManager_Factory create(Provider<Context> contextProvider) {
    return new CredentialManager_Factory(contextProvider);
  }

  public static CredentialManager newInstance(Context context) {
    return new CredentialManager(context);
  }
}
