'use client'

import { useState } from 'react'
import { Copy, Download, Trash2 } from 'lucide-react'
import { api, ProxyConnection, Device } from '@/lib/api'
import { getToken } from '@/lib/auth'
import { cn, copyToClipboard } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
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

  async function handleDownloadOVPN(conn: ProxyConnection) {
    const token = getToken()
    if (!token) return
    setDownloadingId(conn.id)
    try {
      await api.connections.downloadOVPN(token, conn.id)
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
    </>
  )
}
