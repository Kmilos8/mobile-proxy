package com.mobileproxy.core.rotation;

import android.content.Context;
import com.mobileproxy.core.network.NetworkManager;
import com.mobileproxy.core.status.DeviceStatusReporter;
import dagger.Lazy;
import dagger.internal.DaggerGenerated;
import dagger.internal.DoubleCheck;
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
public final class IPRotationManager_Factory implements Factory<IPRotationManager> {
  private final Provider<Context> contextProvider;

  private final Provider<NetworkManager> networkManagerProvider;

  private final Provider<DeviceStatusReporter> statusReporterProvider;

  public IPRotationManager_Factory(Provider<Context> contextProvider,
      Provider<NetworkManager> networkManagerProvider,
      Provider<DeviceStatusReporter> statusReporterProvider) {
    this.contextProvider = contextProvider;
    this.networkManagerProvider = networkManagerProvider;
    this.statusReporterProvider = statusReporterProvider;
  }

  @Override
  public IPRotationManager get() {
    return newInstance(contextProvider.get(), networkManagerProvider.get(), DoubleCheck.lazy(statusReporterProvider));
  }

  public static IPRotationManager_Factory create(Provider<Context> contextProvider,
      Provider<NetworkManager> networkManagerProvider,
      Provider<DeviceStatusReporter> statusReporterProvider) {
    return new IPRotationManager_Factory(contextProvider, networkManagerProvider, statusReporterProvider);
  }

  public static IPRotationManager newInstance(Context context, NetworkManager networkManager,
      Lazy<DeviceStatusReporter> statusReporter) {
    return new IPRotationManager(context, networkManager, statusReporter);
  }
}
