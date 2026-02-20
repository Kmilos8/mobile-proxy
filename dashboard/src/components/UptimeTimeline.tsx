'use client'

import { useState } from 'react'
import { UptimeSegment } from '@/lib/api'

const STATUS_BG: Record<string, string> = {
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
const ROWS = 12        // 12 five-minute slots per hour
const COLS = 24        // 24 hours

interface Bucket {
  index: number
  row: number          // 0-11  (slot within the hour)
  col: number          // 0-23  (hour)
  status: string
  startTime: string
  endTime: string
}

function pad2(n: number) { return n.toString().padStart(2, '0') }

function buildGrid(segments: UptimeSegment[]): Bucket[] {
  const parsed = segments?.length
    ? segments.map(s => ({
        status: s.status,
        start: new Date(s.start_time).getTime(),
        end: new Date(s.end_time).getTime(),
      }))
    : null

  // The backend returns segments for the correct timezone-adjusted day.
  // dayStart is the earliest segment start (which is midnight in the selected timezone).
  const dayStart = parsed
    ? parsed[0].start
    : 0

  const grid: Bucket[] = []

  // Build row-major: row 0 across all 24 hours, then row 1, etc.
  for (let row = 0; row < ROWS; row++) {
    for (let col = 0; col < COLS; col++) {
      const slotIndex = col * ROWS + row          // absolute 5-min slot (0-287)
      const startMin = slotIndex * BUCKET_MINUTES
      const endMin = startMin + BUCKET_MINUTES

      let status = 'unknown'
      if (parsed) {
        const mid = dayStart + (startMin + BUCKET_MINUTES / 2) * 60 * 1000
        for (const seg of parsed) {
          if (mid >= seg.start && mid < seg.end) {
            status = seg.status
            break
          }
        }
      }

      grid.push({
        index: slotIndex,
        row,
        col,
        status,
        startTime: `${pad2(Math.floor(startMin / 60))}:${pad2(startMin % 60)}`,
        endTime: `${pad2(Math.floor(endMin / 60))}:${pad2(endMin % 60)}`,
      })
    }
  }

  return grid
}

interface UptimeTimelineProps {
  segments: UptimeSegment[]
}

export default function UptimeTimeline({ segments }: UptimeTimelineProps) {
  const [hovered, setHovered] = useState<Bucket | null>(null)

  const grid = buildGrid(segments)

  // Stats
  const onlineCount = grid.filter(b => b.status === 'online').length
  const knownCount = grid.filter(b => b.status !== 'unknown').length
  const uptimeHours = ((onlineCount * BUCKET_MINUTES) / 60).toFixed(1)
  const uptimePercent = knownCount > 0
    ? ((onlineCount / knownCount) * 100).toFixed(1)
    : '0.0'

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-medium text-zinc-400">Uptime Timeline</h3>
        {hovered ? (
          <div className="text-xs text-zinc-400">
            <span className="font-medium text-zinc-200">{STATUS_LABELS[hovered.status]}</span>
            {' '}{hovered.startTime} - {hovered.endTime}
          </div>
        ) : (
          <div className="text-xs text-zinc-600">Hover for details</div>
        )}
      </div>

      {/* Hour labels row */}
      <div
        className="grid gap-[3px] mb-[3px] ml-[30px]"
        style={{ gridTemplateColumns: `repeat(${COLS}, 1fr)` }}
      >
        {Array.from({ length: COLS }, (_, h) => (
          <div key={h} className="text-[10px] text-zinc-500 text-center leading-none">
            {pad2(h)}
          </div>
        ))}
      </div>

      {/* 12 rows x 24 cols grid */}
      <div className="flex">
        {/* Minute labels */}
        <div className="flex flex-col gap-[3px] mr-[3px] w-[27px] flex-shrink-0">
          {Array.from({ length: ROWS }, (_, r) => (
            <div key={r} className="text-[10px] text-zinc-600 text-right leading-none flex items-center justify-end" style={{ height: '14px' }}>
              :{pad2(r * BUCKET_MINUTES)}
            </div>
          ))}
        </div>

        {/* Grid */}
        <div
          className="grid flex-1 gap-[3px]"
          style={{ gridTemplateColumns: `repeat(${COLS}, 1fr)` }}
        >
          {grid.map((bucket) => {
            const bg = STATUS_BG[bucket.status] || STATUS_BG.unknown
            const isHovered = hovered?.index === bucket.index
            return (
              <div
                key={`${bucket.row}-${bucket.col}`}
                className={`rounded-[2px] cursor-pointer transition-opacity ${bg} ${
                  hovered !== null && !isHovered ? 'opacity-40' : ''
                } ${isHovered ? 'ring-1 ring-white' : ''}`}
                style={{ height: '14px' }}
                onMouseEnter={() => setHovered(bucket)}
                onMouseLeave={() => setHovered(null)}
              />
            )
          })}
        </div>
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
            <div className="w-3 h-3 rounded bg-zinc-800 border border-zinc-700" />
            <span className="text-zinc-400">No data</span>
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
