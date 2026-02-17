'use client'

import { useEffect, useState, useCallback } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft, ArrowDownUp, Calendar } from 'lucide-react'
import { api, Device, DeviceBandwidth, DeviceCommand, ProxyConnection, IPHistoryEntry } from '@/lib/api'
import { getToken } from '@/lib/auth'
import { addWSHandler } from '@/lib/websocket'
import { formatBytes, formatDate, timeAgo } from '@/lib/utils'
import StatusBadge from '@/components/ui/StatusBadge'
import BatteryIndicator from '@/components/ui/BatteryIndicator'
import StatCard from '@/components/ui/StatCard'
import BandwidthBar from '@/components/ui/BandwidthBar'

export default function DeviceDetailPage() {
  const params = useParams()
  const id = params.id as string

  const [device, setDevice] = useState<Device | null>(null)
  const [bandwidth, setBandwidth] = useState<DeviceBandwidth | null>(null)
  const [connections, setConnections] = useState<ProxyConnection[]>([])
  const [ipHistory, setIpHistory] = useState<IPHistoryEntry[]>([])
  const [commands, setCommands] = useState<DeviceCommand[]>([])
  const [loading, setLoading] = useState(true)

  const fetchData = useCallback(async () => {
    const token = getToken()
    if (!token) return
    try {
      const [deviceRes, bwRes, connRes, histRes, cmdRes] = await Promise.all([
        api.devices.get(token, id),
        api.devices.bandwidth(token, id),
        api.connections.list(token, id),
        api.devices.ipHistory(token, id),
        api.devices.commands(token, id),
      ])
      setDevice(deviceRes)
      setBandwidth(bwRes)
      setConnections(connRes.connections || [])
      setIpHistory(histRes.history || [])
      setCommands(cmdRes.commands || [])
    } catch (err) {
      console.error('Failed to fetch device detail:', err)
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    fetchData()
    const unsub = addWSHandler((msg) => {
      if (msg.type === 'device_update') {
        const updated = msg.payload as Device
        if (updated.id === id) {
          setDevice(updated)
        }
      }
    })
    return () => { unsub() }
  }, [fetchData, id])

  async function handleRotateIP() {
    const token = getToken()
    if (!token || !device) return
    try {
      await api.devices.sendCommand(token, device.id, 'rotate_ip')
    } catch (err) {
      console.error('Failed to send rotate command:', err)
    }
  }

  if (loading) {
    return <div className="text-zinc-500">Loading device...</div>
  }

  if (!device) {
    return <div className="text-zinc-500">Device not found</div>
  }

  return (
    <div>
      <Link href="/devices" className="text-sm text-zinc-400 hover:text-white flex items-center gap-1 mb-4">
        <ArrowLeft className="w-4 h-4" /> Back to Devices
      </Link>

      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">{device.name}</h1>
          <div className="flex items-center gap-2 mt-1">
            <StatusBadge status={device.status} />
            <span className="text-sm text-zinc-500">{device.device_model}</span>
          </div>
        </div>
        <button
          onClick={handleRotateIP}
          disabled={device.status !== 'online'}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-zinc-700 disabled:text-zinc-500 text-white rounded text-sm"
        >
          Rotate IP
        </button>
      </div>

      {/* Device Info Grid */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 mb-6">
        <h2 className="text-sm font-semibold text-zinc-400 mb-3">Device Info</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
          <div>
            <span className="text-zinc-500">Cellular IP</span>
            <div className="font-mono">{device.cellular_ip || '-'}</div>
          </div>
          <div>
            <span className="text-zinc-500">WiFi IP</span>
            <div className="font-mono">{device.wifi_ip || '-'}</div>
          </div>
          <div>
            <span className="text-zinc-500">VPN IP</span>
            <div className="font-mono">{device.vpn_ip || '-'}</div>
          </div>
          <div>
            <span className="text-zinc-500">Carrier</span>
            <div>{device.carrier || '-'}</div>
          </div>
          <div>
            <span className="text-zinc-500">Network</span>
            <div>{device.network_type || '-'}</div>
          </div>
          <div>
            <span className="text-zinc-500">Battery</span>
            <div><BatteryIndicator level={device.battery_level} charging={device.battery_charging} /></div>
          </div>
          <div>
            <span className="text-zinc-500">Signal</span>
            <div>{device.signal_strength} dBm</div>
          </div>
          <div>
            <span className="text-zinc-500">Ports</span>
            <div className="text-xs">HTTP:{device.http_port} SOCKS:{device.socks5_port}</div>
          </div>
          <div>
            <span className="text-zinc-500">App Version</span>
            <div>{device.app_version || '-'}</div>
          </div>
          <div>
            <span className="text-zinc-500">Android</span>
            <div>{device.android_version || '-'}</div>
          </div>
          <div>
            <span className="text-zinc-500">Last Seen</span>
            <div>{timeAgo(device.last_heartbeat)}</div>
          </div>
          <div>
            <span className="text-zinc-500">Registered</span>
            <div>{formatDate(device.created_at)}</div>
          </div>
        </div>
      </div>

      {/* Bandwidth */}
      {bandwidth && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
          <StatCard
            title="Bandwidth Today"
            value={formatBytes((bandwidth.today_in) + (bandwidth.today_out))}
            subtitle={`${formatBytes(bandwidth.today_in)} in / ${formatBytes(bandwidth.today_out)} out`}
            icon={ArrowDownUp}
          />
          <StatCard
            title="Bandwidth Month"
            value={formatBytes((bandwidth.month_in) + (bandwidth.month_out))}
            subtitle={`${formatBytes(bandwidth.month_in)} in / ${formatBytes(bandwidth.month_out)} out`}
            icon={Calendar}
          />
        </div>
      )}

      {/* Connections */}
      <h2 className="text-lg font-semibold mb-3">Connections</h2>
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden mb-6">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-zinc-800 text-zinc-400 text-left">
              <th className="px-4 py-3 font-medium">Username</th>
              <th className="px-4 py-3 font-medium">Bandwidth</th>
              <th className="px-4 py-3 font-medium">Status</th>
            </tr>
          </thead>
          <tbody>
            {connections.map(conn => (
              <tr key={conn.id} className="border-b border-zinc-800/50 hover:bg-zinc-800/30">
                <td className="px-4 py-3 font-mono text-xs">{conn.username}</td>
                <td className="px-4 py-3 w-48">
                  <BandwidthBar used={conn.bandwidth_used} limit={conn.bandwidth_limit} />
                </td>
                <td className="px-4 py-3">
                  <span className={conn.active ? 'text-green-400' : 'text-zinc-500'}>
                    {conn.active ? 'Active' : 'Disabled'}
                  </span>
                </td>
              </tr>
            ))}
            {connections.length === 0 && (
              <tr>
                <td colSpan={3} className="px-4 py-6 text-center text-zinc-500">
                  No connections for this device
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* IP History */}
      <h2 className="text-lg font-semibold mb-3">IP History</h2>
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden mb-6">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-zinc-800 text-zinc-400 text-left">
              <th className="px-4 py-3 font-medium">IP Address</th>
              <th className="px-4 py-3 font-medium">Method</th>
              <th className="px-4 py-3 font-medium">Time</th>
            </tr>
          </thead>
          <tbody>
            {ipHistory.map(entry => (
              <tr key={entry.id} className="border-b border-zinc-800/50 hover:bg-zinc-800/30">
                <td className="px-4 py-3 font-mono text-xs">{entry.ip}</td>
                <td className="px-4 py-3">
                  <span className="px-2 py-0.5 rounded text-xs bg-zinc-800 text-zinc-300">
                    {entry.method}
                  </span>
                </td>
                <td className="px-4 py-3 text-zinc-400">{formatDate(entry.created_at)}</td>
              </tr>
            ))}
            {ipHistory.length === 0 && (
              <tr>
                <td colSpan={3} className="px-4 py-6 text-center text-zinc-500">
                  No IP history yet
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Command History */}
      <h2 className="text-lg font-semibold mb-3">Command History</h2>
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-zinc-800 text-zinc-400 text-left">
              <th className="px-4 py-3 font-medium">Type</th>
              <th className="px-4 py-3 font-medium">Status</th>
              <th className="px-4 py-3 font-medium">Created</th>
              <th className="px-4 py-3 font-medium">Executed</th>
            </tr>
          </thead>
          <tbody>
            {commands.map(cmd => (
              <tr key={cmd.id} className="border-b border-zinc-800/50 hover:bg-zinc-800/30">
                <td className="px-4 py-3">
                  <span className="px-2 py-0.5 rounded text-xs bg-zinc-800 text-zinc-300">
                    {cmd.type}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <StatusBadge status={cmd.status} />
                </td>
                <td className="px-4 py-3 text-zinc-400">{formatDate(cmd.created_at)}</td>
                <td className="px-4 py-3 text-zinc-400">{cmd.executed_at ? formatDate(cmd.executed_at) : '-'}</td>
              </tr>
            ))}
            {commands.length === 0 && (
              <tr>
                <td colSpan={4} className="px-4 py-6 text-center text-zinc-500">
                  No commands sent yet
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
