'use client'

import { useState, useEffect, useCallback } from 'react'
import Image from 'next/image'
import { api } from '@/lib/api'

export default function SignupVerifyPage() {
  const [email, setEmail] = useState('')
  const [resendLoading, setResendLoading] = useState(false)
  const [resendMessage, setResendMessage] = useState('')
  const [resendError, setResendError] = useState('')
  const [cooldown, setCooldown] = useState(0)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const emailParam = params.get('email')
    if (emailParam) setEmail(emailParam)
  }, [])

  useEffect(() => {
    if (cooldown <= 0) return
    const timer = setTimeout(() => setCooldown(c => c - 1), 1000)
    return () => clearTimeout(timer)
  }, [cooldown])

  const handleResend = useCallback(async () => {
    if (!email || cooldown > 0) return
    setResendLoading(true)
    setResendMessage('')
    setResendError('')
    try {
      await api.customerAuth.resendVerification(email)
      setResendMessage('Verification email sent!')
      setCooldown(60)
    } catch {
      setResendError('Failed to resend. Please try again.')
    } finally {
      setResendLoading(false)
    }
  }, [email, cooldown])

  return (
    <div className="min-h-screen flex items-center justify-center relative">
      {/* Radial glow */}
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <div className="w-[600px] h-[600px] bg-brand-500/5 rounded-full blur-3xl" />
      </div>

      <div className="bg-zinc-900 p-8 rounded-xl border border-zinc-800 w-full max-w-md shadow-glow-sm relative text-center">
        <div className="flex flex-col items-center mb-6">
          <Image src="/logo.svg" alt="PocketProxy" width={48} height={48} className="rounded-xl mb-3" />
        </div>

        {/* Mail icon */}
        <div className="flex justify-center mb-4">
          <svg className="w-16 h-16 text-brand-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75" />
          </svg>
        </div>

        <h2 className="text-2xl font-bold text-white mb-3">Check your email</h2>
        <p className="text-zinc-400 text-sm mb-6">
          We sent a verification link to{' '}
          <span className="text-white font-medium">{email || 'your email address'}</span>.
          Click the link in the email to verify your account.
        </p>

        {resendMessage && (
          <div className="bg-green-900/50 border border-green-800 text-green-200 px-4 py-2 rounded text-sm mb-4">
            {resendMessage}
          </div>
        )}
        {resendError && (
          <div className="bg-red-900/50 border border-red-800 text-red-200 px-4 py-2 rounded text-sm mb-4">
            {resendError}
          </div>
        )}

        <button
          onClick={handleResend}
          disabled={resendLoading || cooldown > 0}
          className="w-full py-2 border border-zinc-700 hover:bg-zinc-800 text-white rounded font-medium disabled:opacity-50 disabled:cursor-not-allowed mb-4"
        >
          {resendLoading
            ? 'Sending...'
            : cooldown > 0
            ? `Resend available in ${cooldown}s`
            : 'Resend verification email'}
        </button>

        <a href="/login" className="text-sm text-brand-400 hover:text-brand-300">
          Back to login
        </a>
      </div>
    </div>
  )
}
