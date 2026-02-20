package com.mobileproxy.core.proxy;

import com.mobileproxy.core.network.NetworkManager;
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
public final class HttpProxyServer_Factory implements Factory<HttpProxyServer> {
  private final Provider<NetworkManager> networkManagerProvider;

  public HttpProxyServer_Factory(Provider<NetworkManager> networkManagerProvider) {
    this.networkManagerProvider = networkManagerProvider;
  }

  @Override
  public HttpProxyServer get() {
    return newInstance(networkManagerProvider.get());
  }

  public static HttpProxyServer_Factory create(Provider<NetworkManager> networkManagerProvider) {
    return new HttpProxyServer_Factory(networkManagerProvider);
  }

  public static HttpProxyServer newInstance(NetworkManager networkManager) {
    return new HttpProxyServer(networkManager);
  }
}
