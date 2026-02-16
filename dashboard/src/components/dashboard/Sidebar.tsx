'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'
import { clearAuth } from '@/lib/auth'
import { useRouter } from 'next/navigation'

const navItems = [
  { href: '/devices', label: 'Devices' },
  { href: '/connections', label: 'Connections' },
  { href: '/customers', label: 'Customers' },
]

export default function Sidebar() {
  const pathname = usePathname()
  const router = useRouter()

  return (
    <aside className="w-64 bg-zinc-900 border-r border-zinc-800 min-h-screen p-4 flex flex-col">
      <div className="mb-8">
        <h1 className="text-xl font-bold">MobileProxy</h1>
        <p className="text-sm text-zinc-500">Dashboard</p>
      </div>

      <nav className="flex-1 space-y-1">
        {navItems.map(item => (
          <Link
            key={item.href}
            href={item.href}
            className={cn(
              'block px-3 py-2 rounded text-sm',
              pathname.startsWith(item.href)
                ? 'bg-blue-600/20 text-blue-400'
                : 'text-zinc-400 hover:text-white hover:bg-zinc-800'
            )}
          >
            {item.label}
          </Link>
        ))}
      </nav>

      <button
        onClick={() => { clearAuth(); router.push('/login') }}
        className="px-3 py-2 text-sm text-zinc-500 hover:text-white text-left"
      >
        Sign Out
      </button>
    </aside>
  )
}
