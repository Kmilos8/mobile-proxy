'use client'

import { useState, useEffect, FormEvent } from 'react'
import { useRouter } from 'next/navigation'
import Image from 'next/image'
import { api } from '@/lib/api'

export default function ResetPasswordPage() {
  const router = useRouter()
  const [token, setToken] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [noToken, setNoToken] = useState(false)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const t = params.get('token')
    if (!t) {
      setNoToken(true)
    } else {
      setToken(t)
    }
  }, [])

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')

    if (password.length < 8) {
      setError('Password must be at least 8 characters.')
      return
    }
    if (password !== confirmPassword) {
      setError('Passwords do not match.')
      return
    }

    setLoading(true)
    try {
      await api.customerAuth.resetPassword(token, password, confirmPassword)
      router.push('/login?message=password_updated')
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Reset failed'
      if (
        msg.toLowerCase().includes('expired') ||
        msg.toLowerCase().includes('invalid') ||
        msg.toLowerCase().includes('already been used')
      ) {
        setError('This reset link has expired or has already been used. Please request a new one.')
      } else {
        setError(msg)
      }
    } finally {
      setLoading(false)
    }
  }

  if (noToken) {
    return (
      <div className="min-h-screen flex items-center justify-center relative">
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
          <div className="w-[600px] h-[600px] bg-brand-500/5 rounded-full blur-3xl" />
        </div>
        <div className="bg-zinc-900 p-8 rounded-xl border border-zinc-800 w-full max-w-md shadow-glow-sm relative text-center">
          <div className="flex flex-col items-center mb-6">
            <Image src="/logo.svg" alt="PocketProxy" width={48} height={48} className="rounded-xl mb-3" />
          </div>
          <h2 className="text-xl font-bold text-white mb-3">Invalid reset link</h2>
          <p className="text-zinc-400 text-sm mb-6">This reset link is missing a token.</p>
          <a href="/forgot-password" className="text-brand-400 hover:text-brand-300 text-sm">
            Request a new reset link
          </a>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center relative">
      {/* Radial glow */}
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <div className="w-[600px] h-[600px] bg-brand-500/5 rounded-full blur-3xl" />
      </div>

      <div className="bg-zinc-900 p-8 rounded-xl border border-zinc-800 w-full max-w-md shadow-glow-sm relative">
        <div className="flex flex-col items-center mb-6">
          <Image src="/logo.svg" alt="PocketProxy" width={48} height={48} className="rounded-xl mb-3" />
          <h1 className="text-2xl font-bold">
            <span className="text-brand-400">Pocket</span><span className="text-brand-500">Proxy</span>
          </h1>
        </div>

        <h2 className="text-lg font-semibold text-white mb-4">Set new password</h2>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="bg-red-900/50 border border-red-800 text-red-200 px-4 py-2 rounded text-sm">
              {error}
              {(error.includes('expired') || error.includes('already been used')) && (
                <div className="mt-2">
                  <a href="/forgot-password" className="underline text-red-300 hover:text-red-100">
                    Request a new reset link
                  </a>
                </div>
              )}
            </div>
          )}
          <div>
            <label className="block text-sm text-zinc-400 mb-1">New password (min. 8 characters)</label>
            <input
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500/50"
              required
              minLength={8}
            />
          </div>
          <div>
            <label className="block text-sm text-zinc-400 mb-1">Confirm new password</label>
            <input
              type="password"
              value={confirmPassword}
              onChange={e => setConfirmPassword(e.target.value)}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500/50"
              required
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            className="w-full py-2 bg-brand-600 hover:bg-brand-500 disabled:bg-brand-800 text-white rounded font-medium"
          >
            {loading ? 'Resetting...' : 'Reset password'}
          </button>
        </form>
      </div>
    </div>
  )
}
