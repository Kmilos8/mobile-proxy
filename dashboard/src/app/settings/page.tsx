'use client'

import { useEffect, useState } from 'react'
import { api } from '@/lib/api'
import { getToken } from '@/lib/auth'

export default function SettingsPage() {
  const [webhookUrl, setWebhookUrl] = useState('')
  const [savedUrl, setSavedUrl] = useState('')
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  useEffect(() => {
    async function loadWebhookUrl() {
      const token = getToken()
      if (!token) return
      try {
        const result = await api.settings.getWebhook(token)
        const url = result.webhook_url ?? ''
        setWebhookUrl(url)
        setSavedUrl(url)
      } catch (err) {
        console.error('Failed to load webhook URL:', err)
      }
    }
    loadWebhookUrl()
  }, [])

  async function handleSave() {
    const token = getToken()
    if (!token) return
    setSaving(true)
    setMessage(null)
    try {
      await api.settings.setWebhook(token, webhookUrl)
      setSavedUrl(webhookUrl)
      setMessage({ type: 'success', text: 'Webhook URL saved.' })
    } catch (err) {
      setMessage({ type: 'error', text: err instanceof Error ? err.message : 'Failed to save' })
    } finally {
      setSaving(false)
    }
  }

  async function handleTest() {
    const token = getToken()
    if (!token) return
    if (!webhookUrl) {
      setMessage({ type: 'error', text: 'Enter a webhook URL before testing.' })
      return
    }
    setTesting(true)
    setMessage(null)
    try {
      const result = await api.settings.testWebhook(token, webhookUrl)
      setMessage({ type: 'success', text: `Test sent successfully (HTTP ${result.status}).` })
    } catch (err) {
      setMessage({ type: 'error', text: err instanceof Error ? err.message : 'Test failed' })
    } finally {
      setTesting(false)
    }
  }

  return (
    <div className="max-w-2xl">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-xl font-semibold text-white">Settings</h1>
        <p className="text-sm text-zinc-500 mt-1">Configure monitoring and notification settings.</p>
      </div>

      {/* Webhook card */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6">
        <div className="mb-4">
          <h2 className="text-base font-medium text-white">Webhook Notifications</h2>
          <p className="text-sm text-zinc-500 mt-1">
            Receive notifications when devices go offline or come back online.
          </p>
        </div>

        <div className="space-y-3">
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-1.5">
              Webhook URL
            </label>
            <input
              type="url"
              value={webhookUrl}
              onChange={(e) => setWebhookUrl(e.target.value)}
              placeholder="https://your-server.com/webhook"
              className="w-full bg-zinc-800 border border-zinc-700 text-white text-sm rounded-lg px-3 py-2 focus:outline-none focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 placeholder-zinc-600"
            />
          </div>

          <div className="flex items-center gap-2 pt-1">
            <button
              onClick={handleSave}
              disabled={saving || webhookUrl === savedUrl}
              className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 disabled:bg-zinc-700 disabled:text-zinc-500 text-white text-sm font-medium rounded-lg transition-colors"
            >
              {saving ? 'Saving...' : 'Save'}
            </button>
            <button
              onClick={handleTest}
              disabled={testing || !webhookUrl}
              className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 disabled:bg-zinc-800 disabled:text-zinc-600 text-zinc-300 hover:text-white text-sm font-medium rounded-lg border border-zinc-700 transition-colors"
            >
              {testing ? 'Sending...' : 'Send Test'}
            </button>
          </div>

          {message && (
            <div
              className={`text-sm rounded-lg px-3 py-2 ${
                message.type === 'success'
                  ? 'bg-emerald-500/10 border border-emerald-500/20 text-emerald-400'
                  : 'bg-red-500/10 border border-red-500/20 text-red-400'
              }`}
            >
              {message.text}
            </div>
          )}
        </div>

        <div className="mt-4 pt-4 border-t border-zinc-800">
          <p className="text-xs text-zinc-600">
            The webhook receives POST requests with a JSON payload. Events:{' '}
            <code className="text-zinc-500">device.offline</code>,{' '}
            <code className="text-zinc-500">device.online</code>,{' '}
            <code className="text-zinc-500">test</code>.
          </p>
        </div>
      </div>
    </div>
  )
}
