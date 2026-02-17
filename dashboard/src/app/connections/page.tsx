'use client'

import { useEffect, useState, FormEvent } from 'react'
import Link from 'next/link'
import { api, ProxyConnection, Device } from '@/lib/api'
import { getToken } from '@/lib/auth'
import { formatDate } from '@/lib/utils'
import StatusBadge from '@/components/ui/StatusBadge'
import BandwidthBar from '@/components/ui/BandwidthBar'

export default function ConnectionsPage() {
  const [connections, setConnections] = useState<ProxyConnection[]>([])
  const [devices, setDevices] = useState<Device[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)

  // Form state
  const [formDeviceId, setFormDeviceId] = useState('')
  const [formUsername, setFormUsername] = useState('')
  const [formPassword, setFormPassword] = useState('')
  const [createdConn, setCreatedConn] = useState<ProxyConnection | null>(null)

  useEffect(() => {
    fetchData()
  }, [])

  async function fetchData() {
    const token = getToken()
    if (!token) return
    try {
      const [connRes, devRes] = await Promise.all([
        api.connections.list(token),
        api.devices.list(token),
      ])
      setConnections(connRes.connections || [])
      setDevices(devRes.devices || [])
    } catch (err) {
      console.error('Failed to fetch:', err)
    } finally {
      setLoading(false)
    }
  }

  async function handleCreate(e: FormEvent) {
    e.preventDefault()
    const token = getToken()
    if (!token) return
    try {
      const conn = await api.connections.create(token, {
        device_id: formDeviceId,
        username: formUsername,
        password: formPassword,
      })
      setCreatedConn(conn)
      setConnections(prev => [conn, ...prev])
      setFormUsername('')
      setFormPassword('')
    } catch (err) {
      console.error('Failed to create:', err)
    }
  }

  async function handleToggle(id: string, active: boolean) {
    const token = getToken()
    if (!token) return
    try {
      await api.connections.setActive(token, id, !active)
      setConnections(prev => prev.map(c => c.id === id ? { ...c, active: !active } : c))
    } catch (err) {
      console.error('Failed to toggle:', err)
    }
  }

  async function handleDelete(id: string) {
    const token = getToken()
    if (!token) return
    if (!confirm('Delete this connection?')) return
    try {
      await api.connections.delete(token, id)
      setConnections(prev => prev.filter(c => c.id !== id))
    } catch (err) {
      console.error('Failed to delete:', err)
    }
  }

  function getDevice(deviceId: string): Device | undefined {
    return devices.find(d => d.id === deviceId)
  }

  if (loading) return <div className="text-zinc-500">Loading connections...</div>

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Proxy Connections</h1>
        <button
          onClick={() => setShowCreate(!showCreate)}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded text-sm"
        >
          {showCreate ? 'Cancel' : 'New Connection'}
        </button>
      </div>

      {showCreate && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 mb-6">
          <form onSubmit={handleCreate} className="flex gap-4 items-end">
            <div className="flex-1">
              <label className="block text-sm text-zinc-400 mb-1">Device</label>
              <select
                value={formDeviceId}
                onChange={e => setFormDeviceId(e.target.value)}
                className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white"
                required
              >
                <option value="">Select device</option>
                {devices.filter(d => d.status === 'online').map(d => (
                  <option key={d.id} value={d.id}>{d.name} ({d.cellular_ip})</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm text-zinc-400 mb-1">Username</label>
              <input
                value={formUsername}
                onChange={e => setFormUsername(e.target.value)}
                className="px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white"
                required
              />
            </div>
            <div>
              <label className="block text-sm text-zinc-400 mb-1">Password</label>
              <input
                value={formPassword}
                onChange={e => setFormPassword(e.target.value)}
                className="px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white"
                required
              />
            </div>
            <button type="submit" className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded text-sm">
              Create
            </button>
          </form>
          {createdConn && (
            <div className="mt-4 bg-green-900/30 border border-green-800 rounded p-3 text-sm">
              <p className="font-medium text-green-400">Connection created!</p>
              <p className="text-zinc-300 mt-1 font-mono">
                Proxy: relay-server:{devices.find(d => d.id === createdConn.device_id)?.http_port || 'N/A'}
              </p>
              <p className="text-zinc-300 font-mono">
                Auth: {createdConn.username}:{createdConn.password}
              </p>
            </div>
          )}
        </div>
      )}

      <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-zinc-800 text-zinc-400 text-left">
              <th className="px-4 py-3 font-medium">Device</th>
              <th className="px-4 py-3 font-medium">Username</th>
              <th className="px-4 py-3 font-medium">Bandwidth</th>
              <th className="px-4 py-3 font-medium">Status</th>
              <th className="px-4 py-3 font-medium">Created</th>
              <th className="px-4 py-3 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {connections.map(conn => {
              const device = getDevice(conn.device_id)
              return (
                <tr key={conn.id} className="border-b border-zinc-800/50 hover:bg-zinc-800/30">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <Link href={`/devices/${conn.device_id}`} className="text-blue-400 hover:text-blue-300">
                        {device?.name || conn.device_id.slice(0, 8)}
                      </Link>
                      {device && <StatusBadge status={device.status} />}
                    </div>
                    {device?.cellular_ip && (
                      <div className="text-xs text-zinc-500 font-mono mt-0.5">{device.cellular_ip}</div>
                    )}
                  </td>
                  <td className="px-4 py-3 font-mono text-xs">{conn.username}</td>
                  <td className="px-4 py-3 w-48">
                    <BandwidthBar used={conn.bandwidth_used} limit={conn.bandwidth_limit} />
                  </td>
                  <td className="px-4 py-3">
                    <span className={conn.active ? 'text-green-400' : 'text-zinc-500'}>
                      {conn.active ? 'Active' : 'Disabled'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-zinc-400 text-xs">{formatDate(conn.created_at)}</td>
                  <td className="px-4 py-3 space-x-2">
                    <button
                      onClick={() => handleToggle(conn.id, conn.active)}
                      className="px-2 py-1 text-xs bg-zinc-700 hover:bg-zinc-600 text-white rounded"
                    >
                      {conn.active ? 'Disable' : 'Enable'}
                    </button>
                    <button
                      onClick={() => handleDelete(conn.id)}
                      className="px-2 py-1 text-xs bg-red-700 hover:bg-red-600 text-white rounded"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              )
            })}
            {connections.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-zinc-500">
                  No proxy connections yet
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
