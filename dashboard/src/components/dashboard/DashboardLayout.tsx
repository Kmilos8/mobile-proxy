'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { isAuthenticated } from '@/lib/auth'
import { connectWebSocket, disconnectWebSocket } from '@/lib/websocket'
import Sidebar from './Sidebar'

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
    <div className="flex min-h-screen">
      <Sidebar />
      <main className="flex-1 p-6">
        {children}
      </main>
    </div>
  )
}
