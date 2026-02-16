'use client'

import { useEffect, useState, useCallback } from 'react'
import { api, Device } from '@/lib/api'
import { getToken } from '@/lib/auth'
import { addWSHandler } from '@/lib/websocket'
import { timeAgo, cn } from '@/lib/utils'

function StatusBadge({ status }: { status: Device['status'] }) {
  const colors = {
    online: 'bg-green-500/20 text-green-400 border-green-500/30',
    offline: 'bg-zinc-500/20 text-zinc-400 border-zinc-500/30',
    rotating: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
    error: 'bg-red-500/20 text-red-400 border-red-500/30',
  }
  return (
    <span className={cn('px-2 py-0.5 rounded-full text-xs border', colors[status])}>
      {status}
    </span>
  )
}

function BatteryIndicator({ level, charging }: { level: number; charging: boolean }) {
  const color = level > 50 ? 'text-green-400' : level > 20 ? 'text-yellow-400' : 'text-red-400'
  return (
    <span className={cn('text-sm', color)}>
      {level}%{charging ? ' +' : ''}
    </span>
  )
}

export default function DevicesPage() {
  const [devices, setDevices] = useState<Device[]>([])
  const [loading, setLoading] = useState(true)

  const fetchDevices = useCallback(async () => {
    const token = getToken()
    if (!token) return
    try {
      const res = await api.devices.list(token)
      setDevices(res.devices || [])
    } catch (err) {
      console.error('Failed to fetch devices:', err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchDevices()
    // Listen for real-time updates
    const unsub = addWSHandler((msg) => {
      if (msg.type === 'device_update') {
        const updated = msg.payload as Device
        setDevices(prev => prev.map(d => d.id === updated.id ? updated : d))
      }
    })
    // Poll as fallback
    const interval = setInterval(fetchDevices, 15000)
    return () => { unsub(); clearInterval(interval) }
  }, [fetchDevices])

  async function handleRotateIP(deviceId: string) {
    const token = getToken()
    if (!token) return
    try {
      await api.devices.sendCommand(token, deviceId, 'rotate_ip')
    } catch (err) {
      console.error('Failed to send rotate command:', err)
    }
  }

  if (loading) {
    return <div className="text-zinc-500">Loading devices...</div>
  }

  const onlineCount = devices.filter(d => d.status === 'online').length

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">Devices</h1>
          <p className="text-zinc-500 text-sm">{onlineCount} online / {devices.length} total</p>
        </div>
      </div>

      <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-zinc-800 text-zinc-400 text-left">
              <th className="px-4 py-3 font-medium">Name</th>
              <th className="px-4 py-3 font-medium">Status</th>
              <th className="px-4 py-3 font-medium">Cellular IP</th>
              <th className="px-4 py-3 font-medium">Carrier</th>
              <th className="px-4 py-3 font-medium">Network</th>
              <th className="px-4 py-3 font-medium">Battery</th>
              <th className="px-4 py-3 font-medium">Ports</th>
              <th className="px-4 py-3 font-medium">Last Seen</th>
              <th className="px-4 py-3 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {devices.map(device => (
              <tr key={device.id} className="border-b border-zinc-800/50 hover:bg-zinc-800/30">
                <td className="px-4 py-3">
                  <div className="font-medium">{device.name}</div>
                  <div className="text-xs text-zinc-500">{device.device_model}</div>
                </td>
                <td className="px-4 py-3">
                  <StatusBadge status={device.status} />
                </td>
                <td className="px-4 py-3 font-mono text-xs">{device.cellular_ip || '-'}</td>
                <td className="px-4 py-3">{device.carrier || '-'}</td>
                <td className="px-4 py-3">{device.network_type || '-'}</td>
                <td className="px-4 py-3">
                  <BatteryIndicator level={device.battery_level} charging={device.battery_charging} />
                </td>
                <td className="px-4 py-3 text-xs text-zinc-400">
                  HTTP:{device.http_port} SOCKS:{device.socks5_port}
                </td>
                <td className="px-4 py-3 text-zinc-400">
                  {timeAgo(device.last_heartbeat)}
                </td>
                <td className="px-4 py-3">
                  <button
                    onClick={() => handleRotateIP(device.id)}
                    disabled={device.status !== 'online'}
                    className="px-2 py-1 text-xs bg-blue-600 hover:bg-blue-700 disabled:bg-zinc-700 disabled:text-zinc-500 text-white rounded"
                  >
                    Rotate IP
                  </button>
                </td>
              </tr>
            ))}
            {devices.length === 0 && (
              <tr>
                <td colSpan={9} className="px-4 py-8 text-center text-zinc-500">
                  No devices registered yet
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
