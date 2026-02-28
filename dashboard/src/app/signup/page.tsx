'use client'

import { useState, FormEvent } from 'react'
import { useRouter } from 'next/navigation'
import Image from 'next/image'
import { Turnstile } from '@marsidev/react-turnstile'
import { api } from '@/lib/api'

export default function SignupPage() {
  const router = useRouter()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [turnstileToken, setTurnstileToken] = useState('')

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')

    if (password.length < 8) {
      setError('Password must be at least 8 characters.')
      return
    }

    setLoading(true)
    try {
      await api.customerAuth.signup(email, password, turnstileToken)
      router.push(`/signup/verify?email=${encodeURIComponent(email)}`)
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Signup failed'
      if (msg.includes('409') || msg.toLowerCase().includes('already exists') || msg.toLowerCase().includes('conflict')) {
        setError('An account with this email already exists.')
      } else {
        setError(msg)
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

        <h2 className="text-lg font-semibold text-white mb-4">Create Account</h2>

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
          <div>
            <label className="block text-sm text-zinc-400 mb-1">Password (min. 8 characters)</label>
            <input
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-white focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500/50"
              required
              minLength={8}
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
            {loading ? 'Creating account...' : 'Create Account'}
          </button>
        </form>

        {/* Divider */}
        <div className="relative my-4">
          <div className="absolute inset-0 flex items-center">
            <div className="w-full border-t border-zinc-700" />
          </div>
          <div className="relative flex justify-center text-xs">
            <span className="bg-zinc-900 px-2 text-zinc-500">or</span>
          </div>
        </div>

        {/* Google sign-up */}
        <a
          href={`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api'}/auth/customer/google`}
          className="w-full py-2 bg-white hover:bg-gray-100 text-gray-800 rounded font-medium flex items-center justify-center gap-2 border border-gray-300"
        >
          <svg className="w-5 h-5" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
            <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
            <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
            <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z" fill="#FBBC05"/>
            <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
          </svg>
          Sign up with Google
        </a>

        {/* Login link */}
        <div className="mt-4 text-center text-sm text-zinc-400">
          Already have an account?{' '}
          <a href="/login" className="text-brand-400 hover:text-brand-300">Log in</a>
        </div>
      </div>
    </div>
  )
}
