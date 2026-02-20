'use client'

import { useEffect, useState, useCallback, FormEvent } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import {
  ArrowLeft, Smartphone, Settings, Link2, Clock, Activity,
  RotateCw, Power, Search, Wifi, WifiOff, Copy, Trash2, Plus,
  Battery, Signal, Globe, Cpu, RefreshCw, ChevronRight, BarChart3,
  Pencil, Check, X
} from 'lucide-react'
import { api, Device, DeviceBandwidth, DeviceCommand, ProxyConnection, IPHistoryEntry, RotationLink, BandwidthHourly, UptimeSegment } from '@/lib/api'
import { getToken } from '@/lib/auth'
import { addWSHandler } from '@/lib/websocket'
import { formatBytes, formatDate, timeAgo, cn } from '@/lib/utils'
import StatusBadge from '@/components/ui/StatusBadge'
import BatteryIndicator from '@/components/ui/BatteryIndicator'
import BandwidthChart from '@/components/BandwidthChart'
import UptimeTimeline from '@/components/UptimeTimeline'

type SidebarTab = 'primary' | 'advanced' | 'change-ip' | 'history' | 'metrics' | 'usage'

const SERVER_HOST = process.env.NEXT_PUBLIC_SERVER_HOST || '178.156.240.184'

export default function ConnectionDetailPage() {
  const params = useParams()
  const id = params.id as string

  const [device, setDevice] = useState<Device | null>(null)
  const [bandwidth, setBandwidth] = useState<DeviceBandwidth | null>(null)
  const [connections, setConnections] = useState<ProxyConnection[]>([])
  const [ipHistory, setIpHistory] = useState<IPHistoryEntry[]>([])
  const [commands, setCommands] = useState<DeviceCommand[]>([])
  const [rotationLinks, setRotationLinks] = useState<RotationLink[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<SidebarTab>('primary')
  const [commandFeedback, setCommandFeedback] = useState<string | null>(null)
  const [copiedId, setCopiedId] = useState<string | null>(null)

  // Editable name
  const [editingName, setEditingName] = useState(false)
  const [nameInput, setNameInput] = useState('')

  // Editable description
  const [descriptionInput, setDescriptionInput] = useState('')
  const [descriptionDirty, setDescriptionDirty] = useState(false)

  const fetchData = useCallback(async () => {
    const token = getToken()
    if (!token) return
    try {
      const [deviceRes, bwRes, connRes, histRes, cmdRes, linksRes] = await Promise.all([
        api.devices.get(token, id),
        api.devices.bandwidth(token, id),
        api.connections.list(token, id),
        api.devices.ipHistory(token, id),
        api.devices.commands(token, id),
        api.rotationLinks.list(token, id),
      ])
      setDevice(deviceRes)
      setBandwidth(bwRes)
      setConnections(connRes.connections || [])
      setIpHistory(histRes.history || [])
      setCommands(cmdRes.commands || [])
      setRotationLinks(linksRes.links || [])
      setNameInput(deviceRes.name || '')
      if (!descriptionDirty) {
        setDescriptionInput(deviceRes.description || '')
      }
    } catch (err) {
      console.error('Failed to fetch device detail:', err)
    } finally {
      setLoading(false)
    }
  }, [id, descriptionDirty])

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
    const interval = setInterval(fetchData, 15000)
    return () => { unsub(); clearInterval(interval) }
  }, [fetchData, id])

  async function saveName() {
    const token = getToken()
    if (!token || !device) return
    try {
      const updated = await api.devices.update(token, device.id, { name: nameInput })
      setDevice(updated)
      setEditingName(false)
    } catch (err) {
      console.error('Failed to update name:', err)
    }
  }

  async function saveDescription() {
    const token = getToken()
    if (!token || !device) return
    try {
      const updated = await api.devices.update(token, device.id, { description: descriptionInput })
      setDevice(updated)
      setDescriptionDirty(false)
    } catch (err) {
      console.error('Failed to update description:', err)
    }
  }

  async function sendCommand(type: string, label: string) {
    const token = getToken()
    if (!token || !device) return
    try {
      await api.devices.sendCommand(token, device.id, type)
      setCommandFeedback(`${label} command sent`)
      setTimeout(() => setCommandFeedback(null), 3000)
      const cmdRes = await api.devices.commands(token, device.id)
      setCommands(cmdRes.commands || [])
    } catch (err) {
      setCommandFeedback(`Failed to send ${label} command`)
      setTimeout(() => setCommandFeedback(null), 3000)
    }
  }

  async function handleCreateRotationLink() {
    const token = getToken()
    if (!token || !device) return
    try {
      const link = await api.rotationLinks.create(token, { device_id: device.id, name: `Link ${rotationLinks.length + 1}` })
      setRotationLinks(prev => [link, ...prev])
    } catch (err) {
      console.error('Failed to create rotation link:', err)
    }
  }

  async function handleDeleteRotationLink(linkId: string) {
    const token = getToken()
    if (!token) return
    try {
      await api.rotationLinks.delete(token, linkId)
      setRotationLinks(prev => prev.filter(l => l.id !== linkId))
    } catch (err) {
      console.error('Failed to delete rotation link:', err)
    }
  }

  function copyToClipboard(text: string, itemId: string) {
    navigator.clipboard.writeText(text)
    setCopiedId(itemId)
    setTimeout(() => setCopiedId(null), 2000)
  }

  function getRotationUrl(token: string) {
    const base = process.env.NEXT_PUBLIC_API_URL || `http://${SERVER_HOST}:8080/api`
    return `${base}/public/rotate/${token}`
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-zinc-500">Loading connection...</div>
      </div>
    )
  }

  if (!device) {
    return <div className="text-zinc-500">Connection not found</div>
  }

  const sidebarItems: { id: SidebarTab; label: string; icon: typeof Smartphone }[] = [
    { id: 'primary', label: 'Primary', icon: Smartphone },
    { id: 'advanced', label: 'Advanced', icon: Settings },
    { id: 'change-ip', label: 'Change IP', icon: Link2 },
    { id: 'history', label: 'History', icon: Clock },
    { id: 'metrics', label: 'Device Metrics', icon: Activity },
    { id: 'usage', label: 'Usage', icon: BarChart3 },
  ]

  return (
    <div>
      {/* Header */}
      <Link href="/connections" className="text-sm text-zinc-400 hover:text-white flex items-center gap-1 mb-4">
        <ArrowLeft className="w-4 h-4" /> Back to Connections
      </Link>

      <div className="mb-6">
        <div className="flex items-center gap-3">
          {editingName ? (
            <div className="flex items-center gap-2">
              <input
                value={nameInput}
                onChange={e => setNameInput(e.target.value)}
                className="text-xl font-bold bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-white focus:outline-none focus:ring-1 focus:ring-brand-500"
                autoFocus
                onKeyDown={e => {
                  if (e.key === 'Enter') saveName()
                  if (e.key === 'Escape') { setEditingName(false); setNameInput(device.name || '') }
                }}
              />
              <button onClick={saveName} className="text-green-400 hover:text-green-300 p-1">
                <Check className="w-4 h-4" />
              </button>
              <button onClick={() => { setEditingName(false); setNameInput(device.name || '') }} className="text-zinc-400 hover:text-white p-1">
                <X className="w-4 h-4" />
              </button>
            </div>
          ) : (
            <h1 className="text-xl font-bold flex items-center gap-2">
              {device.name || 'Unnamed Device'}
              <button
                onClick={() => { setEditingName(true); setNameInput(device.name || '') }}
                className="text-zinc-500 hover:text-white p-1"
                title="Edit name"
              >
                <Pencil className="w-3.5 h-3.5" />
              </button>
              <StatusBadge status={device.status} />
            </h1>
          )}
        </div>

        {/* Description */}
        <div className="mt-2">
          <textarea
            value={descriptionInput}
            onChange={e => { setDescriptionInput(e.target.value); setDescriptionDirty(true) }}
            onBlur={() => { if (descriptionDirty) saveDescription() }}
            placeholder="Add a description..."
            className="w-full max-w-xl text-sm bg-transparent border border-transparent hover:border-zinc-700 focus:border-zinc-600 rounded px-2 py-1 text-zinc-400 placeholder-zinc-600 resize-none focus:outline-none focus:ring-0 transition-colors"
            rows={1}
          />
        </div>

        <div className="text-sm text-zinc-500 mt-1">
          {device.device_model} &middot; ID: {device.id.slice(0, 8)}
        </div>
      </div>

      {/* Command feedback toast */}
      {commandFeedback && (
        <div className="fixed top-4 right-4 bg-zinc-800 border border-zinc-700 text-sm px-4 py-2 rounded-lg shadow-lg z-50">
          {commandFeedback}
        </div>
      )}

      {/* Main layout: Sidebar + Content */}
      <div className="flex gap-6">
        {/* Left Sidebar Tabs */}
        <div className="w-52 flex-shrink-0">
          <nav className="space-y-0.5">
            {sidebarItems.map(item => {
              const Icon = item.icon
              return (
                <button
                  key={item.id}
                  onClick={() => setActiveTab(item.id)}
                  className={cn(
                    'w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-colors text-left',
                    activeTab === item.id
                      ? 'bg-brand-600 text-white'
                      : 'text-zinc-400 hover:text-white hover:bg-zinc-800'
                  )}
                >
                  <Icon className="w-4 h-4" />
                  {item.label}
                </button>
              )
            })}
          </nav>
        </div>

        {/* Content Area */}
        <div className="flex-1 min-w-0">
          {activeTab === 'primary' && <PrimaryTab device={device} connections={connections} bandwidth={bandwidth} serverHost={SERVER_HOST} copyToClipboard={copyToClipboard} copiedId={copiedId} onConnectionsChange={setConnections} />}
          {activeTab === 'advanced' && <AdvancedTab device={device} commands={commands} sendCommand={sendCommand} />}
          {activeTab === 'change-ip' && <ChangeIPTab device={device} rotationLinks={rotationLinks} onCreateLink={handleCreateRotationLink} onDeleteLink={handleDeleteRotationLink} getRotationUrl={getRotationUrl} copyToClipboard={copyToClipboard} copiedId={copiedId} />}
          {activeTab === 'history' && <HistoryTab ipHistory={ipHistory} commands={commands} />}
          {activeTab === 'metrics' && <MetricsTab device={device} bandwidth={bandwidth} />}
          {activeTab === 'usage' && <UsageTab deviceId={device.id} />}
        </div>
      </div>
    </div>
  )
}

