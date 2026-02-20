'use client'

import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import { BandwidthHourly } from '@/lib/api'

function formatBytesShort(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function formatYAxis(bytes: number): string {
  if (bytes === 0) return '0'
  const k = 1024
  if (bytes < k) return bytes + ' B'
  if (bytes < k * k) return (bytes / k).toFixed(0) + ' KB'
  if (bytes < k * k * k) return (bytes / (k * k)).toFixed(1) + ' MB'
  return (bytes / (k * k * k)).toFixed(1) + ' GB'
}

interface BandwidthChartProps {
  data: BandwidthHourly[]
}

export default function BandwidthChart({ data }: BandwidthChartProps) {
  const chartData = data.map(h => ({
    hour: `${h.hour}:00`,
    Download: h.download_bytes,
    Upload: h.upload_bytes,
  }))

  const totalDownload = data.reduce((sum, h) => sum + h.download_bytes, 0)
  const totalUpload = data.reduce((sum, h) => sum + h.upload_bytes, 0)

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
      <h3 className="text-sm font-medium text-zinc-400 mb-4">Hourly Bandwidth</h3>

      <div className="h-64">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={chartData} margin={{ top: 5, right: 5, left: 5, bottom: 5 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
            <XAxis
              dataKey="hour"
              tick={{ fontSize: 10, fill: '#71717a' }}
              tickLine={false}
              interval={2}
            />
            <YAxis
              tick={{ fontSize: 10, fill: '#71717a' }}
              tickLine={false}
              tickFormatter={formatYAxis}
              width={60}
            />
            <Tooltip
              contentStyle={{ backgroundColor: '#18181b', border: '1px solid #3f3f46', borderRadius: '8px' }}
              labelStyle={{ color: '#a1a1aa' }}
              formatter={(value: number | undefined) => formatBytesShort(value ?? 0)}
            />
            <Legend wrapperStyle={{ fontSize: '12px' }} />
            <Bar dataKey="Download" fill="#3b82f6" radius={[2, 2, 0, 0]} stackId="stack" />
            <Bar dataKey="Upload" fill="#22c55e" radius={[2, 2, 0, 0]} stackId="stack" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <div className="flex gap-6 mt-4 text-sm">
        <div>
          <span className="text-zinc-500">Total Download: </span>
          <span className="text-blue-400 font-medium">{formatBytesShort(totalDownload)}</span>
        </div>
        <div>
          <span className="text-zinc-500">Total Upload: </span>
          <span className="text-green-400 font-medium">{formatBytesShort(totalUpload)}</span>
        </div>
        <div>
          <span className="text-zinc-500">Total: </span>
          <span className="text-zinc-200 font-medium">{formatBytesShort(totalDownload + totalUpload)}</span>
        </div>
      </div>
    </div>
  )
}
