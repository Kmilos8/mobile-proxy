'use client'

import { useState } from 'react'
import { Copy, Download, RefreshCw, RotateCcw, Trash2 } from 'lucide-react'
import { api, ProxyConnection, Device } from '@/lib/api'
import { getToken } from '@/lib/auth'
import { cn, copyToClipboard } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogCancel,
} from '@/components/ui/alert-dialog'
import BandwidthBar from '@/components/ui/BandwidthBar'
import DeleteConnectionDialog from './DeleteConnectionDialog'

interface ConnectionTableProps {
  connections: ProxyConnection[]
  device: Device
  serverHost: string
  onDelete: (id: string) => void
  onRefresh: () => void
}

type CopyKey = string

function getPort(conn: ProxyConnection): string {
  if (conn.proxy_type === 'http') return conn.http_port != null ? String(conn.http_port) : '-'
  if (conn.proxy_type === 'socks5') return conn.socks5_port != null ? String(conn.socks5_port) : '-'
  if (conn.proxy_type === 'openvpn') return '1195'
  return '-'
}

function getProtocolLabel(type: ProxyConnection['proxy_type']): string {
  if (type === 'http') return 'HTTP'
  if (type === 'socks5') return 'SOCKS5'
  if (type === 'openvpn') return 'OpenVPN'
  return type
}

function getProtocolBadgeVariant(type: ProxyConnection['proxy_type']): 'default' | 'secondary' | 'outline' {
  if (type === 'http') return 'default'
  if (type === 'socks5') return 'secondary'
  if (type === 'openvpn') return 'outline'
  return 'default'
}

