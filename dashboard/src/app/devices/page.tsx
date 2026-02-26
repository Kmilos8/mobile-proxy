'use client'

import { useEffect, useState, useCallback } from 'react'
import Image from 'next/image'
import { useRouter } from 'next/navigation'
import { Plus, X, Copy, Check, Trash2, LogOut } from 'lucide-react'
import { api, Device, PairingCode, ProxyConnection } from '@/lib/api'
import { getToken, clearAuth } from '@/lib/auth'
import { addWSHandler } from '@/lib/websocket'
import { timeAgo, copyToClipboard } from '@/lib/utils'
import { QRCodeSVG } from 'qrcode.react'
import StatBar from '@/components/devices/StatBar'
import DeviceTable from '@/components/devices/DeviceTable'

const SERVER_HOST = process.env.NEXT_PUBLIC_SERVER_HOST || '178.156.240.184'

function formatCodeDisplay(code: string): string {
  if (code.length === 8) return code.slice(0, 4) + '-' + code.slice(4)
  return code
}

function PairingModal({ onClose }: { onClose: () => void }) {
  const [code, setCode] = useState<string | null>(null)
  const [codeId, setCodeId] = useState<string | null>(null)
  const [expiresAt, setExpiresAt] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    async function create() {
      const token = getToken()
      if (!token) return
      try {
        const res = await api.pairingCodes.create(token)
        if (cancelled) return
        setCode(res.code)
        setCodeId(res.id)
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
  }, [])

  const qrValue = code ? `mobileproxy://pair?server=http://${SERVER_HOST}:8080&code=${code}` : ''

  function handleCopy() {
    if (!code) return
    copyToClipboard(formatCodeDisplay(code))
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  // Countdown to expiry
  const [timeLeft, setTimeLeft] = useState('')
  useEffect(() => {
    if (!expiresAt) return
    const timer = setInterval(() => {
      const diff = new Date(expiresAt).getTime() - Date.now()
      if (diff <= 0) {
        setTimeLeft('Expired')
        clearInterval(timer)
        return
      }
      const hours = Math.floor(diff / 3600000)
      const minutes = Math.floor((diff % 3600000) / 60000)
      setTimeLeft(`${hours}h ${minutes}m remaining`)
    }, 1000)
    return () => clearInterval(timer)
  }, [expiresAt])

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-zinc-900 border border-zinc-700 rounded-xl p-6 max-w-md w-full mx-4" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-white">Add Device</h2>
          <button onClick={onClose} className="text-zinc-400 hover:text-white p-1">
            <X className="w-5 h-5" />
          </button>
        </div>

        {loading && (
          <div className="text-center py-8 text-zinc-500">Generating pairing code...</div>
        )}

        {error && (
          <div className="text-center py-8 text-red-400">{error}</div>
        )}

        {code && !loading && (
          <div className="space-y-6">
            <p className="text-sm text-zinc-400">
              Scan this QR code with the PocketProxy app, or enter the code manually.
            </p>

            {/* QR Code */}
            <div className="flex justify-center">
              <div className="bg-white p-4 rounded-lg">
                <QRCodeSVG value={qrValue} size={200} />
              </div>
            </div>

            {/* Text Code */}
            <div className="text-center">
              <div className="text-3xl font-mono font-bold tracking-widest text-white">
                {formatCodeDisplay(code)}
              </div>
              <button
                onClick={handleCopy}
                className="mt-2 inline-flex items-center gap-1 text-sm text-brand-400 hover:text-brand-300"
              >
                {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                {copied ? 'Copied!' : 'Copy code'}
              </button>
            </div>

            {/* Expiry */}
            <div className="text-center text-xs text-zinc-500">
              {timeLeft}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default function DevicesPage() {
  const router = useRouter()
  const [devices, setDevices] = useState<Device[]>([])
  const [loading, setLoading] = useState(true)
  const [showPairingModal, setShowPairingModal] = useState(false)
  const [pairingCodes, setPairingCodes] = useState<PairingCode[]>([])
  const [connectionCounts, setConnectionCounts] = useState<Record<string, number>>({})

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

  const fetchPairingCodes = useCallback(async () => {
    const token = getToken()
    if (!token) return
    try {
      const res = await api.pairingCodes.list(token)
      setPairingCodes((res.pairing_codes || []).filter(pc => !pc.claimed_at && new Date(pc.expires_at) > new Date()))
    } catch (err) {
      console.error('Failed to fetch pairing codes:', err)
    }
  }, [])

  const fetchConnections = useCallback(async () => {
    const token = getToken()
    if (!token) return
    try {
      const res = await api.connections.list(token)
      const counts: Record<string, number> = {}
      for (const conn of res.connections || []) {
        if (conn.active) {
          counts[conn.device_id] = (counts[conn.device_id] || 0) + 1
        }
      }
      setConnectionCounts(counts)
    } catch (err) {
      console.error('Failed to fetch connections:', err)
    }
  }, [])

  useEffect(() => {
    fetchDevices()
    fetchPairingCodes()
    fetchConnections()
    const unsub = addWSHandler((msg) => {
      if (msg.type === 'device_update') {
        const updated = msg.payload as Device
        setDevices(prev => prev.map(d => d.id === updated.id ? updated : d))
      }
    })
    const interval = setInterval(fetchDevices, 15000)
    return () => { unsub(); clearInterval(interval) }
  }, [fetchDevices, fetchPairingCodes, fetchConnections])

  async function handleRevokePairingCode(id: string) {
    const token = getToken()
    if (!token) return
    try {
      await api.pairingCodes.delete(token, id)
      setPairingCodes(prev => prev.filter(pc => pc.id !== id))
    } catch (err) {
      console.error('Failed to revoke pairing code:', err)
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
  const offlineCount = devices.length - onlineCount

  return (
    <div>
      {/* Top bar with branding */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <Image src="/logo.jpg" alt="PocketProxy" width={28} height={28} className="rounded-md" />
          <div>
            <h1 className="text-sm font-bold leading-none">
              <span className="text-brand-400">Pocket</span><span className="text-brand-500">Proxy</span>
            </h1>
            <p className="text-zinc-500 text-xs mt-0.5">
              Device Fleet
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => { setShowPairingModal(true) }}
            className="inline-flex items-center gap-2 px-4 py-2 bg-brand-600 hover:bg-brand-500 text-white text-sm font-medium rounded-lg transition-colors"
          >
            <Plus className="w-4 h-4" />
            Add Device
          </button>
          <button
            onClick={() => { clearAuth(); router.push('/login') }}
            className="inline-flex items-center gap-2 px-3 py-2 text-sm text-zinc-500 hover:text-white hover:bg-zinc-800/70 rounded-lg transition-colors"
          >
            <LogOut className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Stat bar */}
      <StatBar total={devices.length} online={onlineCount} offline={offlineCount} />

      {/* Active pairing codes */}
      {pairingCodes.length > 0 && (
        <div className="mb-4 space-y-2">
          {pairingCodes.map(pc => (
            <div key={pc.id} className="flex items-center justify-between bg-zinc-900 border border-dashed border-zinc-700 rounded-lg px-4 py-3">
              <div className="flex items-center gap-3">
                <div className="w-2.5 h-2.5 rounded-full bg-yellow-500 animate-pulse" />
                <span className="text-sm text-zinc-400">Waiting for device...</span>
                <span className="font-mono text-sm text-white font-medium">{formatCodeDisplay(pc.code)}</span>
                <span className="text-xs text-zinc-600">expires {timeAgo(pc.expires_at)}</span>
              </div>
              <button
                onClick={() => handleRevokePairingCode(pc.id)}
                className="p-1.5 text-zinc-500 hover:text-red-400 transition-colors"
                title="Revoke code"
              >
                <Trash2 className="w-4 h-4" />
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Device table or empty state */}
      {devices.length === 0 && pairingCodes.length === 0 ? (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-12 text-center">
          <p className="text-zinc-500">No devices registered yet</p>
          <p className="text-zinc-600 text-sm mt-1">Click "Add Device" to generate a pairing code.</p>
        </div>
      ) : (
        <DeviceTable devices={devices} connectionCounts={connectionCounts} />
      )}

      {showPairingModal && (
        <PairingModal onClose={() => { setShowPairingModal(false); fetchPairingCodes() }} />
      )}
    </div>
  )
}
