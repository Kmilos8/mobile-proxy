package com.mobileproxy.core.status;

import android.content.Context;
import com.mobileproxy.core.commands.CommandExecutor;
import com.mobileproxy.core.network.NetworkManager;
import com.mobileproxy.core.proxy.HttpProxyServer;
import com.mobileproxy.core.proxy.Socks5ProxyServer;
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
public final class DeviceStatusReporter_Factory implements Factory<DeviceStatusReporter> {
  private final Provider<Context> contextProvider;

  private final Provider<NetworkManager> networkManagerProvider;

  private final Provider<HttpProxyServer> httpProxyProvider;

  private final Provider<Socks5ProxyServer> socks5ProxyProvider;

  private final Provider<CommandExecutor> commandExecutorProvider;

  public DeviceStatusReporter_Factory(Provider<Context> contextProvider,
      Provider<NetworkManager> networkManagerProvider, Provider<HttpProxyServer> httpProxyProvider,
      Provider<Socks5ProxyServer> socks5ProxyProvider,
      Provider<CommandExecutor> commandExecutorProvider) {
    this.contextProvider = contextProvider;
    this.networkManagerProvider = networkManagerProvider;
    this.httpProxyProvider = httpProxyProvider;
    this.socks5ProxyProvider = socks5ProxyProvider;
    this.commandExecutorProvider = commandExecutorProvider;
  }

  @Override
  public DeviceStatusReporter get() {
    return newInstance(contextProvider.get(), networkManagerProvider.get(), httpProxyProvider.get(), socks5ProxyProvider.get(), commandExecutorProvider.get());
  }

  public static DeviceStatusReporter_Factory create(Provider<Context> contextProvider,
      Provider<NetworkManager> networkManagerProvider, Provider<HttpProxyServer> httpProxyProvider,
      Provider<Socks5ProxyServer> socks5ProxyProvider,
      Provider<CommandExecutor> commandExecutorProvider) {
    return new DeviceStatusReporter_Factory(contextProvider, networkManagerProvider, httpProxyProvider, socks5ProxyProvider, commandExecutorProvider);
  }

  public static DeviceStatusReporter newInstance(Context context, NetworkManager networkManager,
      HttpProxyServer httpProxy, Socks5ProxyServer socks5Proxy, CommandExecutor commandExecutor) {
    return new DeviceStatusReporter(context, networkManager, httpProxy, socks5Proxy, commandExecutor);
  }
}
