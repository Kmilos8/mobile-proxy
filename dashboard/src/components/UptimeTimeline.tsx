'use client'

import { useState } from 'react'
import { UptimeSegment } from '@/lib/api'

const STATUS_COLORS: Record<string, string> = {
  online: 'bg-green-500',
  offline: 'bg-red-500',
  rotating: 'bg-yellow-500',
  error: 'bg-red-700',
  unknown: 'bg-zinc-700',
}

const STATUS_LABELS: Record<string, string> = {
  online: 'Online',
  offline: 'Offline',
  rotating: 'Rotating',
  error: 'Error',
  unknown: 'Unknown',
}

const BUCKET_MINUTES = 5
const BUCKETS_PER_DAY = (24 * 60) / BUCKET_MINUTES // 288

interface Bucket {
  index: number
  status: string
  startTime: string // HH:MM
  endTime: string   // HH:MM
}

function buildBuckets(segments: UptimeSegment[]): Bucket[] {
  const buckets: Bucket[] = []

  if (!segments || segments.length === 0) {
    for (let i = 0; i < BUCKETS_PER_DAY; i++) {
      const startMin = i * BUCKET_MINUTES
      const endMin = startMin + BUCKET_MINUTES
      buckets.push({
        index: i,
        status: 'unknown',
        startTime: formatMinutes(startMin),
        endTime: formatMinutes(endMin),
      })
    }
    return buckets
  }

  // Parse segment times once
  const parsed = segments.map(s => ({
    status: s.status,
    start: new Date(s.start_time).getTime(),
    end: new Date(s.end_time).getTime(),
  }))

  // Get the day start from the first segment
  const firstStart = new Date(segments[0].start_time)
  const dayStart = new Date(firstStart.getFullYear(), firstStart.getMonth(), firstStart.getDate()).getTime()

  for (let i = 0; i < BUCKETS_PER_DAY; i++) {
    const startMin = i * BUCKET_MINUTES
    const endMin = startMin + BUCKET_MINUTES
    const bucketMid = dayStart + (startMin + BUCKET_MINUTES / 2) * 60 * 1000

    // Find which segment covers the midpoint of this bucket
    let status = 'unknown'
    for (const seg of parsed) {
      if (bucketMid >= seg.start && bucketMid < seg.end) {
        status = seg.status
        break
      }
    }

    buckets.push({
      index: i,
      status,
      startTime: formatMinutes(startMin),
      endTime: formatMinutes(endMin),
    })
  }

  return buckets
}

function formatMinutes(totalMin: number): string {
  const h = Math.floor(totalMin / 60)
  const m = totalMin % 60
  return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}`
}

interface UptimeTimelineProps {
  segments: UptimeSegment[]
}

export default function UptimeTimeline({ segments }: UptimeTimelineProps) {
  const [hoveredIdx, setHoveredIdx] = useState<number | null>(null)

  const buckets = buildBuckets(segments)

  // Calculate uptime stats
  const onlineBuckets = buckets.filter(b => b.status === 'online').length
  const knownBuckets = buckets.filter(b => b.status !== 'unknown').length
  const uptimeHours = ((onlineBuckets * BUCKET_MINUTES) / 60).toFixed(1)
  const uptimePercent = knownBuckets > 0
    ? ((onlineBuckets / knownBuckets) * 100).toFixed(1)
    : '0.0'

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
      <h3 className="text-sm font-medium text-zinc-400 mb-4">Uptime Timeline</h3>

      {/* Timeline - 288 five-minute blocks */}
      <div className="relative">
        <div className="flex h-8 rounded overflow-hidden">
          {buckets.map((bucket) => {
            const color = STATUS_COLORS[bucket.status] || STATUS_COLORS.unknown
            return (
              <div
                key={bucket.index}
                className={`${color} cursor-pointer transition-opacity border-r border-r-zinc-900/20 last:border-r-0 ${hoveredIdx !== null && hoveredIdx !== bucket.index ? 'opacity-50' : ''}`}
                style={{ width: `${100 / BUCKETS_PER_DAY}%` }}
                onMouseEnter={() => setHoveredIdx(bucket.index)}
                onMouseLeave={() => setHoveredIdx(null)}
              />
            )
          })}
        </div>

        {/* Tooltip */}
        {hoveredIdx !== null && buckets[hoveredIdx] && (
          <div
            className="absolute top-10 bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-xs shadow-lg z-10 whitespace-nowrap pointer-events-none"
            style={{
              left: `${(hoveredIdx / BUCKETS_PER_DAY) * 100}%`,
              transform: hoveredIdx > BUCKETS_PER_DAY * 0.75 ? 'translateX(-100%)' : hoveredIdx > BUCKETS_PER_DAY * 0.25 ? 'translateX(-50%)' : undefined,
            }}
          >
            <div className="font-medium text-zinc-200">
              {STATUS_LABELS[buckets[hoveredIdx].status] || buckets[hoveredIdx].status}
            </div>
            <div className="text-zinc-400 mt-1">
              {buckets[hoveredIdx].startTime} - {buckets[hoveredIdx].endTime}
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
            <div className="w-3 h-3 rounded bg-zinc-700" />
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
