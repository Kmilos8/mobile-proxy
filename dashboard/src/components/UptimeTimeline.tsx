'use client'

import { useState } from 'react'
import { UptimeSegment } from '@/lib/api'

const STATUS_COLORS: Record<string, string> = {
  online: 'bg-green-500',
  offline: 'bg-red-500',
  rotating: 'bg-yellow-500',
  error: 'bg-red-700',
  unknown: 'bg-zinc-600',
}

const STATUS_LABELS: Record<string, string> = {
  online: 'Online',
  offline: 'Offline',
  rotating: 'Rotating',
  error: 'Error',
  unknown: 'Unknown',
}

function formatDuration(ms: number): string {
  const totalSec = Math.floor(ms / 1000)
  const hours = Math.floor(totalSec / 3600)
  const minutes = Math.floor((totalSec % 3600) / 60)
  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
}

function formatTime(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })
}

interface UptimeTimelineProps {
  segments: UptimeSegment[]
}

export default function UptimeTimeline({ segments }: UptimeTimelineProps) {
  const [hoveredIdx, setHoveredIdx] = useState<number | null>(null)

  if (!segments || segments.length === 0) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <h3 className="text-sm font-medium text-zinc-400 mb-4">Uptime Timeline</h3>
        <div className="text-sm text-zinc-500 text-center py-4">No uptime data available for this date</div>
      </div>
    )
  }

  // Calculate total day span from segments
  const dayStart = new Date(segments[0].start_time).getTime()
  const dayEnd = new Date(segments[segments.length - 1].end_time).getTime()
  const totalSpan = dayEnd - dayStart

  // Calculate uptime stats
  let uptimeMs = 0
  let totalMs = 0
  for (const seg of segments) {
    const dur = new Date(seg.end_time).getTime() - new Date(seg.start_time).getTime()
    totalMs += dur
    if (seg.status === 'online') {
      uptimeMs += dur
    }
  }
  const uptimeHours = (uptimeMs / 3600000).toFixed(1)
  const uptimePercent = totalMs > 0 ? ((uptimeMs / totalMs) * 100).toFixed(1) : '0.0'

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
      <h3 className="text-sm font-medium text-zinc-400 mb-4">Uptime Timeline</h3>

      {/* Timeline bar */}
      <div className="relative">
        <div className="flex h-8 rounded overflow-hidden">
          {segments.map((seg, idx) => {
            const start = new Date(seg.start_time).getTime()
            const end = new Date(seg.end_time).getTime()
            const width = totalSpan > 0 ? ((end - start) / totalSpan) * 100 : 0
            const color = STATUS_COLORS[seg.status] || STATUS_COLORS.unknown

            return (
              <div
                key={idx}
                className={`${color} relative cursor-pointer transition-opacity ${hoveredIdx !== null && hoveredIdx !== idx ? 'opacity-60' : ''}`}
                style={{ width: `${width}%`, minWidth: width > 0 ? '2px' : '0' }}
                onMouseEnter={() => setHoveredIdx(idx)}
                onMouseLeave={() => setHoveredIdx(null)}
              />
            )
          })}
        </div>

        {/* Tooltip */}
        {hoveredIdx !== null && segments[hoveredIdx] && (
          <div className="absolute top-10 left-1/2 -translate-x-1/2 bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-xs shadow-lg z-10 whitespace-nowrap">
            <div className="font-medium text-zinc-200">
              {STATUS_LABELS[segments[hoveredIdx].status] || segments[hoveredIdx].status}
            </div>
            <div className="text-zinc-400 mt-1">
              {formatTime(segments[hoveredIdx].start_time)} - {formatTime(segments[hoveredIdx].end_time)}
            </div>
            <div className="text-zinc-500">
              Duration: {formatDuration(new Date(segments[hoveredIdx].end_time).getTime() - new Date(segments[hoveredIdx].start_time).getTime())}
            </div>
          </div>
        )}
      </div>

      {/* Hour labels */}
      <div className="flex justify-between mt-1 text-xs text-zinc-600">
        <span>0:00</span>
        <span>6:00</span>
        <span>12:00</span>
        <span>18:00</span>
        <span>24:00</span>
      </div>

      {/* Legend and summary */}
      <div className="flex items-center justify-between mt-4">
        <div className="flex gap-4 text-xs">
          <div className="flex items-center gap-1.5">
            <div className="w-3 h-3 rounded bg-green-500" />
            <span className="text-zinc-400">Online</span>
          </div>
          <div className="flex items-center gap-1.5">
            <div className="w-3 h-3 rounded bg-red-500" />
            <span className="text-zinc-400">Offline</span>
          </div>
          <div className="flex items-center gap-1.5">
            <div className="w-3 h-3 rounded bg-yellow-500" />
            <span className="text-zinc-400">Rotating</span>
          </div>
          <div className="flex items-center gap-1.5">
            <div className="w-3 h-3 rounded bg-zinc-600" />
            <span className="text-zinc-400">Unknown</span>
          </div>
        </div>
        <div className="text-sm">
          <span className="text-zinc-500">Uptime: </span>
          <span className="text-green-400 font-medium">{uptimeHours}h</span>
          <span className="text-zinc-600 mx-1">/</span>
          <span className="text-zinc-400">{uptimePercent}%</span>
        </div>
      </div>
    </div>
  )
}
