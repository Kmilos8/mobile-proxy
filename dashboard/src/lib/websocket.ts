'use client'

type MessageHandler = (data: { type: string; payload: unknown }) => void

let ws: WebSocket | null = null
let handlers: MessageHandler[] = []
let reconnectTimer: ReturnType<typeof setTimeout> | null = null

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws'

export function connectWebSocket() {
  if (ws?.readyState === WebSocket.OPEN) return

  ws = new WebSocket(WS_URL)

  ws.onopen = () => {
    console.log('WebSocket connected')
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
  }

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      handlers.forEach(h => h(data))
    } catch (e) {
      console.error('WebSocket parse error:', e)
    }
  }

  ws.onclose = () => {
    console.log('WebSocket disconnected, reconnecting...')
    reconnectTimer = setTimeout(connectWebSocket, 3000)
  }

  ws.onerror = (err) => {
    console.error('WebSocket error:', err)
    ws?.close()
  }
}

export function addWSHandler(handler: MessageHandler) {
  handlers.push(handler)
  return () => {
    handlers = handlers.filter(h => h !== handler)
  }
}

export function disconnectWebSocket() {
  if (reconnectTimer) clearTimeout(reconnectTimer)
  ws?.close()
  ws = null
  handlers = []
}
