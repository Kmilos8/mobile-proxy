'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { ArrowUp, ArrowDown, ChevronsUpDown, Search } from 'lucide-react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import StatusBadge from '@/components/ui/StatusBadge'
import { cn } from '@/lib/utils'
import type { Device } from '@/lib/api'

type SortKey = 'name' | 'status' | 'cellular_ip' | 'auto_rotate' | 'connections'
type SortDir = 'asc' | 'desc'
type StatusFilter = 'all' | 'online' | 'offline'

interface DeviceTableProps {
  devices: Device[]
  connectionCounts: Record<string, number>
  connectionIds: Record<string, string>
}

function SortIcon({ column, sortKey, sortDir }: { column: SortKey; sortKey: SortKey; sortDir: SortDir }) {
  if (column !== sortKey) return <ChevronsUpDown className="w-3.5 h-3.5 text-zinc-600 ml-1 inline" />
  if (sortDir === 'asc') return <ArrowUp className="w-3.5 h-3.5 text-brand-400 ml-1 inline" />
  return <ArrowDown className="w-3.5 h-3.5 text-brand-400 ml-1 inline" />
}

export default function DeviceTable({ devices, connectionCounts, connectionIds }: DeviceTableProps) {
  const router = useRouter()
  const [sortKey, setSortKey] = useState<SortKey>('name')
  const [sortDir, setSortDir] = useState<SortDir>('asc')
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [searchQuery, setSearchQuery] = useState('')

  function handleSort(key: SortKey) {
    if (sortKey === key) {
      setSortDir(prev => (prev === 'asc' ? 'desc' : 'asc'))
    } else {
      setSortKey(key)
      setSortDir('asc')
    }
  }

  const filtered = devices.filter(d => {
    if (searchQuery && !(d.name || '').toLowerCase().includes(searchQuery.toLowerCase())) return false
    if (statusFilter === 'online') return d.status === 'online'
    if (statusFilter === 'offline') return d.status !== 'online'
    return true
  })

  const sorted = [...filtered].sort((a, b) => {
    let aVal: string | number = ''
    let bVal: string | number = ''
    if (sortKey === 'name') {
      aVal = (a.name || '').toLowerCase()
      bVal = (b.name || '').toLowerCase()
    } else if (sortKey === 'status') {
      aVal = a.status
      bVal = b.status
    } else if (sortKey === 'cellular_ip') {
      aVal = a.cellular_ip || ''
      bVal = b.cellular_ip || ''
    } else if (sortKey === 'auto_rotate') {
      aVal = a.auto_rotate_minutes || 0
      bVal = b.auto_rotate_minutes || 0
    } else if (sortKey === 'connections') {
      aVal = connectionCounts[a.id] ?? 0
      bVal = connectionCounts[b.id] ?? 0
    }
    if (aVal < bVal) return sortDir === 'asc' ? -1 : 1
    if (aVal > bVal) return sortDir === 'asc' ? 1 : -1
    return 0
  })

  const isOffline = (d: Device) => d.status !== 'online' && d.status !== 'rotating'

  return (
    <div>
      {/* Filter controls */}
      <div className="flex items-center gap-3 mb-3">
        <span className="text-xs text-zinc-500">Status:</span>
        {(['all', 'online', 'offline'] as StatusFilter[]).map(f => (
          <button
            key={f}
            onClick={() => setStatusFilter(f)}
            className={cn(
              'px-3 py-1 rounded text-xs font-medium transition-colors',
              statusFilter === f
                ? 'bg-brand-500/15 text-brand-400 border border-brand-500/30'
                : 'text-zinc-400 hover:text-white hover:bg-zinc-800 border border-transparent'
            )}
          >
            {f.charAt(0).toUpperCase() + f.slice(1)}
          </button>
        ))}
        <div className="relative ml-auto">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-zinc-500 pointer-events-none" />
          <input
            type="text"
            placeholder="Search devices..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-8 pr-3 py-1 bg-zinc-800 border border-zinc-700 text-white text-xs rounded focus:outline-none focus:border-brand-500"
          />
        </div>
        <span className="text-xs text-zinc-600">{sorted.length} device{sorted.length !== 1 ? 's' : ''}</span>
      </div>

      <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow className="border-zinc-800 hover:bg-transparent">
              <TableHead
                className="cursor-pointer select-none text-zinc-400 hover:text-white h-10 px-4"
                onClick={() => handleSort('name')}
              >
                Device Name <SortIcon column="name" sortKey={sortKey} sortDir={sortDir} />
              </TableHead>
              <TableHead
                className="cursor-pointer select-none text-zinc-400 hover:text-white h-10 px-4"
                onClick={() => handleSort('status')}
              >
                Status <SortIcon column="status" sortKey={sortKey} sortDir={sortDir} />
              </TableHead>
              <TableHead
                className="cursor-pointer select-none text-zinc-400 hover:text-white h-10 px-4 hidden md:table-cell"
                onClick={() => handleSort('cellular_ip')}
              >
                IP <SortIcon column="cellular_ip" sortKey={sortKey} sortDir={sortDir} />
              </TableHead>
              <TableHead className="text-zinc-400 h-10 px-4 hidden md:table-cell">
                Conn. ID
              </TableHead>
              <TableHead
                className="text-zinc-400 hover:text-white cursor-pointer select-none h-10 px-4 hidden md:table-cell"
                onClick={() => handleSort('auto_rotate')}
              >
                Auto-Rotate <SortIcon column="auto_rotate" sortKey={sortKey} sortDir={sortDir} />
              </TableHead>
              <TableHead
                className="cursor-pointer select-none text-zinc-400 hover:text-white h-10 px-4 text-right"
                onClick={() => handleSort('connections')}
              >
                Connections <SortIcon column="connections" sortKey={sortKey} sortDir={sortDir} />
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {sorted.length === 0 ? (
              <TableRow className="border-zinc-800 hover:bg-transparent">
                <TableCell colSpan={6} className="text-center text-zinc-500 py-10">
                  No devices match the current filter.
                </TableCell>
              </TableRow>
            ) : (
              sorted.map(device => (
                <TableRow
                  key={device.id}
                  className={cn(
                    'border-zinc-800 cursor-pointer transition-colors',
                    isOffline(device)
                      ? 'opacity-50 hover:opacity-70 hover:bg-zinc-800/30'
                      : 'hover:bg-zinc-800/50'
                  )}
                  onClick={() => router.push(`/devices/${device.id}`)}
                >
                  <TableCell className="px-4 py-3 font-medium text-white">
                    {device.name || 'Unnamed Device'}
                  </TableCell>
                  <TableCell className="px-4 py-3">
                    <StatusBadge status={device.status} />
                  </TableCell>
                  <TableCell className="px-4 py-3 font-mono text-xs text-zinc-400 hidden md:table-cell">
                    {device.cellular_ip || '-'}
                  </TableCell>
                  <TableCell className="px-4 py-3 font-mono text-xs text-zinc-500 hidden md:table-cell">
                    {connectionIds[device.id] ? connectionIds[device.id].slice(0, 8) : <span className="text-zinc-600">&mdash;</span>}
                  </TableCell>
                  <TableCell className="px-4 py-3 text-xs hidden md:table-cell">
                    {device.auto_rotate_minutes > 0
                      ? <span className="text-emerald-400">every {device.auto_rotate_minutes}m</span>
                      : <span className="text-zinc-600">&mdash;</span>}
                  </TableCell>
                  <TableCell className="px-4 py-3 text-right text-zinc-300">
                    {connectionCounts[device.id] ?? 0}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
