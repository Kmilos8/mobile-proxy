import { cn } from '@/lib/utils'

const statusColors: Record<string, string> = {
  online: 'bg-green-500/20 text-green-400 border-green-500/30',
  offline: 'bg-zinc-500/20 text-zinc-400 border-zinc-500/30',
  rotating: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
  error: 'bg-red-500/20 text-red-400 border-red-500/30',
  completed: 'bg-green-500/20 text-green-400 border-green-500/30',
  pending: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
  sent: 'bg-brand-500/20 text-brand-400 border-brand-500/30',
  failed: 'bg-red-500/20 text-red-400 border-red-500/30',
}

export default function StatusBadge({ status }: { status: string }) {
  return (
    <span className={cn('px-2 py-0.5 rounded-full text-xs border', statusColors[status] || statusColors.offline)}>
      {status}
    </span>
  )
}
