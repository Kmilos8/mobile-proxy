'use client'

import { useEffect, useState, useCallback } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import {
  ArrowLeft, Smartphone, Settings, Link2, Clock, Activity,
  RotateCw, Power, Search, Wifi, WifiOff, Copy, Trash2, Plus,
  Battery, Signal, Globe, Cpu, RefreshCw, ChevronRight, BarChart3,
  X, Check, QrCode
} from 'lucide-react'
import { QRCodeSVG } from 'qrcode.react'
import { api, Device, DeviceBandwidth, DeviceCommand, ProxyConnection, IPHistoryEntry, RotationLink, BandwidthHourly, UptimeSegment } from '@/lib/api'
import { getToken } from '@/lib/auth'
import { addWSHandler } from '@/lib/websocket'
import { formatBytes, formatDate, timeAgo, cn, copyToClipboard } from '@/lib/utils'
import StatusBadge from '@/components/ui/StatusBadge'
import BatteryIndicator from '@/components/ui/BatteryIndicator'
import BandwidthChart from '@/components/BandwidthChart'
import UptimeTimeline from '@/components/UptimeTimeline'
import ConnectionTable from '@/components/connections/ConnectionTable'
import AddConnectionModal from '@/components/connections/AddConnectionModal'
import { Button } from '@/components/ui/button'

type SidebarTab = 'primary' | 'openvpn' | 'advanced' | 'change-ip' | 'history' | 'metrics' | 'usage'

const SERVER_HOST = process.env.NEXT_PUBLIC_SERVER_HOST || '178.156.240.184'

