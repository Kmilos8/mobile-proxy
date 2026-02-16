import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'MobileProxy Dashboard',
  description: 'Mobile proxy management dashboard',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  )
}
