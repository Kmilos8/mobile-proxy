'use client'

import { useEffect, useState, useCallback } from 'react'
import Link from 'next/link'
import { Smartphone, LinkIcon, ArrowDownUp, Calendar, Activity, ChevronRight } from 'lucide-react'
import { api, Device, OverviewStats } from '@/lib/api'
import { getToken } from '@/lib/auth'
import { addWSHandler } from '@/lib/websocket'
import { formatBytes, timeAgo, cn } from '@/lib/utils'
import StatCard from '@/components/ui/StatCard'
import StatusBadge from '@/components/ui/StatusBadge'

export default function OverviewPage() {
  const [stats, setStats] = useState<OverviewStats | null>(null)
  const [devices, setDevices] = useState<Device[]>([])
  const [loading, setLoading] = useState(true)

  const fetchData = useCallback(async () => {
    const token = getToken()
    if (!token) return
    try {
      const [statsRes, devRes] = await Promise.all([
        api.stats.overview(token),
        api.devices.list(token),
      ])
      setStats(statsRes)
      setDevices(devRes.devices || [])
    } catch (err) {
      console.error('Failed to fetch overview:', err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchData()
    const unsub = addWSHandler((msg) => {
      if (msg.type === 'device_update') {
        const updated = msg.payload as Device
        setDevices(prev => prev.map(d => d.id === updated.id ? updated : d))
      }
    })
    const interval = setInterval(fetchData, 30000)
    return () => { unsub(); clearInterval(interval) }
  }, [fetchData])

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-zinc-500">Loading overview...</div>
      </div>
    )
  }

  const onlineDevices = devices.filter(d => d.status === 'online')
  const offlineDevices = devices.filter(d => d.status === 'offline')

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Overview</h1>

      {/* Stat Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard
          title="Devices"
          value={`${stats?.devices_online ?? 0} / ${stats?.devices_total ?? 0}`}
          subtitle="online / total"
          icon={Smartphone}
        />
        <StatCard
          title="Active Connections"
          value={String(stats?.connections_active ?? 0)}
          subtitle="proxy connections"
          icon={LinkIcon}
        />
        <StatCard
          title="Bandwidth Today"
          value={formatBytes((stats?.bandwidth_today_in ?? 0) + (stats?.bandwidth_today_out ?? 0))}
          subtitle={`${formatBytes(stats?.bandwidth_today_in ?? 0)} in / ${formatBytes(stats?.bandwidth_today_out ?? 0)} out`}
          icon={ArrowDownUp}
        />
        <StatCard
          title="Bandwidth Month"
          value={formatBytes((stats?.bandwidth_month_in ?? 0) + (stats?.bandwidth_month_out ?? 0))}
          subtitle={`${formatBytes(stats?.bandwidth_month_in ?? 0)} in / ${formatBytes(stats?.bandwidth_month_out ?? 0)} out`}
          icon={Calendar}
        />
      </div>

      {/* Online Devices */}
      <div className="flex items-center justify-between mb-3">
        <h2 className="text-lg font-semibold flex items-center gap-2">
          <Activity className="w-4 h-4 text-green-400" />
          Online Devices
        </h2>
        <Link href="/connections" className="text-sm text-brand-400 hover:text-brand-300 flex items-center gap-1">
          View all <ChevronRight className="w-3 h-3" />
        </Link>
      </div>

      {onlineDevices.length === 0 ? (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 text-center text-zinc-500 mb-6">
          No devices online
        </div>
      ) : (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden mb-6">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-zinc-800 text-zinc-500 text-left">
                <th className="px-4 py-2.5 font-medium text-xs">Device</th>
                <th className="px-4 py-2.5 font-medium text-xs">Carrier</th>
                <th className="px-4 py-2.5 font-medium text-xs">External IP</th>
                <th className="px-4 py-2.5 font-medium text-xs">HTTP / SOCKS5</th>
                <th className="px-4 py-2.5 font-medium text-xs">Battery</th>
                <th className="px-4 py-2.5 font-medium text-xs">Last Seen</th>
              </tr>
            </thead>
            <tbody>
              {onlineDevices.map(device => (
                <tr key={device.id} className="border-b border-zinc-800/30 hover:bg-zinc-800/20">
                  <td className="px-4 py-2.5">
                    <Link href={`/connections/${device.id}`} className="text-brand-400 hover:text-brand-300 font-medium">
                      {device.name}
                    </Link>
                    <div className="text-xs text-zinc-600">{device.device_model}</div>
                  </td>
                  <td className="px-4 py-2.5 text-zinc-300">{device.carrier || '-'}</td>
                  <td className="px-4 py-2.5 font-mono text-xs text-zinc-300">{device.cellular_ip || '-'}</td>
                  <td className="px-4 py-2.5 font-mono text-xs text-zinc-400">{device.http_port} / {device.socks5_port}</td>
                  <td className="px-4 py-2.5">
                    <div className="flex items-center gap-1.5">
                      <div className={cn(
                        'w-5 h-2.5 rounded-full',
                        device.battery_level > 50 ? 'bg-green-500' :
                        device.battery_level > 20 ? 'bg-yellow-500' : 'bg-red-500'
                      )} style={{ width: `${Math.max(device.battery_level / 4, 4)}px` }} />
                      <span className="text-xs text-zinc-400">{device.battery_level}%</span>
                    </div>
                  </td>
                  <td className="px-4 py-2.5 text-xs text-zinc-500">{timeAgo(device.last_heartbeat)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Offline Devices */}
      {offlineDevices.length > 0 && (
        <>
          <h2 className="text-lg font-semibold flex items-center gap-2 mb-3 text-zinc-400">
            Offline Devices
            <span className="text-sm font-normal text-zinc-600">({offlineDevices.length})</span>
          </h2>
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-zinc-800 text-zinc-500 text-left">
                  <th className="px-4 py-2.5 font-medium text-xs">Device</th>
                  <th className="px-4 py-2.5 font-medium text-xs">Last IP</th>
                  <th className="px-4 py-2.5 font-medium text-xs">Last Seen</th>
                </tr>
              </thead>
              <tbody>
                {offlineDevices.map(device => (
                  <tr key={device.id} className="border-b border-zinc-800/30 hover:bg-zinc-800/20">
                    <td className="px-4 py-2.5">
                      <Link href={`/connections/${device.id}`} className="text-zinc-400 hover:text-zinc-300">
                        {device.name}
                      </Link>
                      <div className="text-xs text-zinc-600">{device.device_model}</div>
                    </td>
                    <td className="px-4 py-2.5 font-mono text-xs text-zinc-500">{device.cellular_ip || '-'}</td>
                    <td className="px-4 py-2.5 text-xs text-zinc-600">{timeAgo(device.last_heartbeat)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  )
}