export default function ConnectionTable({ connections, device, serverHost, onDelete, onRefresh }: ConnectionTableProps) {
  const [copiedKey, setCopiedKey] = useState<CopyKey | null>(null)
  const [downloadingId, setDownloadingId] = useState<string | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<ProxyConnection | null>(null)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [regeneratingId, setRegeneratingId] = useState<string | null>(null)
  const [regeneratedPassword, setRegeneratedPassword] = useState<string | null>(null)
  const [regeneratedConnId, setRegeneratedConnId] = useState<string | null>(null)
  const [regenDialogOpen, setRegenDialogOpen] = useState(false)
  const [copiedRegen, setCopiedRegen] = useState(false)
  const [resettingBandwidthId, setResettingBandwidthId] = useState<string | null>(null)

  function handleCopy(text: string, key: CopyKey) {
    copyToClipboard(text)
    setCopiedKey(key)
    setTimeout(() => setCopiedKey(null), 2000)
  }

  function CopyButton({ text, copyKey }: { text: string; copyKey: CopyKey }) {
    const isCopied = copiedKey === copyKey
    return (
      <button
        onClick={() => handleCopy(text, copyKey)}
        className="ml-1 inline-flex items-center text-zinc-500 hover:text-white transition-colors"
        title="Copy"
        type="button"
      >
        {isCopied
          ? <span className="text-green-400 text-xs">Copied!</span>
          : <Copy className="w-3 h-3" />}
      </button>
    )
  }

  async function handleDownloadOVPN(conn: ProxyConnection, password?: string) {
    const token = getToken()
    if (!token) return
    setDownloadingId(conn.id)
    try {
      await api.connections.downloadOVPN(token, conn.id, password)
    } catch (err) {
      console.error('Failed to download .ovpn:', err)
    } finally {
      setDownloadingId(null)
    }
  }

  function handleDeleteClick(conn: ProxyConnection) {
    setDeleteTarget(conn)
    setDeleteOpen(true)
  }

  function handleDeleteConfirm(id: string) {
    setDeleteOpen(false)
    setDeleteTarget(null)
    onDelete(id)
  }

  async function handleRegeneratePassword(conn: ProxyConnection) {
    const token = getToken()
    if (!token) return
    setRegeneratingId(conn.id)
    try {
      const result = await api.connections.regeneratePassword(token, conn.id)
      setRegeneratedPassword(result.password)
      setRegeneratedConnId(conn.id)
      setCopiedRegen(false)
      setRegenDialogOpen(true)
    } catch (err) {
      console.error('Failed to regenerate password:', err)
    } finally {
      setRegeneratingId(null)
    }
  }

  function handleCopyRegenPassword() {
    if (regeneratedPassword) {
      copyToClipboard(regeneratedPassword)
      setCopiedRegen(true)
      setTimeout(() => setCopiedRegen(false), 2000)
    }
  }

  function handleRegenDialogClose() {
    setRegenDialogOpen(false)
    setRegeneratedPassword(null)
    setRegeneratedConnId(null)
    setCopiedRegen(false)
  }

  async function handleResetBandwidth(conn: ProxyConnection) {
    const token = getToken()
    if (!token) return
    setResettingBandwidthId(conn.id)
    try {
      await api.connections.resetBandwidth(token, conn.id)
      onRefresh()
    } catch (err) {
      console.error('Failed to reset bandwidth:', err)
    } finally {
      setResettingBandwidthId(null)
    }
  }

  const regenConn = regeneratedConnId
    ? connections.find(c => c.id === regeneratedConnId)
    : null

  if (connections.length === 0) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-8 text-center text-zinc-500 text-sm">
        No connections yet. Click &lsquo;Add Connection&rsquo; to create one.
      </div>
    )
  }

  return (
    <>
      {/* Desktop table — shown at md+ */}
      <div className="hidden md:block bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow className="border-zinc-800 text-zinc-500 hover:bg-transparent">
              <TableHead className="text-xs font-medium text-zinc-500">Type</TableHead>
              <TableHead className="text-xs font-medium text-zinc-500">Host</TableHead>
              <TableHead className="text-xs font-medium text-zinc-500">Port</TableHead>
              <TableHead className="text-xs font-medium text-zinc-500">Username</TableHead>
              <TableHead className="text-xs font-medium text-zinc-500">Password</TableHead>
              <TableHead className="text-xs font-medium text-zinc-500">Usage</TableHead>
              <TableHead className="text-xs font-medium text-zinc-500">Status</TableHead>
              <TableHead className="text-xs font-medium text-zinc-500 text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {connections.map(conn => {
              const port = getPort(conn)
              const copyAllUrl = conn.proxy_type !== 'openvpn'
                ? `${conn.proxy_type}://${conn.username}:${conn.password ?? ''}@${serverHost}:${port}`
                : null

              return (
                <TableRow key={conn.id} className="border-zinc-800/30 hover:bg-zinc-800/20">
                  <TableCell>
                    <Badge variant={getProtocolBadgeVariant(conn.proxy_type)}>
                      {getProtocolLabel(conn.proxy_type)}
                    </Badge>
                  </TableCell>
                  <TableCell className="font-mono text-xs text-zinc-300">
                    {serverHost}
                    <CopyButton text={serverHost} copyKey={`${conn.id}-host`} />
                  </TableCell>
                  <TableCell className="font-mono text-xs text-zinc-300">
                    {port}
                    <CopyButton text={port} copyKey={`${conn.id}-port`} />
                  </TableCell>
                  <TableCell className="font-mono text-xs text-zinc-300">
                    {conn.username}
                    <CopyButton text={conn.username} copyKey={`${conn.id}-user`} />
                  </TableCell>
                  <TableCell className="font-mono text-xs text-zinc-300">
                    {conn.password}
                    {conn.password && <CopyButton text={conn.password} copyKey={`${conn.id}-pass`} />}
                  </TableCell>
                  <TableCell className="min-w-[120px]">
                    <BandwidthBar used={conn.bandwidth_used} limit={conn.bandwidth_limit} />
                  </TableCell>
                  <TableCell>
                    <span className={cn('text-xs', conn.active ? 'text-green-400' : 'text-zinc-500')}>
                      {conn.active ? 'Active' : 'Disabled'}
                    </span>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center justify-end gap-1">
                      {copyAllUrl && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleCopy(copyAllUrl, `${conn.id}-all`)}
                          className="text-xs h-7 px-2 text-zinc-400 hover:text-white"
                        >
                          {copiedKey === `${conn.id}-all`
                            ? <span className="text-green-400">Copied!</span>
                            : 'Copy All'}
                        </Button>
                      )}
                      {conn.proxy_type === 'openvpn' && (
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleDownloadOVPN(conn)}
                          disabled={downloadingId === conn.id}
                          className="h-7 w-7 text-zinc-400 hover:text-white"
                          title="Download .ovpn"
                        >
                          <Download className="w-3.5 h-3.5" />
                        </Button>
                      )}
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleResetBandwidth(conn)}
                        disabled={resettingBandwidthId === conn.id}
                        className="h-7 w-7 text-zinc-400 hover:text-emerald-400"
                        title="Reset usage"
                      >
                        <RotateCcw className={cn('w-3.5 h-3.5', resettingBandwidthId === conn.id && 'animate-spin')} />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleRegeneratePassword(conn)}
                        disabled={regeneratingId === conn.id}
                        className="h-7 w-7 text-zinc-400 hover:text-amber-400"
                        title="Regenerate password"
                      >
                        <RefreshCw className={cn('w-3.5 h-3.5', regeneratingId === conn.id && 'animate-spin')} />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDeleteClick(conn)}
                        className="h-7 w-7 text-zinc-500 hover:text-red-400"
                        title="Delete connection"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      </div>

      {/* Mobile cards — shown below md */}
      <div className="md:hidden space-y-3">
        {connections.map(conn => {
          const port = getPort(conn)
          const copyAllUrl = conn.proxy_type !== 'openvpn'
            ? `${conn.proxy_type}://${conn.username}:${conn.password ?? ''}@${serverHost}:${port}`
            : null

          return (
            <div key={conn.id} className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-3">
              <div className="flex items-center justify-between">
                <Badge variant={getProtocolBadgeVariant(conn.proxy_type)}>
                  {getProtocolLabel(conn.proxy_type)}
                </Badge>
                <span className={cn('text-xs', conn.active ? 'text-green-400' : 'text-zinc-500')}>
                  {conn.active ? 'Active' : 'Disabled'}
                </span>
              </div>

              <div className="space-y-2 text-sm">
                <div className="flex items-center justify-between">
                  <span className="text-zinc-500">Host</span>
                  <span className="font-mono text-xs text-zinc-300 flex items-center">
                    {serverHost}
                    <CopyButton text={serverHost} copyKey={`m-${conn.id}-host`} />
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-zinc-500">Port</span>
                  <span className="font-mono text-xs text-zinc-300 flex items-center">
                    {port}
                    <CopyButton text={port} copyKey={`m-${conn.id}-port`} />
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-zinc-500">Username</span>
                  <span className="font-mono text-xs text-zinc-300 flex items-center">
                    {conn.username}
                    <CopyButton text={conn.username} copyKey={`m-${conn.id}-user`} />
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-zinc-500">Password</span>
                  <span className="font-mono text-xs text-zinc-300 flex items-center">
                    {conn.password}
                    {conn.password && <CopyButton text={conn.password} copyKey={`m-${conn.id}-pass`} />}
                  </span>
                </div>
                <div>
                  <span className="text-zinc-500 block mb-1 text-xs">Usage</span>
                  <BandwidthBar used={conn.bandwidth_used} limit={conn.bandwidth_limit} />
                </div>
              </div>

              <div className="flex items-center justify-end gap-2 pt-1 border-t border-zinc-800">
                {copyAllUrl && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleCopy(copyAllUrl, `m-${conn.id}-all`)}
                    className="text-xs h-7 px-2 text-zinc-400 hover:text-white"
                  >
                    {copiedKey === `m-${conn.id}-all`
                      ? <span className="text-green-400">Copied!</span>
                      : 'Copy All'}
                  </Button>
                )}
                {conn.proxy_type === 'openvpn' && (
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => handleDownloadOVPN(conn)}
                    disabled={downloadingId === conn.id}
                    className="h-7 w-7 text-zinc-400 hover:text-white"
                    title="Download .ovpn"
                  >
                    <Download className="w-3.5 h-3.5" />
                  </Button>
                )}
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => handleResetBandwidth(conn)}
                  disabled={resettingBandwidthId === conn.id}
                  className="h-7 w-7 text-zinc-400 hover:text-emerald-400"
                  title="Reset usage"
                >
                  <RotateCcw className={cn('w-3.5 h-3.5', resettingBandwidthId === conn.id && 'animate-spin')} />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => handleRegeneratePassword(conn)}
                  disabled={regeneratingId === conn.id}
                  className="h-7 w-7 text-zinc-400 hover:text-amber-400"
                  title="Regenerate password"
                >
                  <RefreshCw className={cn('w-3.5 h-3.5', regeneratingId === conn.id && 'animate-spin')} />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => handleDeleteClick(conn)}
                  className="h-7 w-7 text-zinc-500 hover:text-red-400"
                  title="Delete connection"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </Button>
              </div>
            </div>
          )
        })}
      </div>

      <DeleteConnectionDialog
        connection={deleteTarget}
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        onConfirm={handleDeleteConfirm}
      />

      {/* Regenerate password dialog — shows new password once */}
      <AlertDialog open={regenDialogOpen} onOpenChange={open => { if (!open) handleRegenDialogClose() }}>
        <AlertDialogContent className="bg-zinc-900 border border-zinc-800">
          <AlertDialogHeader>
            <AlertDialogTitle className="text-white">New Password Generated</AlertDialogTitle>
            <AlertDialogDescription className="text-zinc-400">
              Save this password now. It will not be shown again.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="my-4">
            <div className="flex items-center gap-2 bg-zinc-800 border border-zinc-700 rounded-md px-3 py-2">
              <span className="font-mono text-sm text-white flex-1 break-all">
                {regeneratedPassword}
              </span>
              <button
                onClick={handleCopyRegenPassword}
                className="text-zinc-400 hover:text-white transition-colors flex-shrink-0"
                title="Copy password"
                type="button"
              >
                {copiedRegen
                  ? <span className="text-green-400 text-xs">Copied!</span>
                  : <Copy className="w-4 h-4" />}
              </button>
            </div>
            {regenConn?.proxy_type === 'openvpn' && regeneratedPassword && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleDownloadOVPN(regenConn, regeneratedPassword)}
                disabled={downloadingId === regenConn.id}
                className="mt-3 w-full text-xs text-zinc-400 hover:text-white border border-zinc-700"
              >
                <Download className="w-3.5 h-3.5 mr-2" />
                Download .ovpn with embedded credentials
              </Button>
            )}
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel
              onClick={handleRegenDialogClose}
              className="bg-zinc-800 border-zinc-700 text-zinc-300 hover:bg-zinc-700 hover:text-white"
            >
              Done
            </AlertDialogCancel>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
