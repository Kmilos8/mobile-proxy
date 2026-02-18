'use client'

import { useEffect, useState, useCallback } from 'react'
import Link from 'next/link'
import { RotateCw, ChevronRight, Signal, Wifi } from 'lucide-react'
import { api, Device } from '@/lib/api'
import { getToken } from '@/lib/auth'
import { addWSHandler } from '@/lib/websocket'
import { timeAgo, cn } from '@/lib/utils'
import StatusBadge from '@/components/ui/StatusBadge'
import BatteryIndicator from '@/components/ui/BatteryIndicator'

const SERVER_HOST = process.env.NEXT_PUBLIC_SERVER_HOST || '178.156.240.184'

export default function DevicesPage() {
  const [devices, setDevices] = useState<Device[]>([])
  const [loading, setLoading] = useState(true)
  const [rotatingId, setRotatingId] = useState<string | null>(null)

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
    const unsub = addWSHandler((msg) => {
      if (msg.type === 'device_update') {
        const updated = msg.payload as Device
        setDevices(prev => prev.map(d => d.id === updated.id ? updated : d))
      }
    })
    const interval = setInterval(fetchDevices, 15000)
    return () => { unsub(); clearInterval(interval) }
  }, [fetchDevices])

  async function handleRotateIP(e: React.MouseEvent, deviceId: string) {
    e.preventDefault()
    e.stopPropagation()
    const token = getToken()
    if (!token) return
    setRotatingId(deviceId)
    try {
      await api.devices.sendCommand(token, deviceId, 'rotate_ip')
    } catch (err) {
      console.error('Failed to send rotate command:', err)
    } finally {
      setTimeout(() => setRotatingId(null), 2000)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-zinc-500">Loading devices...</div>
      </div>
    )
  }

  const onlineCount = devices.filter(d => d.status === 'online').length

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">Devices</h1>
          <p className="text-zinc-500 text-sm mt-1">
            <span className="text-green-400">{onlineCount} online</span> / {devices.length} total
          </p>
        </div>
      </div>

      {devices.length === 0 ? (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-12 text-center">
          <p className="text-zinc-500">No devices registered yet</p>
          <p className="text-zinc-600 text-sm mt-1">Install the MobileProxy app on an Android device to get started.</p>
        </div>
      ) : (
        <div className="space-y-2">
          {devices.map(device => (
            <Link
              key={device.id}
              href={`/devices/${device.id}`}
              className="block bg-zinc-900 border border-zinc-800 rounded-lg p-4 hover:border-zinc-700 transition-colors group"
            >
              <div className="flex items-center justify-between">
                {/* Left: Device info */}
                <div className="flex items-center gap-4 flex-1 min-w-0">
                  {/* Status indicator */}
                  <div className={cn(
                    'w-2.5 h-2.5 rounded-full flex-shrink-0',
                    device.status === 'online' ? 'bg-green-500' :
                    device.status === 'rotating' ? 'bg-yellow-500 animate-pulse' :
                    device.status === 'error' ? 'bg-red-500' : 'bg-zinc-600'
                  )} />

                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-white">{device.name || 'Unnamed Device'}</span>
                      <StatusBadge status={device.status} />
                    </div>
                    <div className="flex items-center gap-4 text-xs text-zinc-500 mt-1">
                      <span>{device.device_model}</span>
                      <span className="text-zinc-700">|</span>
                      <span>{device.carrier || 'No carrier'}</span>
                      <span className="text-zinc-700">|</span>
                      <span>{device.network_type || '-'}</span>
                    </div>
                  </div>
                </div>

                {/* Center: IPs and Ports */}
                <div className="hidden lg:flex items-center gap-8 text-sm mx-4">
                  <div className="text-right">
                    <div className="text-xs text-zinc-500">External IP</div>
                    <div className="font-mono text-xs text-zinc-300">{device.cellular_ip || '-'}</div>
                  </div>
                  <div className="text-right">
                    <div className="text-xs text-zinc-500">HTTP / SOCKS5</div>
                    <div className="font-mono text-xs text-zinc-300">{device.http_port} / {device.socks5_port}</div>
                  </div>
                  <div className="text-right">
                    <div className="text-xs text-zinc-500">Battery</div>
                    <div><BatteryIndicator level={device.battery_level} charging={device.battery_charging} /></div>
                  </div>
                  <div className="text-right w-16">
                    <div className="text-xs text-zinc-500">Last seen</div>
                    <div className="text-xs text-zinc-400">{timeAgo(device.last_heartbeat)}</div>
                  </div>
                </div>

                {/* Right: Actions */}
                <div className="flex items-center gap-2 ml-4">
                  <button
                    onClick={(e) => handleRotateIP(e, device.id)}
                    disabled={device.status !== 'online' || rotatingId === device.id}
                    className={cn(
                      'p-2 rounded-lg transition-colors',
                      device.status === 'online' && rotatingId !== device.id
                        ? 'text-blue-400 hover:bg-blue-600/20'
                        : 'text-zinc-600 cursor-not-allowed'
                    )}
                    title="Rotate IP"
                  >
                    <RotateCw className={cn('w-4 h-4', rotatingId === device.id && 'animate-spin')} />
                  </button>
                  <ChevronRight className="w-4 h-4 text-zinc-600 group-hover:text-zinc-400 transition-colors" />
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
