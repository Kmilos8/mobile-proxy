'use client'

import Link from 'next/link'
import Image from 'next/image'
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
      <div className="mb-8 flex items-center gap-3">
        <Image src="/logo.jpg" alt="PocketProxy" width={32} height={32} className="rounded-lg" />
        <div>
          <h1 className="text-lg font-bold leading-none">
            <span className="text-brand-400">Pocket</span><span className="text-brand-500">Proxy</span>
          </h1>
          <p className="text-xs text-zinc-500">Dashboard</p>
        </div>
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
                  ? 'bg-brand-500/15 text-brand-400'
                  : 'text-zinc-400 hover:text-white hover:bg-zinc-800/70'
              )}
            >
              <Icon className="w-4 h-4" />
              {item.label}
            </Link>
          )
        })}
      </nav>

      <div className="border-t border-zinc-800 pt-3">
        <button
          onClick={() => { clearAuth(); router.push('/login') }}
          className="flex items-center gap-3 px-3 py-2 text-sm text-zinc-500 hover:text-white hover:bg-zinc-800/70 rounded w-full text-left"
        >
          <LogOut className="w-4 h-4" />
          Sign Out
        </button>
      </div>
    </aside>
  )
}
