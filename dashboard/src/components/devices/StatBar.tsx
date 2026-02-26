'use client'

interface StatBarProps {
  total: number
  online: number
  offline: number
}

export default function StatBar({ total, online, offline }: StatBarProps) {
  return (
    <div className="grid grid-cols-3 gap-4 mb-6">
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg px-4 py-3">
        <div className="text-2xl font-bold text-white">{total}</div>
        <div className="text-xs text-zinc-500 mt-0.5">Total Devices</div>
      </div>
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg px-4 py-3">
        <div className="text-2xl font-bold text-green-400">{online}</div>
        <div className="text-xs text-zinc-500 mt-0.5">Online</div>
      </div>
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg px-4 py-3">
        <div className="text-2xl font-bold text-zinc-500">{offline}</div>
        <div className="text-xs text-zinc-500 mt-0.5">Offline</div>
      </div>
    </div>
  )
}
