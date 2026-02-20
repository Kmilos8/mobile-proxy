package com.mobileproxy.core.network;

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
public final class NetworkManager_Factory implements Factory<NetworkManager> {
  private final Provider<Context> contextProvider;

  public NetworkManager_Factory(Provider<Context> contextProvider) {
    this.contextProvider = contextProvider;
  }

  @Override
  public NetworkManager get() {
    return newInstance(contextProvider.get());
  }

  public static NetworkManager_Factory create(Provider<Context> contextProvider) {
    return new NetworkManager_Factory(contextProvider);
  }

  public static NetworkManager newInstance(Context context) {
    return new NetworkManager(context);
  }
}