function RepairModal({ deviceId, onClose }: { deviceId: string; onClose: () => void }) {
  const [code, setCode] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [expiresAt, setExpiresAt] = useState<string | null>(null)
  const [timeLeft, setTimeLeft] = useState('')

  useEffect(() => {
    let cancelled = false
    async function create() {
      const token = getToken()
      if (!token) return
      try {
        const res = await api.pairingCodes.create(token, 5, undefined, deviceId)
        if (cancelled) return
        setCode(res.code)
        setExpiresAt(res.expires_at)
      } catch (err) {
        if (cancelled) return
        setError(err instanceof Error ? err.message : 'Failed to create pairing code')
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    create()
    return () => { cancelled = true }
  }, [deviceId])

  useEffect(() => {
    if (!expiresAt) return
    const timer = setInterval(() => {
      const diff = new Date(expiresAt).getTime() - Date.now()
      if (diff <= 0) {
        setTimeLeft('Expired')
        clearInterval(timer)
        return
      }
      const minutes = Math.floor(diff / 60000)
      const seconds = Math.floor((diff % 60000) / 1000)
      setTimeLeft(`${minutes}m ${seconds}s remaining`)
    }, 1000)
    return () => clearInterval(timer)
  }, [expiresAt])

  function formatCode(c: string): string {
    if (c.length === 8) return c.slice(0, 4) + '-' + c.slice(4)
    return c
  }

  const qrValue = code ? `mobileproxy://pair?server=http://${SERVER_HOST}:8080&code=${code}` : ''

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-zinc-900 border border-zinc-700 rounded-xl p-6 max-w-md w-full mx-4" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-white">Replace Device</h2>
          <button onClick={onClose} className="text-zinc-400 hover:text-white p-1">
            <X className="w-5 h-5" />
          </button>
        </div>

        <p className="text-sm text-zinc-400 mb-4">
          Scan this QR code with the <strong>new phone</strong> to replace this device. All connections will transfer to the new device and the old phone will be logged out.
        </p>

        {loading && (
          <div className="text-center py-8 text-zinc-500">Generating pairing code...</div>
        )}

        {error && (
          <div className="text-center py-8 text-red-400">{error}</div>
        )}

        {code && !loading && (
          <div className="space-y-4">
            <div className="flex justify-center">
              <div className="bg-white p-4 rounded-lg">
                <QRCodeSVG value={qrValue} size={200} />
              </div>
            </div>

            <div className="text-center">
              <div className="text-3xl font-mono font-bold tracking-widest text-white">
                {formatCode(code)}
              </div>
              <button
                onClick={() => {
                  if (!code) return
                  copyToClipboard(formatCode(code))
                  setCopied(true)
                  setTimeout(() => setCopied(false), 2000)
                }}
                className="mt-2 inline-flex items-center gap-1 text-sm text-brand-400 hover:text-brand-300"
              >
                {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                {copied ? 'Copied!' : 'Copy code'}
              </button>
            </div>

            <div className="text-center text-xs text-zinc-500">{timeLeft}</div>
          </div>
        )}
      </div>
    </div>
  )
}

export default function DeviceDetailPage() {
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
  const [addConnectionOpen, setAddConnectionOpen] = useState(false)
  const [showRepairModal, setShowRepairModal] = useState(false)
  const [editing, setEditing] = useState(false)
  const [editName, setEditName] = useState('')
  const [editDescription, setEditDescription] = useState('')
  const [saving, setSaving] = useState(false)

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
    const interval = setInterval(fetchData, 15000)
    return () => { unsub(); clearInterval(interval) }
  }, [fetchData, id])

  async function sendCommand(type: string, label: string) {
    const token = getToken()
    if (!token || !device) return
    try {
      await api.devices.sendCommand(token, device.id, type)
      setCommandFeedback(`${label} command sent`)
      setTimeout(() => setCommandFeedback(null), 3000)
      // Refresh commands list
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

  async function handleDeleteConnection(connectionId: string) {
    const token = getToken()
    if (!token) return
    try {
      await api.connections.delete(token, connectionId)
      await fetchData()
    } catch (err) {
      console.error('Failed to delete connection:', err)
    }
  }

  function handleCopy(text: string, itemId: string) {
    copyToClipboard(text)
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
        <div className="text-zinc-500">Loading device...</div>
      </div>
    )
  }

  if (!device) {
    return <div className="text-zinc-500">Device not found</div>
  }

  const sidebarItems: { id: SidebarTab; label: string; icon: typeof Smartphone }[] = [
    { id: 'primary', label: 'Primary', icon: Smartphone },
    { id: 'openvpn', label: 'OpenVPN', icon: Globe },
    { id: 'advanced', label: 'Advanced', icon: Settings },
    { id: 'change-ip', label: 'Change IP', icon: Link2 },
    { id: 'history', label: 'History', icon: Clock },
    { id: 'metrics', label: 'Device Metrics', icon: Activity },
    { id: 'usage', label: 'Usage', icon: BarChart3 },
  ]

  return (
    <div>
      {/* Header */}
      <Link href="/devices" className="text-sm text-zinc-400 hover:text-white flex items-center gap-1 mb-4">
        <ArrowLeft className="w-4 h-4" /> Back to Devices
      </Link>

      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <div>
            {editing ? (
              <div className="space-y-2">
                <input
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  placeholder="Device name"
                  className="bg-zinc-800 border border-zinc-700 text-white text-lg font-bold px-3 py-1 rounded focus:outline-none focus:border-brand-500 w-64"
                  autoFocus
                />
                <input
                  value={editDescription}
                  onChange={(e) => setEditDescription(e.target.value)}
                  placeholder="Description (optional)"
                  className="bg-zinc-800 border border-zinc-700 text-zinc-300 text-sm px-3 py-1 rounded focus:outline-none focus:border-brand-500 w-64 block"
                />
                <div className="flex items-center gap-2">
                  <button
                    onClick={async () => {
                      const token = getToken()
                      if (!token) return
                      setSaving(true)
                      try {
                        const updated = await api.devices.update(token, device.id, { name: editName, description: editDescription })
                        setDevice(updated)
                        setEditing(false)
                      } catch (err) {
                        console.error('Failed to update device:', err)
                      } finally {
                        setSaving(false)
                      }
                    }}
                    disabled={saving}
                    className="inline-flex items-center gap-1 px-3 py-1 bg-brand-600 hover:bg-brand-500 text-white text-xs font-medium rounded transition-colors"
                  >
                    <Check className="w-3 h-3" />
                    {saving ? 'Saving...' : 'Save'}
                  </button>
                  <button
                    onClick={() => setEditing(false)}
                    className="inline-flex items-center gap-1 px-3 py-1 text-zinc-400 hover:text-white text-xs rounded transition-colors"
                  >
                    <X className="w-3 h-3" />
                    Cancel
                  </button>
                </div>
              </div>
            ) : (
              <>
                <h1 className="text-xl font-bold flex items-center gap-2">
                  {device.name || 'Unnamed Device'}
                  <StatusBadge status={device.status} />
                  <button
                    onClick={() => {
                      setEditName(device.name || '')
                      setEditDescription(device.description || '')
                      setEditing(true)
                    }}
                    className="text-zinc-500 hover:text-white transition-colors"
                    title="Edit device"
                  >
                    <Settings className="w-4 h-4" />
                  </button>
                </h1>
                <div className="text-sm text-zinc-500 mt-0.5">
                  {device.device_model} &middot; Connection ID: {device.id.slice(0, 8)}
                  {device.description && <span> &middot; {device.description}</span>}
                </div>
              </>
            )}
          </div>
        </div>
        <button
          onClick={() => setShowRepairModal(true)}
          className="inline-flex items-center gap-2 px-3 py-2 text-sm text-zinc-400 hover:text-white hover:bg-zinc-800 border border-zinc-700 rounded-lg transition-colors"
        >
          <QrCode className="w-4 h-4" />
          Replace Device
        </button>
      </div>

      {showRepairModal && (
        <RepairModal deviceId={device.id} onClose={() => setShowRepairModal(false)} />
      )}

      {/* Command feedback toast */}
      {commandFeedback && (
        <div className="fixed top-4 right-4 bg-zinc-800 border border-zinc-700 text-sm px-4 py-2 rounded-lg shadow-lg z-50">
          {commandFeedback}
        </div>
      )}

      {/* Main layout: Sidebar + Content */}
      <div className="flex gap-6">
        {/* Left Sidebar Tabs — hidden on mobile, shown at md+ */}
        <div className="hidden md:block w-52 flex-shrink-0">
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

        {/* Mobile horizontal tab bar — shown below md */}
        <div className="md:hidden w-full">
          <div className="flex overflow-x-auto gap-1 border-b border-zinc-800 mb-4 pb-1">
            {sidebarItems.map(item => (
              <button
                key={item.id}
                onClick={() => setActiveTab(item.id)}
                className={cn(
                  'flex-shrink-0 px-3 py-2 text-sm rounded-t-lg transition-colors',
                  activeTab === item.id
                    ? 'bg-brand-600 text-white'
                    : 'text-zinc-400 hover:text-white'
                )}
              >
                {item.label}
              </button>
            ))}
          </div>
        </div>

        {/* Content Area */}
        <div className="flex-1 min-w-0">
          {activeTab === 'primary' && (
            <PrimaryTab
              device={device}
              connections={connections}
              bandwidth={bandwidth}
              serverHost={device.relay_server_ip || '178.156.210.156'}
              copyToClipboard={handleCopy}
              copiedId={copiedId}
              addConnectionOpen={addConnectionOpen}
              onAddConnectionOpenChange={setAddConnectionOpen}
              onDeleteConnection={handleDeleteConnection}
              onRefresh={fetchData}
            />
          )}
          {activeTab === 'openvpn' && (
            <OpenVPNTab
              device={device}
              connections={connections.filter(c => c.proxy_type === 'openvpn')}
              onRefresh={fetchData}
            />
          )}
          {activeTab === 'advanced' && <AdvancedTab device={device} commands={commands} sendCommand={sendCommand} />}
          {activeTab === 'change-ip' && <ChangeIPTab device={device} rotationLinks={rotationLinks} onCreateLink={handleCreateRotationLink} onDeleteLink={handleDeleteRotationLink} getRotationUrl={getRotationUrl} copyToClipboard={handleCopy} copiedId={copiedId} onAutoRotateChange={(minutes) => {
            const token = getToken()
            if (!token) return
            api.devices.update(token, device.id, { auto_rotate_minutes: minutes }).then(updated => setDevice(updated))
          }} />}
          {activeTab === 'history' && <HistoryTab ipHistory={ipHistory} />}
          {activeTab === 'metrics' && <MetricsTab device={device} bandwidth={bandwidth} />}
          {activeTab === 'usage' && <UsageTab deviceId={device.id} />}
        </div>
      </div>

      {/* Add Connection Modal — rendered at page level so it isn't clipped */}
      <AddConnectionModal
        deviceId={device.id}
        open={addConnectionOpen}
        onOpenChange={setAddConnectionOpen}
        onCreated={fetchData}
      />
    </div>
  )
}

// ============= PRIMARY TAB =============
function OpenVPNTab({ device, connections, onRefresh }: {
  device: Device
  connections: ProxyConnection[]
  onRefresh: () => void
}) {
  const [showCreate, setShowCreate] = useState(false)
  const [name, setName] = useState('')
  const [creating, setCreating] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    const token = getToken()
    if (!token || !name.trim()) return
    setCreating(true)
    setError(null)
    try {
      await api.connections.create(token, {
        device_id: device.id,
        username: name.trim(),
        password: Math.random().toString(36).slice(2, 14),
        proxy_type: 'openvpn',
      })
      setName('')
      setShowCreate(false)
      onRefresh()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create config')
    } finally {
      setCreating(false)
    }
  }

  async function handleDownload(connId: string) {
    const token = getToken()
    if (!token) return
    try {
      await api.connections.downloadOVPN(token, connId)
    } catch (err) {
      console.error('Download failed:', err)
    }
  }

  async function handleDelete(connId: string) {
    const token = getToken()
    if (!token) return
    try {
      await api.connections.delete(token, connId)
      onRefresh()
    } catch (err) {
      console.error('Delete failed:', err)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">OpenVPN Configs</h2>
        <Button
          onClick={() => setShowCreate(true)}
          className="bg-brand-600 hover:bg-brand-500 text-white"
          size="sm"
        >
          <Plus className="w-4 h-4 mr-1" />
          Create Config
        </Button>
      </div>

      {/* Create form */}
      {showCreate && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
          <form onSubmit={handleCreate} className="flex items-end gap-3">
            <div className="flex-1 space-y-1.5">
              <label className="text-sm text-zinc-400">Config Name</label>
              <input
                type="text"
                value={name}
                onChange={e => setName(e.target.value)}
                placeholder="e.g. customer-1"
                required
                autoFocus
                className="w-full bg-zinc-800 border border-zinc-700 text-white text-sm rounded-lg px-3 py-2 focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500"
              />
            </div>
            <Button type="submit" disabled={creating || !name.trim()} className="bg-brand-600 hover:bg-brand-500 text-white" size="sm">
              {creating ? 'Creating...' : 'Create'}
            </Button>
            <Button type="button" variant="ghost" onClick={() => { setShowCreate(false); setError(null) }} className="text-zinc-400" size="sm">
              Cancel
            </Button>
          </form>
          {error && (
            <div className="mt-2 text-sm text-red-400 bg-red-400/10 border border-red-400/20 rounded-lg px-3 py-2">
              {error}
            </div>
          )}
        </div>
      )}

      {/* Config list */}
      {connections.length === 0 && !showCreate ? (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-8 text-center">
          <Globe className="w-8 h-8 text-zinc-600 mx-auto mb-2" />
          <p className="text-zinc-500">No OpenVPN configs yet</p>
          <p className="text-zinc-600 text-sm mt-1">Click "Create Config" to generate one.</p>
        </div>
      ) : (
        <div className="space-y-2">
          {connections.map(conn => (
            <div key={conn.id} className="flex items-center justify-between bg-zinc-900 border border-zinc-800 rounded-lg px-4 py-3">
              <div className="flex items-center gap-3">
                <Globe className="w-4 h-4 text-brand-400" />
                <span className="text-sm font-medium text-white">{conn.username}</span>
                <span className="text-xs text-zinc-500">{conn.active ? 'Active' : 'Inactive'}</span>
              </div>
              <div className="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleDownload(conn.id)}
                  className="text-brand-400 hover:text-brand-300 text-xs"
                >
                  Download .ovpn
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleDelete(conn.id)}
                  className="text-zinc-500 hover:text-red-400"
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function PrimaryTab({ device, connections, bandwidth, serverHost, copyToClipboard, copiedId, addConnectionOpen, onAddConnectionOpenChange, onDeleteConnection, onRefresh }: {
  device: Device
  connections: ProxyConnection[]
  bandwidth: DeviceBandwidth | null
  serverHost: string
  copyToClipboard: (text: string, id: string) => void
  copiedId: string | null
  addConnectionOpen: boolean
  onAddConnectionOpenChange: (open: boolean) => void
  onDeleteConnection: (id: string) => void
  onRefresh: () => void
}) {
  const [subTab, setSubTab] = useState<'proxy' | 'info'>('proxy')

  return (
    <div>
      {/* Sub-tabs */}
      <div className="flex gap-1 border-b border-zinc-800 mb-6">
        <button
          onClick={() => setSubTab('proxy')}
          className={cn('px-4 py-2 text-sm border-b-2 -mb-px transition-colors',
            subTab === 'proxy' ? 'border-brand-500 text-white' : 'border-transparent text-zinc-400 hover:text-white'
          )}
        >
          Connections
        </button>
        <button
          onClick={() => setSubTab('info')}
          className={cn('px-4 py-2 text-sm border-b-2 -mb-px transition-colors',
            subTab === 'info' ? 'border-brand-500 text-white' : 'border-transparent text-zinc-400 hover:text-white'
          )}
        >
          Basic Info
        </button>
      </div>

      {subTab === 'proxy' && (
        <div>
          {/* Section header with Add Connection button */}
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-medium text-zinc-400">Proxy Connections</h3>
            <Button
              onClick={() => onAddConnectionOpenChange(true)}
              size="sm"
              className="bg-brand-600 hover:bg-brand-500 text-white flex items-center gap-1.5"
            >
              <Plus className="w-4 h-4" />
              Add Connection
            </Button>
          </div>

          <ConnectionTable
            connections={connections}
            device={device}
            serverHost={serverHost}
            onDelete={onDeleteConnection}
            onRefresh={onRefresh}
          />

          {/* External IP */}
          <div className="mt-6 bg-zinc-900 border border-zinc-800 rounded-lg p-4">
            <div className="text-sm text-zinc-400 mb-2">External IP (Cellular)</div>
            <div className="font-mono text-lg">{device.cellular_ip || 'Unknown'}</div>
          </div>
        </div>
      )}

      {subTab === 'info' && (
        <div className="space-y-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
            <h3 className="text-sm font-medium text-zinc-400 mb-4">Device Information</h3>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <InfoRow label="Name" value={device.name || '-'} />
              <InfoRow label="Model" value={device.device_model || '-'} />
              <InfoRow label="Android ID" value={device.android_id} mono />
              <InfoRow label="Android Version" value={device.android_version || '-'} />
              <InfoRow label="App Version" value={device.app_version || '-'} />
              <InfoRow label="Registered" value={formatDate(device.created_at)} />
            </div>
          </div>

          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
            <h3 className="text-sm font-medium text-zinc-400 mb-4">Network</h3>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <InfoRow label="Cellular IP" value={device.cellular_ip || '-'} mono />
              <InfoRow label="WiFi IP" value={device.wifi_ip || '-'} mono />
              <InfoRow label="VPN IP" value={device.vpn_ip || '-'} mono />
              <InfoRow label="Carrier" value={device.carrier || '-'} />
              <InfoRow label="Network Type" value={device.network_type || '-'} />
              <InfoRow label="Last Heartbeat" value={timeAgo(device.last_heartbeat)} />
            </div>
          </div>

          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
            <h3 className="text-sm font-medium text-zinc-400 mb-4">Ports</h3>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <InfoRow label="HTTP Proxy" value={`${device.http_port}`} mono />
              <InfoRow label="SOCKS5 Proxy" value={`${device.socks5_port}`} mono />
              <InfoRow label="Base Port" value={`${device.base_port}`} mono />
            </div>
          </div>

          {bandwidth && (
            <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
              <h3 className="text-sm font-medium text-zinc-400 mb-4">Bandwidth</h3>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <InfoRow label="Today In" value={formatBytes(bandwidth.today_in)} />
                <InfoRow label="Today Out" value={formatBytes(bandwidth.today_out)} />
                <InfoRow label="Month In" value={formatBytes(bandwidth.month_in)} />
                <InfoRow label="Month Out" value={formatBytes(bandwidth.month_out)} />
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// ============= ADVANCED TAB =============
function AdvancedTab({ device, commands, sendCommand }: {
  device: Device
  commands: DeviceCommand[]
  sendCommand: (type: string, label: string) => void
}) {
  const actionButtons = [
    { type: 'rotate_ip_airplane', label: 'Rotate IP', icon: RotateCw, color: 'bg-brand-600 hover:bg-brand-500', description: 'Change cellular IP via airplane mode toggle' },
    { type: 'find_phone', label: 'Find Phone', icon: Search, color: 'bg-purple-600 hover:bg-purple-700', description: 'Vibrate and flash light' },
    { type: 'reboot', label: 'Reboot', icon: Power, color: 'bg-red-600 hover:bg-red-700', description: 'Reboot the device' },
  ]

  return (
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
  )
}

// ============= CHANGE IP TAB =============
function ChangeIPTab({ device, rotationLinks, onCreateLink, onDeleteLink, getRotationUrl, copyToClipboard, copiedId, onAutoRotateChange }: {
  device: Device
  rotationLinks: RotationLink[]
  onCreateLink: () => void
  onDeleteLink: (id: string) => void
  getRotationUrl: (token: string) => string
  copyToClipboard: (text: string, id: string) => void
  copiedId: string | null
  onAutoRotateChange: (minutes: number) => void
}) {
  return (
    <div>
      {/* Auto-Rotation Interval */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-sm font-medium text-white">Auto-Rotation Interval</h3>
            <p className="text-xs text-zinc-500 mt-1">Automatically rotate IP at a set interval</p>
          </div>
          <div className="flex items-center gap-2">
            <input
              type="number"
              min={0}
              value={device.auto_rotate_minutes}
              onChange={(e) => {
                const v = Math.max(0, Math.floor(Number(e.target.value) || 0))
                onAutoRotateChange(v)
              }}
              className="bg-zinc-800 border border-zinc-700 text-white text-sm rounded-lg px-3 py-2 w-20 focus:outline-none focus:border-brand-500"
            />
            <span className="text-sm text-zinc-400">min</span>
            <span className="text-xs text-zinc-600 ml-1">(0 = disabled)</span>
          </div>
        </div>
      </div>

      {/* Rotation Links */}
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
  )
}

// ============= HISTORY TAB =============
function HistoryTab({ ipHistory }: {
  ipHistory: IPHistoryEntry[]
}) {
  return (
    <div>
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-zinc-800 text-zinc-500 text-left">
              <th className="px-4 py-2 font-medium text-xs">#</th>
              <th className="px-4 py-2 font-medium text-xs">Date</th>
              <th className="px-4 py-2 font-medium text-xs">IP</th>
            </tr>
          </thead>
          <tbody>
            {ipHistory.map((entry, idx) => (
              <tr key={entry.id} className="border-b border-zinc-800/30 hover:bg-zinc-800/20">
                <td className="px-4 py-2 text-zinc-500">{idx + 1}</td>
                <td className="px-4 py-2 text-zinc-300">{formatDate(entry.created_at)}</td>
                <td className="px-4 py-2 font-mono text-xs">{entry.ip}</td>
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
  const now = new Date()
  const today = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
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
