'use client'

import { useState } from 'react'
import { UptimeSegment } from '@/lib/api'

const STATUS_COLORS: Record<string, string> = {
  online: 'bg-green-500',
  offline: 'bg-red-500',
  rotating: 'bg-yellow-500',
  error: 'bg-red-700',
  unknown: 'bg-zinc-800',
}

const STATUS_LABELS: Record<string, string> = {
  online: 'Online',
  offline: 'Offline',
  rotating: 'Rotating',
  error: 'Error',
  unknown: 'Unknown',
}

const BUCKET_MINUTES = 5
const BUCKETS_PER_HOUR = 60 / BUCKET_MINUTES // 12

interface Bucket {
  index: number
  hour: number
  status: string
  startTime: string
  endTime: string
}

function buildBuckets(segments: UptimeSegment[]): Bucket[] {
  const buckets: Bucket[] = []

  const parsed = segments?.length
    ? segments.map(s => ({
        status: s.status,
        start: new Date(s.start_time).getTime(),
        end: new Date(s.end_time).getTime(),
      }))
    : null

  const dayStart = parsed
    ? new Date(new Date(segments[0].start_time).toDateString()).getTime()
    : 0

  for (let i = 0; i < 24 * BUCKETS_PER_HOUR; i++) {
    const startMin = i * BUCKET_MINUTES
    const endMin = startMin + BUCKET_MINUTES
    const hour = Math.floor(startMin / 60)

    let status = 'unknown'
    if (parsed) {
      const bucketMid = dayStart + (startMin + BUCKET_MINUTES / 2) * 60 * 1000
      for (const seg of parsed) {
        if (bucketMid >= seg.start && bucketMid < seg.end) {
          status = seg.status
          break
        }
      }
    }

    const h = Math.floor(startMin / 60)
    const m = startMin % 60
    const h2 = Math.floor(endMin / 60)
    const m2 = endMin % 60

    buckets.push({
      index: i,
      hour,
      status,
      startTime: `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}`,
      endTime: `${h2.toString().padStart(2, '0')}:${m2.toString().padStart(2, '0')}`,
    })
  }

  return buckets
}

interface UptimeTimelineProps {
  segments: UptimeSegment[]
}

export default function UptimeTimeline({ segments }: UptimeTimelineProps) {
  const [hoveredIdx, setHoveredIdx] = useState<number | null>(null)

  const buckets = buildBuckets(segments)

  // Group buckets by hour
  const hours: Bucket[][] = []
  for (let h = 0; h < 24; h++) {
    hours.push(buckets.filter(b => b.hour === h))
  }

  // Stats
  const onlineBuckets = buckets.filter(b => b.status === 'online').length
  const knownBuckets = buckets.filter(b => b.status !== 'unknown').length
  const uptimeHours = ((onlineBuckets * BUCKET_MINUTES) / 60).toFixed(1)
  const uptimePercent = knownBuckets > 0
    ? ((onlineBuckets / knownBuckets) * 100).toFixed(1)
    : '0.0'

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
      <h3 className="text-sm font-medium text-zinc-400 mb-4">Uptime Timeline</h3>

      {/* Grid: 24 hour columns, each with 12 squares */}
      <div className="relative grid grid-cols-24 gap-x-1" style={{ gridTemplateColumns: 'repeat(24, 1fr)' }}>
        {hours.map((hourBuckets, h) => (
          <div key={h} className="flex flex-col items-center">
            {/* 12 squares per hour: 2 rows x 6 cols */}
            <div className="grid grid-cols-6 gap-[2px] w-full">
              {hourBuckets.map((bucket) => {
                const color = STATUS_COLORS[bucket.status] || STATUS_COLORS.unknown
                const isHovered = hoveredIdx === bucket.index
                return (
                  <div
                    key={bucket.index}
                    className={`aspect-square rounded-[2px] cursor-pointer transition-all ${color} ${
                      hoveredIdx !== null && !isHovered ? 'opacity-40' : ''
                    } ${isHovered ? 'ring-1 ring-white scale-125 z-10' : ''}`}
                    onMouseEnter={() => setHoveredIdx(bucket.index)}
                    onMouseLeave={() => setHoveredIdx(null)}
                  />
                )
              })}
            </div>
            {/* Hour label */}
            <span className="text-[9px] text-zinc-600 mt-1 leading-none">{h}</span>
          </div>
        ))}
      </div>

      {/* Tooltip */}
      {hoveredIdx !== null && buckets[hoveredIdx] && (
        <div className="mt-2 text-xs text-zinc-400">
          <span className="font-medium text-zinc-200">
            {STATUS_LABELS[buckets[hoveredIdx].status]}
          </span>
          {' '}
          {buckets[hoveredIdx].startTime} - {buckets[hoveredIdx].endTime}
        </div>
      )}

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
            <div className="w-3 h-3 rounded bg-zinc-800 border border-zinc-700" />
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
