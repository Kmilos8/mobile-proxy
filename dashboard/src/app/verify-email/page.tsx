'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Image from 'next/image'
import { api } from '@/lib/api'
import { setAuth } from '@/lib/auth'

type State = 'loading' | 'valid' | 'invalid' | 'no-token' | 'verifying' | 'error'

export default function VerifyEmailPage() {
  const router = useRouter()
  const [token, setToken] = useState('')
  const [state, setState] = useState<State>('loading')
  const [errorMessage, setErrorMessage] = useState('')
  const [verifying, setVerifying] = useState(false)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const t = params.get('token')
    if (!t) {
      setState('no-token')
      return
    }
    setToken(t)

    // GET check token validity
    api.customerAuth.verifyEmailCheck(t)
      .then(res => {
        if (res.valid) {
          setState('valid')
        } else {
          setState('invalid')
        }
      })
      .catch(() => {
        setState('invalid')
      })
  }, [])

  async function handleVerify() {
    if (!token || verifying) return
    setVerifying(true)
    try {
      const res = await api.customerAuth.verifyEmail(token)
      setAuth(res.token, {
        id: res.customer.id,
        email: res.customer.email,
        name: res.customer.name || res.customer.email.split('@')[0],
        role: res.customer.role,
      })
      router.push('/devices')
    } catch (err) {
      setErrorMessage(err instanceof Error ? err.message : 'Verification failed')
      setState('error')
    } finally {
      setVerifying(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center relative">
      {/* Radial glow */}
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <div className="w-[600px] h-[600px] bg-brand-500/5 rounded-full blur-3xl" />
      </div>

      <div className="bg-zinc-900 p-8 rounded-xl border border-zinc-800 w-full max-w-md shadow-glow-sm relative text-center">
        <div className="flex flex-col items-center mb-6">
          <Image src="/logo.svg" alt="PocketProxy" width={48} height={48} className="rounded-xl mb-3" />
          <h1 className="text-2xl font-bold">
            <span className="text-brand-400">Pocket</span><span className="text-brand-500">Proxy</span>
          </h1>
        </div>

        {state === 'loading' && (
          <div className="text-zinc-400">
            <div className="inline-block w-6 h-6 border-2 border-brand-500 border-t-transparent rounded-full animate-spin mb-3" />
            <p>Checking verification link...</p>
          </div>
        )}

        {state === 'valid' && (
          <div>
            <h2 className="text-xl font-bold text-white mb-3">Verify your email</h2>
            <p className="text-zinc-400 text-sm mb-6">
              Click the button below to verify your email address.
            </p>
            <button
              onClick={handleVerify}
              disabled={verifying}
              className="w-full py-2 bg-brand-600 hover:bg-brand-500 disabled:bg-brand-800 text-white rounded font-medium"
            >
              {verifying ? 'Verifying...' : 'Verify Email'}
            </button>
          </div>
        )}

        {state === 'invalid' && (
          <div>
            <h2 className="text-xl font-bold text-white mb-3">Verification link expired</h2>
            <p className="text-zinc-400 text-sm mb-6">
              This link has expired or has already been used.
            </p>
            <a href="/login" className="text-brand-400 hover:text-brand-300 text-sm">
              Back to login
            </a>
          </div>
        )}

        {state === 'no-token' && (
          <div>
            <h2 className="text-xl font-bold text-white mb-3">Invalid verification link</h2>
            <p className="text-zinc-400 text-sm mb-6">
              This verification link is missing a token.
            </p>
            <a href="/login" className="text-brand-400 hover:text-brand-300 text-sm">
              Back to login
            </a>
          </div>
        )}

        {state === 'error' && (
          <div>
            <div className="bg-red-900/50 border border-red-800 text-red-200 px-4 py-2 rounded text-sm mb-4">
              {errorMessage || 'Verification failed. Please try again.'}
            </div>
            <a href="/login" className="text-brand-400 hover:text-brand-300 text-sm">
              Back to login
            </a>
          </div>
        )}
      </div>
    </div>
  )
}
