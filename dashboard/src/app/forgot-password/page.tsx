'use client'

import { useState, FormEvent } from 'react'
import Image from 'next/image'
import { Turnstile } from '@marsidev/react-turnstile'
import { api } from '@/lib/api'

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('')
  const [turnstileToken, setTurnstileToken] = useState('')
  const [loading, setLoading] = useState(false)
  const [submitted, setSubmitted] = useState(false)
  const [error, setError] = useState('')

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await api.customerAuth.forgotPassword(email, turnstileToken)
      setSubmitted(true)
    } catch (err) {
      // Backend always returns 200 regardless of email existence
      // Only show error for unexpected failures
      const msg = err instanceof Error ? err.message : ''
      if (msg.includes('429')) {
        setError('Too many requests. Please wait a moment and try again.')
      } else {
        // Show generic success to prevent email enumeration
        setSubmitted(true)
      }
    } finally {
      setLoading(false)
    }
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

        {submitted ? (
          <div className="text-center">
            <h2 className="text-xl font-bold text-white mb-3">Check your email</h2>
            <p className="text-zinc-400 text-sm mb-6">
              If an account exists with that email, we sent a password reset link. The link expires in 24 hours.
            </p>
            <a href="/login" className="text-brand-400 hover:text-brand-300 text-sm">
              Back to login
            </a>
          </div>
        ) : (
          <>
            <h2 className="text-lg font-semibold text-white mb-1">Reset your password</h2>
            <p className="text-sm text-zinc-400 mb-4">Enter the email address associated with your account</p>

            <form onSubmit={handleSubmit} className="space-y-4">
              {error && (
                <div className="bg-red-900/50 border border-red-800 text-red-200 px-4 py-2 rounded text-sm">
                  {error}
                </div>
              )}
              <div>
                <label className="block text-sm text-zinc-400 mb-1">Email</label>
                <input
                  type="email"
                  value={email}
                  onChange={e => setEmail(e.target.value)}
                  className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500/50"
                  required
                />
              </div>
              {process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY && (
                <Turnstile
                  siteKey={process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY}
                  onSuccess={setTurnstileToken}
                  onError={() => setTurnstileToken('')}
                  onExpire={() => setTurnstileToken('')}
                />
              )}
              <button
                type="submit"
                disabled={loading}
                className="w-full py-2 bg-brand-600 hover:bg-brand-500 disabled:bg-brand-800 text-white rounded font-medium"
              >
                {loading ? 'Sending...' : 'Send reset link'}
              </button>
            </form>

            <div className="mt-4 text-center text-sm">
              <a href="/login" className="text-brand-400 hover:text-brand-300">Back to login</a>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
