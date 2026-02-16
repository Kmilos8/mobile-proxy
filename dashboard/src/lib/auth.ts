'use client'

const TOKEN_KEY = 'mobileproxy_token'
const USER_KEY = 'mobileproxy_user'

export function getToken(): string | null {
  if (typeof window === 'undefined') return null
  return localStorage.getItem(TOKEN_KEY)
}

export function setAuth(token: string, user: { id: string; email: string; name: string; role: string }) {
  localStorage.setItem(TOKEN_KEY, token)
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

export function getUser() {
  if (typeof window === 'undefined') return null
  const data = localStorage.getItem(USER_KEY)
  return data ? JSON.parse(data) : null
}

export function clearAuth() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export function isAuthenticated(): boolean {
  return !!getToken()
}
