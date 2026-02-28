'use client'

import Link from 'next/link'
import Image from 'next/image'
import { usePathname } from 'next/navigation'
import { Smartphone, ChevronLeft, ChevronRight, LogOut, Users } from 'lucide-react'
import { cn } from '@/lib/utils'
import { clearAuth, isAdmin } from '@/lib/auth'
import { useRouter } from 'next/navigation'
import { useState, useEffect } from 'react'

const adminNavItems = [
  { href: '/devices', label: 'Devices', icon: Smartphone },
  { href: '/admin/customers', label: 'Customers', icon: Users },
]

const customerNavItems = [
  { href: '/devices', label: 'Devices', icon: Smartphone },
]

export default function Sidebar() {
  const pathname = usePathname()
  const router = useRouter()
  const [collapsed, setCollapsed] = useState(false)
  const [navItems, setNavItems] = useState(customerNavItems)

  // Read localStorage only in useEffect to avoid hydration mismatch
  useEffect(() => {
    const isMd = typeof window !== 'undefined' && window.innerWidth < 1024
    if (isMd) {
      setCollapsed(true)
    } else {
      const stored = localStorage.getItem('sidebar-collapsed')
      if (stored !== null) {
        setCollapsed(stored === 'true')
      }
    }
    setNavItems(isAdmin() ? adminNavItems : customerNavItems)
  }, [])

  function toggleCollapsed() {
    setCollapsed(prev => {
      const next = !prev
      localStorage.setItem('sidebar-collapsed', String(next))
      return next
    })
  }

  return (
    <aside
      className={cn(
        'bg-zinc-900 border-r border-zinc-800 min-h-screen flex flex-col transition-all duration-200 flex-shrink-0',
        collapsed ? 'w-16' : 'w-[220px]'
      )}
    >
      {/* Header */}
      <div className={cn(
        'flex items-center border-b border-zinc-800 h-14 px-3',
        collapsed ? 'justify-center' : 'justify-between'
      )}>
        {!collapsed && (
          <div className="flex items-center gap-2 min-w-0">
            <Image src="/logo.svg" alt="PocketProxy" width={28} height={28} className="rounded-md flex-shrink-0" />
            <h1 className="text-sm font-bold leading-none whitespace-nowrap overflow-hidden text-ellipsis">
              <span className="text-brand-400">Pocket</span><span className="text-brand-500">Proxy</span>
            </h1>
          </div>
        )}
        {collapsed && (
          <Image src="/logo.svg" alt="PocketProxy" width={28} height={28} className="rounded-md" />
        )}
        <button
          onClick={toggleCollapsed}
          className={cn(
            'p-1 text-zinc-500 hover:text-white hover:bg-zinc-800 rounded transition-colors flex-shrink-0',
            collapsed && 'hidden'
          )}
          title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          <ChevronLeft className="w-4 h-4" />
        </button>
      </div>

      {/* Expand button when collapsed */}
      {collapsed && (
        <button
          onClick={toggleCollapsed}
          className="mx-auto mt-2 p-1.5 text-zinc-500 hover:text-white hover:bg-zinc-800 rounded transition-colors"
          title="Expand sidebar"
        >
          <ChevronRight className="w-4 h-4" />
        </button>
      )}

      {/* Navigation */}
      <nav className="flex-1 px-2 py-3 space-y-1">
        {navItems.map(item => {
          const Icon = item.icon
          const isActive = pathname.startsWith(item.href)
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex items-center rounded text-sm transition-colors',
                collapsed ? 'justify-center px-2 py-2' : 'gap-3 px-3 py-2',
                isActive
                  ? 'bg-brand-500/15 text-brand-400'
                  : 'text-zinc-400 hover:text-white hover:bg-zinc-800/70'
              )}
              title={collapsed ? item.label : undefined}
            >
              <Icon className="w-4 h-4 flex-shrink-0" />
              {!collapsed && <span>{item.label}</span>}
            </Link>
          )
        })}
      </nav>

      {/* Sign Out */}
      <div className="px-2 pb-3 border-t border-zinc-800 pt-3">
        <button
          onClick={() => { clearAuth(); router.push('/login') }}
          className={cn(
            'flex items-center text-sm text-zinc-500 hover:text-white hover:bg-zinc-800/70 rounded transition-colors w-full',
            collapsed ? 'justify-center px-2 py-2' : 'gap-3 px-3 py-2'
          )}
          title={collapsed ? 'Sign Out' : undefined}
        >
          <LogOut className="w-4 h-4 flex-shrink-0" />
          {!collapsed && <span>Sign Out</span>}
        </button>
      </div>
    </aside>
  )
}