// ============= COPIABLE FIELD =============
function CopyField({ label, value, copyId, copyToClipboard, copiedId, mono }: {
  label: string
  value: string
  copyId: string
  copyToClipboard: (text: string, id: string) => void
  copiedId: string | null
  mono?: boolean
}) {
  return (
    <div className="flex items-center justify-between py-1">
      <span className="text-xs text-zinc-500">{label}</span>
      <div className="flex items-center gap-1.5">
        <span className={cn('text-sm text-zinc-200', mono && 'font-mono text-xs')}>{value}</span>
        <button
          onClick={() => copyToClipboard(value, copyId)}
          className="text-zinc-600 hover:text-white transition-colors p-0.5"
          title={`Copy ${label}`}
        >
          {copiedId === copyId ? <span className="text-green-400 text-[10px]">Copied</span> : <Copy className="w-3 h-3" />}
        </button>
      </div>
    </div>
  )
}

// ============= PRIMARY TAB =============
function PrimaryTab({ device, connections, bandwidth, serverHost, copyToClipboard, copiedId, onConnectionsChange }: {
  device: Device
  connections: ProxyConnection[]
  bandwidth: DeviceBandwidth | null
  serverHost: string
  copyToClipboard: (text: string, id: string) => void
  copiedId: string | null
  onConnectionsChange: (conns: ProxyConnection[]) => void
}) {
  const [showAddForm, setShowAddForm] = useState(false)
  const [formUsername, setFormUsername] = useState('')
  const [formPassword, setFormPassword] = useState('')
  const [formType, setFormType] = useState<'http' | 'socks5'>('http')

  async function handleCreateConnection(e: FormEvent) {
    e.preventDefault()
    const token = getToken()
    if (!token) return
    try {
      const conn = await api.connections.create(token, {
        device_id: device.id,
        username: formUsername,
        password: formPassword,
        proxy_type: formType,
      })
      onConnectionsChange([conn, ...connections])
      setFormUsername('')
      setFormPassword('')
      setShowAddForm(false)
    } catch (err) {
      console.error('Failed to create connection:', err)
    }
  }

  async function handleToggle(connId: string, active: boolean) {
    const token = getToken()
    if (!token) return
    try {
      await api.connections.setActive(token, connId, !active)
      onConnectionsChange(connections.map(c => c.id === connId ? { ...c, active: !active } : c))
    } catch (err) {
      console.error('Failed to toggle:', err)
    }
  }

  async function handleDelete(connId: string) {
    const token = getToken()
    if (!token) return
    if (!confirm('Delete this access point?')) return
    try {
      await api.connections.delete(token, connId)
      onConnectionsChange(connections.filter(c => c.id !== connId))
    } catch (err) {
      console.error('Failed to delete:', err)
    }
  }

  function getPort(conn: ProxyConnection): number {
    if (conn.proxy_type === 'socks5') {
      return conn.socks5_port ?? conn.base_port ?? device.socks5_port
    }
    return conn.http_port ?? conn.base_port ?? device.http_port
  }

  function getCopyAllString(conn: ProxyConnection): string {
    const port = getPort(conn)
    const type = conn.proxy_type === 'socks5' ? 'socks5' : 'http'
    return `${type}:${serverHost}:${port}:${conn.username}:${conn.password || ''}`
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-medium text-zinc-400">Access Points</h3>
        <button
          onClick={() => setShowAddForm(!showAddForm)}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-brand-600 hover:bg-brand-500 text-white text-xs font-medium rounded-lg transition-colors"
        >
          <Plus className="w-3.5 h-3.5" />
          Add Access Point
        </button>
      </div>

      {/* Add form */}
      {showAddForm && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 mb-4">
          <form onSubmit={handleCreateConnection} className="space-y-3">
            <div className="flex gap-3 items-end">
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Username</label>
                <input
                  value={formUsername}
                  onChange={e => setFormUsername(e.target.value)}
                  className="px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white text-sm w-40"
                  required
                />
              </div>
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Password</label>
                <input
                  value={formPassword}
                  onChange={e => setFormPassword(e.target.value)}
                  className="px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white text-sm w-40"
                  required
                />
              </div>
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Type</label>
                <div className="flex rounded overflow-hidden border border-zinc-700">
                  <button
                    type="button"
                    onClick={() => setFormType('http')}
                    className={cn(
                      'px-3 py-2 text-sm font-medium transition-colors',
                      formType === 'http'
                        ? 'bg-brand-600 text-white'
                        : 'bg-zinc-800 text-zinc-400 hover:text-white'
                    )}
                  >
                    HTTP
                  </button>
                  <button
                    type="button"
                    onClick={() => setFormType('socks5')}
                    className={cn(
                      'px-3 py-2 text-sm font-medium transition-colors',
                      formType === 'socks5'
                        ? 'bg-brand-600 text-white'
                        : 'bg-zinc-800 text-zinc-400 hover:text-white'
                    )}
                  >
                    SOCKS5
                  </button>
                </div>
              </div>
            </div>
            <div className="flex gap-2">
              <button type="submit" className="px-4 py-2 bg-brand-600 hover:bg-brand-500 text-white rounded text-sm">
                Create
              </button>
              <button type="button" onClick={() => setShowAddForm(false)} className="px-4 py-2 bg-zinc-700 hover:bg-zinc-600 text-white rounded text-sm">
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      {connections.length === 0 ? (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-8 text-center text-zinc-500">
          No access points configured. Click &quot;Add Access Point&quot; to create one.
        </div>
      ) : (
        <div className="space-y-3">
          {connections.map(conn => {
            const port = getPort(conn)
            const typeLabel = conn.proxy_type === 'socks5' ? 'SOCKS5' : 'HTTP'
            const typeBadgeColor = conn.proxy_type === 'socks5' ? 'bg-purple-900/30 text-purple-400' : 'bg-blue-900/30 text-blue-400'

            return (
              <div key={conn.id} className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
                {/* Header row */}
                <div className="px-4 py-2.5 bg-zinc-800/50 border-b border-zinc-800 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <span className={cn('text-xs px-2 py-0.5 rounded font-medium', typeBadgeColor)}>
                      {typeLabel}
                    </span>
                    <span className="text-sm font-medium font-mono">{conn.username}</span>
                    <span className={cn('text-xs px-2 py-0.5 rounded', conn.active ? 'bg-green-900/30 text-green-400' : 'bg-zinc-800 text-zinc-500')}>
                      {conn.active ? 'Active' : 'Disabled'}
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => handleToggle(conn.id, conn.active)}
                      className="px-2 py-1 text-xs bg-zinc-700 hover:bg-zinc-600 text-white rounded"
                    >
                      {conn.active ? 'Disable' : 'Enable'}
                    </button>
                    <button
                      onClick={() => handleDelete(conn.id)}
                      className="p-1 text-zinc-500 hover:text-red-400 transition-colors"
                      title="Delete"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                </div>

                {/* Connection details */}
                <div className="p-4 space-y-0">
                  <CopyField label="IP" value={serverHost} copyId={`ip-${conn.id}`} copyToClipboard={copyToClipboard} copiedId={copiedId} mono />
                  <CopyField label="Port" value={String(port)} copyId={`port-${conn.id}`} copyToClipboard={copyToClipboard} copiedId={copiedId} mono />
                  <CopyField label="Username" value={conn.username} copyId={`user-${conn.id}`} copyToClipboard={copyToClipboard} copiedId={copiedId} mono />
                  <CopyField label="Password" value={conn.password || '••••••'} copyId={`pass-${conn.id}`} copyToClipboard={copyToClipboard} copiedId={copiedId} mono />

                  {/* Copy All button */}
                  <div className="pt-2 mt-2 border-t border-zinc-800/50">
                    <button
                      onClick={() => copyToClipboard(getCopyAllString(conn), `all-${conn.id}`)}
                      className="w-full flex items-center justify-center gap-2 px-3 py-1.5 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs font-mono rounded transition-colors"
                    >
                      {copiedId === `all-${conn.id}` ? (
                        <span className="text-green-400">Copied!</span>
                      ) : (
                        <>
                          <Copy className="w-3 h-3" />
                          {getCopyAllString(conn)}
                        </>
                      )}
                    </button>
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* External IP */}
      <div className="mt-6 bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <div className="text-sm text-zinc-400 mb-2">External IP (Cellular)</div>
        <div className="font-mono text-lg">{device.cellular_ip || 'Unknown'}</div>
      </div>
    </div>
  )
}

// ============= ADVANCED TAB =============
function AdvancedTab({ device, commands, sendCommand }: {
  device: Device
  commands: DeviceCommand[]
  sendCommand: (type: string, label: string) => void
}) {
  const [subTab, setSubTab] = useState<'actions' | 'config'>('actions')

  const actionButtons = [
    { type: 'rotate_ip', label: 'Rotate IP', icon: RotateCw, color: 'bg-brand-600 hover:bg-brand-500', description: 'Change cellular IP via airplane mode' },
    { type: 'reboot', label: 'Reboot Device', icon: Power, color: 'bg-orange-600 hover:bg-orange-700', description: 'Restart the Android device' },
    { type: 'find_phone', label: 'Find Phone', icon: Search, color: 'bg-purple-600 hover:bg-purple-700', description: 'Play sound on the device' },
    { type: 'wifi_on', label: 'WiFi On', icon: Wifi, color: 'bg-green-600 hover:bg-green-700', description: 'Enable WiFi' },
    { type: 'wifi_off', label: 'WiFi Off', icon: WifiOff, color: 'bg-zinc-600 hover:bg-zinc-700', description: 'Disable WiFi' },
  ]

  return (
    <div>
      <div className="flex gap-1 border-b border-zinc-800 mb-6">
        <button
          onClick={() => setSubTab('actions')}
          className={cn('px-4 py-2 text-sm border-b-2 -mb-px transition-colors',
            subTab === 'actions' ? 'border-brand-500 text-white' : 'border-transparent text-zinc-400 hover:text-white'
          )}
        >
          Actions
        </button>
        <button
          onClick={() => setSubTab('config')}
          className={cn('px-4 py-2 text-sm border-b-2 -mb-px transition-colors',
            subTab === 'config' ? 'border-brand-500 text-white' : 'border-transparent text-zinc-400 hover:text-white'
          )}
        >
          Configuration
        </button>
      </div>

      {subTab === 'actions' && (
        <div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-8">
            {actionButtons.map(action => {
              const Icon = action.icon
              const disabled = device.status !== 'online'
              return (
                <button
                  key={action.type}
                  onClick={() => sendCommand(action.type, action.label)}
                  disabled={disabled}
                  className={cn(
                    'flex items-center gap-3 px-4 py-3 rounded-lg text-white text-left transition-all',
                    disabled ? 'bg-zinc-800 text-zinc-500 cursor-not-allowed' : action.color
                  )}
                >
                  <Icon className="w-5 h-5 flex-shrink-0" />
                  <div>
                    <div className="text-sm font-medium">{action.label}</div>
                    <div className={cn('text-xs', disabled ? 'text-zinc-600' : 'text-white/70')}>
                      {action.description}
                    </div>
                  </div>
                </button>
              )
            })}
          </div>

          {/* Recent commands */}
          <h3 className="text-sm font-medium text-zinc-400 mb-3">Recent Commands</h3>
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-zinc-800 text-zinc-500 text-left">
                  <th className="px-4 py-2 font-medium text-xs">Command</th>
                  <th className="px-4 py-2 font-medium text-xs">Status</th>
                  <th className="px-4 py-2 font-medium text-xs">Sent</th>
                  <th className="px-4 py-2 font-medium text-xs">Executed</th>
                </tr>
              </thead>
              <tbody>
                {commands.slice(0, 10).map(cmd => (
                  <tr key={cmd.id} className="border-b border-zinc-800/30 hover:bg-zinc-800/20">
                    <td className="px-4 py-2">
                      <span className="px-2 py-0.5 rounded text-xs bg-zinc-800 text-zinc-300">
                        {cmd.type}
                      </span>
                    </td>
                    <td className="px-4 py-2"><StatusBadge status={cmd.status} /></td>
                    <td className="px-4 py-2 text-xs text-zinc-400">{timeAgo(cmd.created_at)}</td>
                    <td className="px-4 py-2 text-xs text-zinc-400">{cmd.executed_at ? timeAgo(cmd.executed_at) : '-'}</td>
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
      )}

      {subTab === 'config' && (
        <div className="space-y-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
            <h3 className="text-sm font-medium text-zinc-400 mb-4">IP Rotation Method</h3>
            <p className="text-sm text-zinc-300 mb-2">
              Current method: <span className="font-medium text-white">Airplane Mode Toggle</span>
            </p>
            <p className="text-xs text-zinc-500">
              Uses Settings.Global to toggle airplane mode ON/OFF with a 5-second delay.
              Requires WRITE_SECURE_SETTINGS permission granted via ADB.
            </p>
          </div>

          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
            <h3 className="text-sm font-medium text-zinc-400 mb-4">Device Status</h3>
            <div className="space-y-2 text-sm">
              <div className="flex items-center justify-between">
                <span className="text-zinc-400">Status</span>
                <StatusBadge status={device.status} />
              </div>
              <div className="flex items-center justify-between">
                <span className="text-zinc-400">Last heartbeat</span>
                <span className="text-zinc-300">{timeAgo(device.last_heartbeat)}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-zinc-400">VPN Connection</span>
                <span className={device.vpn_ip ? 'text-green-400' : 'text-zinc-500'}>
                  {device.vpn_ip ? `Connected (${device.vpn_ip})` : 'Disconnected'}
                </span>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

// ============= CHANGE IP TAB =============
function ChangeIPTab({ device, rotationLinks, onCreateLink, onDeleteLink, getRotationUrl, copyToClipboard, copiedId }: {
  device: Device
  rotationLinks: RotationLink[]
  onCreateLink: () => void
  onDeleteLink: (id: string) => void
  getRotationUrl: (token: string) => string
  copyToClipboard: (text: string, id: string) => void
  copiedId: string | null
}) {
  const [subTab, setSubTab] = useState<'url' | 'rotation'>('url')

  return (
    <div>
      <div className="flex gap-1 border-b border-zinc-800 mb-6">
        <button
          onClick={() => setSubTab('url')}
          className={cn('px-4 py-2 text-sm border-b-2 -mb-px transition-colors',
            subTab === 'url' ? 'border-brand-500 text-white' : 'border-transparent text-zinc-400 hover:text-white'
          )}
        >
          URL
        </button>
        <button
          onClick={() => setSubTab('rotation')}
          className={cn('px-4 py-2 text-sm border-b-2 -mb-px transition-colors',
            subTab === 'rotation' ? 'border-brand-500 text-white' : 'border-transparent text-zinc-400 hover:text-white'
          )}
        >
          About
        </button>
      </div>

      {subTab === 'url' && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-medium text-zinc-400">URL for IP address change</h3>
            <button
              onClick={onCreateLink}
              className="flex items-center gap-2 px-4 py-2 bg-brand-600 hover:bg-brand-500 text-white rounded-lg text-sm transition-colors"
            >
              <Plus className="w-4 h-4" />
              Add URL
            </button>
          </div>

          {rotationLinks.length === 0 ? (
            <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-8 text-center">
              <Link2 className="w-8 h-8 text-zinc-600 mx-auto mb-3" />
              <p className="text-zinc-500 text-sm">No rotation links yet.</p>
              <p className="text-zinc-600 text-xs mt-1">
                Create a link that anyone can use to trigger IP rotation on this device.
              </p>
            </div>
          ) : (
            <div className="space-y-2">
              {rotationLinks.map(link => {
                const url = getRotationUrl(link.token)
                return (
                  <div key={link.id} className="bg-zinc-900 border border-zinc-800 rounded-lg p-3 flex items-center justify-between gap-3">
                    <div className="flex-1 min-w-0">
                      <a
                        href={url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-brand-400 hover:text-brand-300 text-sm font-mono truncate block"
                      >
                        {url}
                      </a>
                      <div className="text-xs text-zinc-500 mt-1">
                        {link.name && <span className="mr-3">{link.name}</span>}
                        Created {formatDate(link.created_at)}
                        {link.last_used_at && <span className="ml-3">Last used: {timeAgo(link.last_used_at)}</span>}
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => copyToClipboard(url, `link-${link.id}`)}
                        className="p-2 text-zinc-500 hover:text-white transition-colors"
                        title="Copy URL"
                      >
                        {copiedId === `link-${link.id}` ? <span className="text-green-400 text-xs">Copied!</span> : <Copy className="w-4 h-4" />}
                      </button>
                      <button
                        onClick={() => onDeleteLink(link.id)}
                        className="p-2 text-zinc-500 hover:text-red-400 transition-colors"
                        title="Delete link"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      )}

      {subTab === 'rotation' && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
          <h3 className="text-sm font-medium text-zinc-400 mb-3">How Rotation Links Work</h3>
          <div className="text-sm text-zinc-300 space-y-2">
            <p>Rotation links are public URLs that trigger an IP rotation on this device when accessed.</p>
            <p>When someone visits the link, the server sends a <code className="px-1.5 py-0.5 bg-zinc-800 rounded text-xs">rotate_ip</code> command to the device, which toggles airplane mode to get a new cellular IP.</p>
            <p>No authentication is required to use a rotation link - share it with anyone who needs to trigger IP changes.</p>
          </div>
        </div>
      )}
    </div>
  )
}

// ============= HISTORY TAB =============
function HistoryTab({ ipHistory, commands }: {
  ipHistory: IPHistoryEntry[]
  commands: DeviceCommand[]
}) {
  const [subTab, setSubTab] = useState<'ip' | 'commands'>('ip')

  return (
    <div>
      <div className="flex gap-1 border-b border-zinc-800 mb-6">
        <button
          onClick={() => setSubTab('ip')}
          className={cn('px-4 py-2 text-sm border-b-2 -mb-px transition-colors',
            subTab === 'ip' ? 'border-brand-500 text-white' : 'border-transparent text-zinc-400 hover:text-white'
          )}
        >
          IP History
        </button>
        <button
          onClick={() => setSubTab('commands')}
          className={cn('px-4 py-2 text-sm border-b-2 -mb-px transition-colors',
            subTab === 'commands' ? 'border-brand-500 text-white' : 'border-transparent text-zinc-400 hover:text-white'
          )}
        >
          Command History
        </button>
      </div>

      {subTab === 'ip' && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-zinc-800 text-zinc-500 text-left">
                <th className="px-4 py-2 font-medium text-xs">#</th>
                <th className="px-4 py-2 font-medium text-xs">Date</th>
                <th className="px-4 py-2 font-medium text-xs">IP</th>
                <th className="px-4 py-2 font-medium text-xs">Method</th>
              </tr>
            </thead>
            <tbody>
              {ipHistory.map((entry, idx) => (
                <tr key={entry.id} className="border-b border-zinc-800/30 hover:bg-zinc-800/20">
                  <td className="px-4 py-2 text-zinc-500">{idx + 1}</td>
                  <td className="px-4 py-2 text-zinc-300">{formatDate(entry.created_at)}</td>
                  <td className="px-4 py-2 font-mono text-xs">{entry.ip}</td>
                  <td className="px-4 py-2">
                    <span className="px-2 py-0.5 rounded text-xs bg-zinc-800 text-zinc-300">
                      {entry.method}
                    </span>
                  </td>
                </tr>
              ))}
              {ipHistory.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-4 py-6 text-center text-zinc-500">
                    No IP history yet
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      {subTab === 'commands' && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-zinc-800 text-zinc-500 text-left">
                <th className="px-4 py-2 font-medium text-xs">Type</th>
                <th className="px-4 py-2 font-medium text-xs">Status</th>
                <th className="px-4 py-2 font-medium text-xs">Created</th>
                <th className="px-4 py-2 font-medium text-xs">Executed</th>
              </tr>
            </thead>
            <tbody>
              {commands.map(cmd => (
                <tr key={cmd.id} className="border-b border-zinc-800/30 hover:bg-zinc-800/20">
                  <td className="px-4 py-2">
                    <span className="px-2 py-0.5 rounded text-xs bg-zinc-800 text-zinc-300">
                      {cmd.type}
                    </span>
                  </td>
                  <td className="px-4 py-2"><StatusBadge status={cmd.status} /></td>
                  <td className="px-4 py-2 text-zinc-400">{formatDate(cmd.created_at)}</td>
                  <td className="px-4 py-2 text-zinc-400">{cmd.executed_at ? formatDate(cmd.executed_at) : '-'}</td>
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
      )}
    </div>
  )
}

// ============= METRICS TAB =============
function MetricsTab({ device, bandwidth }: {
  device: Device
  bandwidth: DeviceBandwidth | null
}) {
  const [subTab, setSubTab] = useState<'battery' | 'device' | 'network'>('battery')

  return (
    <div>
      <div className="flex gap-1 border-b border-zinc-800 mb-6">
        <button
          onClick={() => setSubTab('battery')}
          className={cn('px-4 py-2 text-sm border-b-2 -mb-px transition-colors',
            subTab === 'battery' ? 'border-brand-500 text-white' : 'border-transparent text-zinc-400 hover:text-white'
          )}
        >
          Battery
        </button>
        <button
          onClick={() => setSubTab('device')}
          className={cn('px-4 py-2 text-sm border-b-2 -mb-px transition-colors',
            subTab === 'device' ? 'border-brand-500 text-white' : 'border-transparent text-zinc-400 hover:text-white'
          )}
        >
          Device
        </button>
        <button
          onClick={() => setSubTab('network')}
          className={cn('px-4 py-2 text-sm border-b-2 -mb-px transition-colors',
            subTab === 'network' ? 'border-brand-500 text-white' : 'border-transparent text-zinc-400 hover:text-white'
          )}
        >
          Network
        </button>
      </div>

      {subTab === 'battery' && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6">
          <div className="space-y-4">
            <div className="flex items-center gap-3">
              <Battery className="w-5 h-5 text-zinc-500" />
              <div>
                <div className="text-sm text-zinc-400">Battery Level</div>
                <div className="text-2xl font-bold">{device.battery_level}%</div>
              </div>
            </div>

            {/* Battery bar */}
            <div className="w-full h-4 bg-zinc-800 rounded-full overflow-hidden">
              <div
                className={cn('h-full rounded-full transition-all',
                  device.battery_level > 50 ? 'bg-green-500' :
                  device.battery_level > 20 ? 'bg-yellow-500' : 'bg-red-500'
                )}
                style={{ width: `${device.battery_level}%` }}
              />
            </div>

            <div className="flex items-center gap-3">
              <span className="text-sm text-zinc-400">Charging:</span>
              <span className={cn('text-sm', device.battery_charging ? 'text-green-400' : 'text-zinc-500')}>
                {device.battery_charging ? 'Yes' : 'No'}
              </span>
            </div>

            <div className="text-xs text-zinc-600 mt-2">
              Last updated: {timeAgo(device.last_heartbeat)}
            </div>
          </div>
        </div>
      )}

      {subTab === 'device' && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6">
          <div className="space-y-3">
            <div className="flex items-center gap-3 mb-4">
              <Cpu className="w-5 h-5 text-zinc-500" />
              <span className="text-sm font-medium text-zinc-300">Device Information</span>
            </div>

            <InfoRow label="Model" value={device.device_model || '-'} />
            <InfoRow label="Android Version" value={device.android_version || '-'} />
            <InfoRow label="App Version" value={device.app_version || '-'} />
            <InfoRow label="Android ID" value={device.android_id} mono />
            <InfoRow label="Registered" value={formatDate(device.created_at)} />
          </div>
        </div>
      )}

      {subTab === 'network' && (
        <div className="space-y-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6">
            <div className="flex items-center gap-3 mb-4">
              <Globe className="w-5 h-5 text-zinc-500" />
              <span className="text-sm font-medium text-zinc-300">Network Status</span>
            </div>
            <div className="space-y-3">
              <InfoRow label="Carrier" value={device.carrier || '-'} />
              <InfoRow label="Network Type" value={device.network_type || '-'} />
              <InfoRow label="Signal Strength" value={`${device.signal_strength} dBm`} />
              <InfoRow label="Cellular IP" value={device.cellular_ip || '-'} mono />
              <InfoRow label="WiFi IP" value={device.wifi_ip || '-'} mono />
              <InfoRow label="VPN IP" value={device.vpn_ip || '-'} mono />
            </div>
          </div>

          {bandwidth && (
            <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6">
              <div className="flex items-center gap-3 mb-4">
                <RefreshCw className="w-5 h-5 text-zinc-500" />
                <span className="text-sm font-medium text-zinc-300">Bandwidth Usage</span>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="bg-zinc-800/50 rounded-lg p-3 text-center">
                  <div className="text-xs text-zinc-500">Today</div>
                  <div className="text-lg font-bold mt-1">{formatBytes(bandwidth.today_in + bandwidth.today_out)}</div>
                  <div className="text-xs text-zinc-500">{formatBytes(bandwidth.today_in)} in / {formatBytes(bandwidth.today_out)} out</div>
                </div>
                <div className="bg-zinc-800/50 rounded-lg p-3 text-center">
                  <div className="text-xs text-zinc-500">This Month</div>
                  <div className="text-lg font-bold mt-1">{formatBytes(bandwidth.month_in + bandwidth.month_out)}</div>
                  <div className="text-xs text-zinc-500">{formatBytes(bandwidth.month_in)} in / {formatBytes(bandwidth.month_out)} out</div>
                </div>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// ============= USAGE TAB =============

const TIMEZONE_OPTIONS = [
  { label: 'Local', value: 'local' },
  { label: 'UTC', value: '0' },
  { label: 'EST (UTC-5)', value: '300' },
  { label: 'CST (UTC-6)', value: '360' },
  { label: 'MST (UTC-7)', value: '420' },
  { label: 'PST (UTC-8)', value: '480' },
  { label: 'CET (UTC+1)', value: '-60' },
  { label: 'EET (UTC+2)', value: '-120' },
  { label: 'IST (UTC+5:30)', value: '-330' },
  { label: 'JST (UTC+9)', value: '-540' },
]

function getLocalOffset() {
  return new Date().getTimezoneOffset()
}

function UsageTab({ deviceId }: { deviceId: string }) {
  const today = new Date().toISOString().split('T')[0]
  const [date, setDate] = useState(today)
  const [tzSelection, setTzSelection] = useState('local')
  const [hourlyData, setHourlyData] = useState<BandwidthHourly[]>([])
  const [uptimeSegments, setUptimeSegments] = useState<UptimeSegment[]>([])
  const [loading, setLoading] = useState(true)

  const tzOffset = tzSelection === 'local' ? getLocalOffset() : parseInt(tzSelection, 10)

  const fetchUsageData = useCallback(async () => {
    const token = getToken()
    if (!token) return
    setLoading(true)
    try {
      const [bwRes, uptimeRes] = await Promise.all([
        api.devices.bandwidthHourly(token, deviceId, date, tzOffset),
        api.devices.uptime(token, deviceId, date, tzOffset),
      ])
      setHourlyData(bwRes.hourly || [])
      setUptimeSegments(uptimeRes.segments || [])
    } catch (err) {
      console.error('Failed to fetch usage data:', err)
    } finally {
      setLoading(false)
    }
  }, [deviceId, date, tzOffset])

  useEffect(() => {
    fetchUsageData()
  }, [fetchUsageData])

  // Compute timezone label for display
  const offsetHours = -tzOffset / 60
  const tzLabel = tzOffset === 0 ? 'UTC' : `UTC${offsetHours >= 0 ? '+' : ''}${offsetHours % 1 === 0 ? offsetHours.toFixed(0) : offsetHours.toFixed(1)}`

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h3 className="text-sm font-medium text-zinc-400">
          Usage & Uptime
          <span className="text-zinc-600 ml-2 text-xs">({tzLabel})</span>
        </h3>
        <div className="flex items-center gap-3">
          <select
            value={tzSelection}
            onChange={(e) => setTzSelection(e.target.value)}
            className="bg-zinc-800 border border-zinc-700 text-zinc-200 text-sm rounded-lg px-3 py-1.5 focus:outline-none focus:ring-1 focus:ring-brand-500"
          >
            {TIMEZONE_OPTIONS.map(tz => (
              <option key={tz.value} value={tz.value}>{tz.label}</option>
            ))}
          </select>
          <input
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            max={today}
            className="bg-zinc-800 border border-zinc-700 text-zinc-200 text-sm rounded-lg px-3 py-1.5 focus:outline-none focus:ring-1 focus:ring-brand-500"
          />
        </div>
      </div>

      {loading ? (
        <div className="text-zinc-500 text-center py-8">Loading usage data...</div>
      ) : (
        <div className="space-y-6">
          <BandwidthChart data={hourlyData} />
          <UptimeTimeline segments={uptimeSegments} />
        </div>
      )}
    </div>
  )
}

// ============= SHARED COMPONENTS =============
function InfoRow({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="flex items-center justify-between py-1.5 border-b border-zinc-800/50 last:border-0">
      <span className="text-zinc-500 text-sm">{label}</span>
      <span className={cn('text-sm text-zinc-200', mono && 'font-mono text-xs')}>{value}</span>
    </div>
  )
}
