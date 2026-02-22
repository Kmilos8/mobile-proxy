import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

export function formatDate(date: string | null): string {
  if (!date) return 'Never'
  return new Date(date).toLocaleString()
}

export function timeAgo(date: string | null): string {
  if (!date) return 'Never'
  const seconds = Math.floor((Date.now() - new Date(date).getTime()) / 1000)
  if (seconds < 60) return `${seconds}s ago`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`
  return `${Math.floor(seconds / 86400)}d ago`
}

export function copyToClipboard(text: string): void {
  if (navigator.clipboard && window.isSecureContext) {
    navigator.clipboard.writeText(text)
  } else {
    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
  }
}
