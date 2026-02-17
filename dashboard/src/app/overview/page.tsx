'use client'

import { useEffect, useState, useCallback } from 'react'
import Link from 'next/link'
import { Smartphone, LinkIcon, ArrowDownUp, Calendar } from 'lucide-react'
import { api, Device, OverviewStats } from '@/lib/api'
import { getToken } from '@/lib/auth'
import { addWSHandler } from '@/lib/websocket'
import { formatBytes, timeAgo } from '@/lib/utils'
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
    return <div className="text-zinc-500">Loading overview...</div>
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Overview</h1>

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

      <h2 className="text-lg font-semibold mb-3">Device Status</h2>
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-zinc-800 text-zinc-400 text-left">
              <th className="px-4 py-3 font-medium">Name</th>
              <th className="px-4 py-3 font-medium">Status</th>
              <th className="px-4 py-3 font-medium">Cellular IP</th>
              <th className="px-4 py-3 font-medium">Last Seen</th>
            </tr>
          </thead>
          <tbody>
            {devices.map(device => (
              <tr key={device.id} className="border-b border-zinc-800/50 hover:bg-zinc-800/30">
                <td className="px-4 py-3">
                  <Link href={`/devices/${device.id}`} className="text-blue-400 hover:text-blue-300">
                    {device.name}
                  </Link>
                  <div className="text-xs text-zinc-500">{device.device_model}</div>
                </td>
                <td className="px-4 py-3">
                  <StatusBadge status={device.status} />
                </td>
                <td className="px-4 py-3 font-mono text-xs">{device.cellular_ip || '-'}</td>
                <td className="px-4 py-3 text-zinc-400">{timeAgo(device.last_heartbeat)}</td>
              </tr>
            ))}
            {devices.length === 0 && (
              <tr>
                <td colSpan={4} className="px-4 py-8 text-center text-zinc-500">
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
