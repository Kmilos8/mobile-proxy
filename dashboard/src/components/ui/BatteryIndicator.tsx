import { cn } from '@/lib/utils'

export default function BatteryIndicator({ level, charging }: { level: number; charging: boolean }) {
  const color = level > 50 ? 'text-green-400' : level > 20 ? 'text-yellow-400' : 'text-red-400'
  return (
    <span className={cn('text-sm', color)}>
      {level}%{charging ? ' +' : ''}
    </span>
  )
}
