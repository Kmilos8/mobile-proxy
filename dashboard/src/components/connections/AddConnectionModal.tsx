'use client'

import { useState } from 'react'
import { api } from '@/lib/api'
import { getToken } from '@/lib/auth'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface AddConnectionModalProps {
  deviceId: string
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreated: () => void
}

function generateUsername(): string {
  return `user_${Math.random().toString(36).slice(2, 8)}`
}

function generatePassword(): string {
  return Math.random().toString(36).slice(2, 14)
}

export default function AddConnectionModal({ deviceId, open, onOpenChange, onCreated }: AddConnectionModalProps) {
  const [proxyType, setProxyType] = useState<'http' | 'socks5'>('http')
  const [username, setUsername] = useState(() => generateUsername())
  const [password, setPassword] = useState(() => generatePassword())
  const [bandwidthLimitGB, setBandwidthLimitGB] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function handleOpenChange(newOpen: boolean) {
    if (!loading) {
      if (!newOpen) {
        // Reset form when closing
        setProxyType('http')
        setUsername(generateUsername())
        setPassword(generatePassword())
        setBandwidthLimitGB('')
        setError(null)
      }
      onOpenChange(newOpen)
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const token = getToken()
    if (!token) {
      setError('Not authenticated')
      return
    }
    setLoading(true)
    setError(null)
    // Convert GB to bytes (0 = unlimited)
    const bandwidthLimit = bandwidthLimitGB
      ? Math.round(parseFloat(bandwidthLimitGB) * 1024 * 1024 * 1024)
      : 0
    try {
      await api.connections.create(token, {
        device_id: deviceId,
        username,
        password,
        proxy_type: proxyType as string,
        bandwidth_limit: bandwidthLimit,
      })
      onOpenChange(false)
      // Reset for next open
      setProxyType('http')
      setUsername(generateUsername())
      setPassword(generatePassword())
      onCreated()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create connection')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="bg-zinc-900 border-zinc-800 text-white sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Add Connection</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 mt-2">
          {/* Protocol selector */}
          <div className="space-y-1.5">
            <label className="text-sm text-zinc-400">Protocol</label>
            <Select
              value={proxyType}
              onValueChange={(val) => setProxyType(val as 'http' | 'socks5')}
            >
              <SelectTrigger className="bg-zinc-800 border-zinc-700 text-white focus:ring-brand-500">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="bg-zinc-800 border-zinc-700">
                <SelectItem value="http" className="text-white focus:bg-zinc-700">HTTP</SelectItem>
                <SelectItem value="socks5" className="text-white focus:bg-zinc-700">SOCKS5</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Username */}
          <div className="space-y-1.5">
            <label className="text-sm text-zinc-400">Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              className="w-full bg-zinc-800 border border-zinc-700 text-white text-sm rounded-lg px-3 py-2 focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500"
            />
          </div>

          {/* Password */}
          <div className="space-y-1.5">
            <label className="text-sm text-zinc-400">Password</label>
            <input
              type="text"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              className="w-full bg-zinc-800 border border-zinc-700 text-white text-sm rounded-lg px-3 py-2 focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500"
            />
          </div>

          {/* Bandwidth Limit */}
          <div className="space-y-1.5">
            <label className="text-sm text-zinc-400">Bandwidth Limit (GB) <span className="text-zinc-600">— optional, leave blank for unlimited</span></label>
            <input
              type="number"
              min="0"
              step="0.1"
              value={bandwidthLimitGB}
              onChange={(e) => setBandwidthLimitGB(e.target.value)}
              placeholder="e.g. 10 for 10 GB, blank = unlimited"
              className="w-full bg-zinc-800 border border-zinc-700 text-white text-sm rounded-lg px-3 py-2 focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500 placeholder-zinc-600"
            />
          </div>

          {/* Error message */}
          {error && (
            <div className="text-sm text-red-400 bg-red-400/10 border border-red-400/20 rounded-lg px-3 py-2">
              {error}
            </div>
          )}

          {/* Actions */}
          <div className="flex justify-end gap-2 pt-2">
            <Button
              type="button"
              variant="ghost"
              onClick={() => handleOpenChange(false)}
              disabled={loading}
              className="text-zinc-400 hover:text-white"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={loading || !username || !password}
              className="bg-brand-600 hover:bg-brand-500 text-white"
            >
              {loading ? 'Creating...' : 'Create Connection'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
