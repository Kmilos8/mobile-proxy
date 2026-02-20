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
public final class Socks5ProxyServer_Factory implements Factory<Socks5ProxyServer> {
  private final Provider<NetworkManager> networkManagerProvider;

  public Socks5ProxyServer_Factory(Provider<NetworkManager> networkManagerProvider) {
    this.networkManagerProvider = networkManagerProvider;
  }

  @Override
  public Socks5ProxyServer get() {
    return newInstance(networkManagerProvider.get());
  }

  public static Socks5ProxyServer_Factory create(Provider<NetworkManager> networkManagerProvider) {
    return new Socks5ProxyServer_Factory(networkManagerProvider);
  }

  public static Socks5ProxyServer newInstance(NetworkManager networkManager) {
    return new Socks5ProxyServer(networkManager);
  }
}
