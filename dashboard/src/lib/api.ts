const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api'

interface RequestOptions {
  method?: string
  body?: unknown
  token?: string
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, token } = options
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: 'Request failed' }))
    throw new Error(error.error || `HTTP ${res.status}`)
  }

  if (res.status === 204) return {} as T
  return res.json()
}

export interface Device {
  id: string
  name: string
  android_id: string
  status: 'online' | 'offline' | 'rotating' | 'error'
  cellular_ip: string
  wifi_ip: string
  vpn_ip: string
  carrier: string
  network_type: string
  battery_level: number
  battery_charging: boolean
  signal_strength: number
  base_port: number
  http_port: number
  socks5_port: number
  last_heartbeat: string | null
  app_version: string
  device_model: string
  android_version: string
  created_at: string
}

export interface ProxyConnection {
  id: string
  device_id: string
  customer_id: string | null
  username: string
  password?: string
  ip_whitelist: string[]
  bandwidth_limit: number
  bandwidth_used: number
  active: boolean
  expires_at: string | null
  created_at: string
}

export interface Customer {
  id: string
  name: string
  email: string
  active: boolean
  created_at: string
}

export interface IPHistoryEntry {
  id: string
  device_id: string
  ip: string
  method: string
  created_at: string
}

export interface DeviceCommand {
  id: string
  device_id: string
  type: string
  status: 'pending' | 'sent' | 'completed' | 'failed'
  payload: string
  result: string
  created_at: string
  executed_at: string | null
}

export interface DeviceBandwidth {
  today_in: number
  today_out: number
  month_in: number
  month_out: number
}

export interface RotationLink {
  id: string
  device_id: string
  token: string
  name: string
  created_at: string
  last_used_at: string | null
}

export interface PairingCode {
  id: string
  code: string
  claimed_by_device_id: string | null
  claimed_at: string | null
  expires_at: string
  created_by: string | null
  created_at: string
}

export interface BandwidthHourly {
  hour: number
  download_bytes: number
  upload_bytes: number
}

export interface UptimeSegment {
  status: string
  start_time: string
  end_time: string
}

export interface OverviewStats {
  devices_total: number
  devices_online: number
  connections_active: number
  bandwidth_today_in: number
  bandwidth_today_out: number
  bandwidth_month_in: number
  bandwidth_month_out: number
}

export const api = {
  auth: {
    login: (email: string, password: string) =>
      request<{ token: string; user: { id: string; email: string; name: string; role: string } }>(
        '/auth/login', { method: 'POST', body: { email, password } }
      ),
  },
  devices: {
    list: (token: string) =>
      request<{ devices: Device[] }>('/devices', { token }),
    get: (token: string, id: string) =>
      request<Device>(`/devices/${id}`, { token }),
    sendCommand: (token: string, id: string, type: string, payload?: string) =>
      request(`/devices/${id}/commands`, { method: 'POST', token, body: { type, payload: payload || '{}' } }),
    ipHistory: (token: string, id: string) =>
      request<{ history: IPHistoryEntry[] }>(`/devices/${id}/ip-history`, { token }),
    bandwidth: (token: string, id: string) =>
      request<DeviceBandwidth>(`/devices/${id}/bandwidth`, { token }),
    commands: (token: string, id: string) =>
      request<{ commands: DeviceCommand[] }>(`/devices/${id}/commands`, { token }),
    bandwidthHourly: (token: string, id: string, date: string) =>
      request<{ hourly: BandwidthHourly[] }>(`/devices/${id}/bandwidth/hourly?date=${date}`, { token }),
    uptime: (token: string, id: string, date: string) =>
      request<{ segments: UptimeSegment[] }>(`/devices/${id}/uptime?date=${date}`, { token }),
  },
  stats: {
    overview: (token: string) =>
      request<OverviewStats>('/stats/overview', { token }),
  },
  connections: {
    list: (token: string, deviceId?: string) =>
      request<{ connections: ProxyConnection[] }>(
        `/connections${deviceId ? `?device_id=${deviceId}` : ''}`, { token }
      ),
    create: (token: string, data: { device_id: string; username: string; password: string; ip_whitelist?: string[]; bandwidth_limit?: number }) =>
      request<ProxyConnection>('/connections', { method: 'POST', token, body: data }),
    setActive: (token: string, id: string, active: boolean) =>
      request(`/connections/${id}`, { method: 'PATCH', token, body: { active } }),
    delete: (token: string, id: string) =>
      request(`/connections/${id}`, { method: 'DELETE', token }),
  },
  customers: {
    list: (token: string) =>
      request<{ customers: Customer[] }>('/customers', { token }),
    create: (token: string, data: { name: string; email: string }) =>
      request<Customer>('/customers', { method: 'POST', token, body: data }),
  },
  rotationLinks: {
    list: (token: string, deviceId: string) =>
      request<{ links: RotationLink[] }>(`/rotation-links?device_id=${deviceId}`, { token }),
    create: (token: string, data: { device_id: string; name?: string }) =>
      request<RotationLink>('/rotation-links', { method: 'POST', token, body: data }),
    delete: (token: string, id: string) =>
      request(`/rotation-links/${id}`, { method: 'DELETE', token }),
  },
  pairingCodes: {
    list: (token: string) =>
      request<{ pairing_codes: PairingCode[] }>('/pairing-codes', { token }),
    create: (token: string, expiresInMinutes?: number) =>
      request<{ id: string; code: string; expires_at: string }>('/pairing-codes', {
        method: 'POST', token, body: { expires_in_minutes: expiresInMinutes || 5 }
      }),
    delete: (token: string, id: string) =>
      request(`/pairing-codes/${id}`, { method: 'DELETE', token }),
  },
}
