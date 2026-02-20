package com.mobileproxy.service;

import com.mobileproxy.core.network.NetworkManager;
import com.mobileproxy.core.proxy.HttpProxyServer;
import com.mobileproxy.core.proxy.Socks5ProxyServer;
import com.mobileproxy.core.status.DeviceStatusReporter;
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
public final class ProxyForegroundService_MembersInjector implements MembersInjector<ProxyForegroundService> {
  private final Provider<NetworkManager> networkManagerProvider;

  private final Provider<HttpProxyServer> httpProxyProvider;

  private final Provider<Socks5ProxyServer> socks5ProxyProvider;

  private final Provider<DeviceStatusReporter> statusReporterProvider;

  public ProxyForegroundService_MembersInjector(Provider<NetworkManager> networkManagerProvider,
      Provider<HttpProxyServer> httpProxyProvider, Provider<Socks5ProxyServer> socks5ProxyProvider,
      Provider<DeviceStatusReporter> statusReporterProvider) {
    this.networkManagerProvider = networkManagerProvider;
    this.httpProxyProvider = httpProxyProvider;
    this.socks5ProxyProvider = socks5ProxyProvider;
    this.statusReporterProvider = statusReporterProvider;
  }

  public static MembersInjector<ProxyForegroundService> create(
      Provider<NetworkManager> networkManagerProvider, Provider<HttpProxyServer> httpProxyProvider,
      Provider<Socks5ProxyServer> socks5ProxyProvider,
      Provider<DeviceStatusReporter> statusReporterProvider) {
    return new ProxyForegroundService_MembersInjector(networkManagerProvider, httpProxyProvider, socks5ProxyProvider, statusReporterProvider);
  }

  @Override
  public void injectMembers(ProxyForegroundService instance) {
    injectNetworkManager(instance, networkManagerProvider.get());
    injectHttpProxy(instance, httpProxyProvider.get());
    injectSocks5Proxy(instance, socks5ProxyProvider.get());
    injectStatusReporter(instance, statusReporterProvider.get());
  }

  @InjectedFieldSignature("com.mobileproxy.service.ProxyForegroundService.networkManager")
  public static void injectNetworkManager(ProxyForegroundService instance,
      NetworkManager networkManager) {
    instance.networkManager = networkManager;
  }

  @InjectedFieldSignature("com.mobileproxy.service.ProxyForegroundService.httpProxy")
  public static void injectHttpProxy(ProxyForegroundService instance, HttpProxyServer httpProxy) {
    instance.httpProxy = httpProxy;
  }

  @InjectedFieldSignature("com.mobileproxy.service.ProxyForegroundService.socks5Proxy")
  public static void injectSocks5Proxy(ProxyForegroundService instance,
      Socks5ProxyServer socks5Proxy) {
    instance.socks5Proxy = socks5Proxy;
  }

  @InjectedFieldSignature("com.mobileproxy.service.ProxyForegroundService.statusReporter")
  public static void injectStatusReporter(ProxyForegroundService instance,
      DeviceStatusReporter statusReporter) {
    instance.statusReporter = statusReporter;
  }
}
