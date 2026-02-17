'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { LayoutDashboard, Smartphone, LinkIcon, Users, LogOut } from 'lucide-react'
import { cn } from '@/lib/utils'
import { clearAuth } from '@/lib/auth'
import { useRouter } from 'next/navigation'

const navItems = [
  { href: '/overview', label: 'Overview', icon: LayoutDashboard },
  { href: '/devices', label: 'Devices', icon: Smartphone },
  { href: '/connections', label: 'Connections', icon: LinkIcon },
  { href: '/customers', label: 'Customers', icon: Users },
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
        {navItems.map(item => {
          const Icon = item.icon
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex items-center gap-3 px-3 py-2 rounded text-sm',
                pathname.startsWith(item.href)
                  ? 'bg-blue-600/20 text-blue-400'
                  : 'text-zinc-400 hover:text-white hover:bg-zinc-800'
              )}
            >
              <Icon className="w-4 h-4" />
              {item.label}
            </Link>
          )
        })}
      </nav>

      <button
        onClick={() => { clearAuth(); router.push('/login') }}
        className="flex items-center gap-3 px-3 py-2 text-sm text-zinc-500 hover:text-white text-left"
      >
        <LogOut className="w-4 h-4" />
        Sign Out
      </button>
    </aside>
  )
}
