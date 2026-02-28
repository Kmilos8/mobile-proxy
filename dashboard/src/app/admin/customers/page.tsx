'use client'

import { useEffect, useState, useCallback } from 'react'
import { api, Customer, CustomerDetail } from '@/lib/api'
import { getToken } from '@/lib/auth'
import { formatDate, formatBytes } from '@/lib/utils'
import { ChevronLeft, CheckCircle, XCircle } from 'lucide-react'

export default function AdminCustomersPage() {
  const [customers, setCustomers] = useState<Customer[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [detail, setDetail] = useState<CustomerDetail | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [actionLoading, setActionLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchCustomers = useCallback(async () => {
    const token = getToken()
    if (!token) return
    try {
      const res = await api.customers.list(token)
      setCustomers(res.customers || [])
    } catch (err) {
      console.error('Failed to fetch customers:', err)
      setError(err instanceof Error ? err.message : 'Failed to load customers')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchCustomers()
  }, [fetchCustomers])

  async function handleSelectCustomer(id: string) {
    if (selectedId === id) {
      setSelectedId(null)
      setDetail(null)
      return
    }
    setSelectedId(id)
    setDetailLoading(true)
    setDetail(null)
    try {
      const token = getToken()
      if (!token) return
      const res = await api.customers.getDetail(token, id)
      setDetail(res)
    } catch (err) {
      console.error('Failed to fetch customer detail:', err)
    } finally {
      setDetailLoading(false)
    }
  }

  async function handleSuspend(id: string) {
    setActionLoading(true)
    try {
      const token = getToken()
      if (!token) return
      await api.customers.suspend(token, id)
      setCustomers(prev => prev.map(c => c.id === id ? { ...c, active: false } : c))
      if (detail && detail.id === id) {
        setDetail(prev => prev ? { ...prev, active: false } : prev)
      }
    } catch (err) {
      console.error('Failed to suspend customer:', err)
    } finally {
      setActionLoading(false)
    }
  }

  async function handleActivate(id: string) {
    setActionLoading(true)
    try {
      const token = getToken()
      if (!token) return
      await api.customers.activate(token, id)
      setCustomers(prev => prev.map(c => c.id === id ? { ...c, active: true } : c))
      if (detail && detail.id === id) {
        setDetail(prev => prev ? { ...prev, active: true } : prev)
      }
    } catch (err) {
      console.error('Failed to activate customer:', err)
    } finally {
      setActionLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-zinc-500">Loading customers...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="bg-red-900/20 border border-red-800 rounded-lg p-4 text-red-400">
        {error}
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-white">Customers</h1>
        <span className="text-sm text-zinc-500">{customers.length} total</span>
      </div>

      <div className="flex gap-6">
        {/* Customer list */}
        <div className="flex-1 bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-zinc-800 text-zinc-400 text-left">
                <th className="px-4 py-3 font-medium">Name</th>
                <th className="px-4 py-3 font-medium">Email</th>
                <th className="px-4 py-3 font-medium">Status</th>
                <th className="px-4 py-3 font-medium">Verified</th>
                <th className="px-4 py-3 font-medium">Created</th>
              </tr>
            </thead>
            <tbody>
              {customers.map(customer => (
                <tr
                  key={customer.id}
                  onClick={() => handleSelectCustomer(customer.id)}
                  className={`border-b border-zinc-800/50 cursor-pointer transition-colors ${
                    selectedId === customer.id
                      ? 'bg-brand-500/10 hover:bg-brand-500/15'
                      : 'hover:bg-zinc-800/30'
                  }`}
                >
                  <td className="px-4 py-3 font-medium text-white">{customer.name}</td>
                  <td className="px-4 py-3 text-zinc-300">{customer.email}</td>
                  <td className="px-4 py-3">
                    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${
                      customer.active
                        ? 'bg-green-900/40 text-green-400 border border-green-800/50'
                        : 'bg-red-900/40 text-red-400 border border-red-800/50'
                    }`}>
                      {customer.active ? 'Active' : 'Suspended'}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    {customer.email_verified
                      ? <CheckCircle className="w-4 h-4 text-green-400" />
                      : <XCircle className="w-4 h-4 text-zinc-600" />
                    }
                  </td>
                  <td className="px-4 py-3 text-zinc-400 text-xs">{formatDate(customer.created_at)}</td>
                </tr>
              ))}
              {customers.length === 0 && (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-zinc-500">
                    No customers yet
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        {/* Detail panel */}
        {selectedId && (
          <div className="w-80 flex-shrink-0">
            <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
              <div className="flex items-center gap-2 mb-4">
                <button
                  onClick={() => { setSelectedId(null); setDetail(null) }}
                  className="p-1 text-zinc-500 hover:text-white transition-colors"
                >
                  <ChevronLeft className="w-4 h-4" />
                </button>
                <h2 className="text-sm font-semibold text-white">Customer Detail</h2>
              </div>

              {detailLoading && (
                <div className="text-zinc-500 text-sm text-center py-4">Loading...</div>
              )}

              {detail && !detailLoading && (
                <div className="space-y-3">
                  <div>
                    <div className="text-xs text-zinc-500 mb-0.5">Name</div>
                    <div className="text-sm text-white font-medium">{detail.name}</div>
                  </div>
                  <div>
                    <div className="text-xs text-zinc-500 mb-0.5">Email</div>
                    <div className="text-sm text-zinc-300">{detail.email}</div>
                  </div>
                  <div>
                    <div className="text-xs text-zinc-500 mb-0.5">Status</div>
                    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${
                      detail.active
                        ? 'bg-green-900/40 text-green-400 border border-green-800/50'
                        : 'bg-red-900/40 text-red-400 border border-red-800/50'
                    }`}>
                      {detail.active ? 'Active' : 'Suspended'}
                    </span>
                  </div>
                  <div>
                    <div className="text-xs text-zinc-500 mb-0.5">Email Verified</div>
                    <div className="text-sm text-zinc-300">{detail.email_verified ? 'Yes' : 'No'}</div>
                  </div>
                  <div>
                    <div className="text-xs text-zinc-500 mb-0.5">Signed Up</div>
                    <div className="text-sm text-zinc-300">{formatDate(detail.created_at)}</div>
                  </div>
                  <div>
                    <div className="text-xs text-zinc-500 mb-0.5">Last Login</div>
                    <div className="text-sm text-zinc-300">{formatDate(detail.last_login)}</div>
                  </div>

                  <hr className="border-zinc-800" />

                  <div className="grid grid-cols-3 gap-2 text-center">
                    <div>
                      <div className="text-lg font-bold text-white">{detail.device_count}</div>
                      <div className="text-xs text-zinc-500">Devices</div>
                    </div>
                    <div>
                      <div className="text-lg font-bold text-white">{detail.share_count}</div>
                      <div className="text-xs text-zinc-500">Shares</div>
                    </div>
                    <div>
                      <div className="text-sm font-bold text-white">{formatBytes(detail.total_bandwidth)}</div>
                      <div className="text-xs text-zinc-500">BW Total</div>
                    </div>
                  </div>

                  <hr className="border-zinc-800" />

                  <div className="flex gap-2">
                    {detail.active ? (
                      <button
                        onClick={() => handleSuspend(detail.id)}
                        disabled={actionLoading}
                        className="flex-1 px-3 py-1.5 bg-red-900/40 hover:bg-red-800/50 text-red-400 border border-red-800/50 rounded text-xs font-medium transition-colors disabled:opacity-50"
                      >
                        Suspend
                      </button>
                    ) : (
                      <button
                        onClick={() => handleActivate(detail.id)}
                        disabled={actionLoading}
                        className="flex-1 px-3 py-1.5 bg-green-900/40 hover:bg-green-800/50 text-green-400 border border-green-800/50 rounded text-xs font-medium transition-colors disabled:opacity-50"
                      >
                        Activate
                      </button>
                    )}
                  </div>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
