import { formatBytes } from '@/lib/utils'

interface BandwidthBarProps {
  used: number
  limit: number
}

export default function BandwidthBar({ used, limit }: BandwidthBarProps) {
  if (limit <= 0) {
    return <span className="text-xs text-zinc-400">{formatBytes(used)} / Unlimited</span>
  }

  const pct = Math.min((used / limit) * 100, 100)
  const color = pct >= 90 ? 'bg-red-500' : pct >= 70 ? 'bg-yellow-500' : 'bg-brand-500'

  return (
    <div className="w-full">
      <div className="flex justify-between text-xs text-zinc-400 mb-1">
        <span>{formatBytes(used)}</span>
        <span>{formatBytes(limit)}</span>
      </div>
      <div className="w-full h-2 bg-zinc-800 rounded-full overflow-hidden">
        <div
          className={`h-full rounded-full ${color}`}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  )
}
