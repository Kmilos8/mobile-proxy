'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { isAuthenticated } from '@/lib/auth'
import { connectWebSocket, disconnectWebSocket } from '@/lib/websocket'

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter()

  useEffect(() => {
    if (!isAuthenticated()) {
      router.push('/login')
      return
    }
    connectWebSocket()
    return () => disconnectWebSocket()
  }, [router])

  return (
    <div className="min-h-screen bg-zinc-950/50">
      <main className="p-6">
        {children}
      </main>
    </div>
  )
}
