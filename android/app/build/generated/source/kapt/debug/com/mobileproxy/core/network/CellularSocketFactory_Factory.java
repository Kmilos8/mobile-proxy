package com.mobileproxy.core.network;

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
public final class CellularSocketFactory_Factory implements Factory<CellularSocketFactory> {
  private final Provider<NetworkManager> networkManagerProvider;

  public CellularSocketFactory_Factory(Provider<NetworkManager> networkManagerProvider) {
    this.networkManagerProvider = networkManagerProvider;
  }

  @Override
  public CellularSocketFactory get() {
    return newInstance(networkManagerProvider.get());
  }

  public static CellularSocketFactory_Factory create(
      Provider<NetworkManager> networkManagerProvider) {
    return new CellularSocketFactory_Factory(networkManagerProvider);
  }

  public static CellularSocketFactory newInstance(NetworkManager networkManager) {
    return new CellularSocketFactory(networkManager);
  }
}
