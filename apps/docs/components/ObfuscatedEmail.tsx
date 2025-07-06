'use client'

import { useState } from 'react'

interface ObfuscatedEmailProps {
  user: string
  domain: string
  className?: string
}

export function ObfuscatedEmail({ user, domain, className }: ObfuscatedEmailProps) {
  const [revealed, setRevealed] = useState(false)
  const email = `${user}@${domain}`
  
  if (revealed) {
    return (
      <a href={`mailto:${email}`} className={className}>
        {email}
      </a>
    )
  }
  
  return (
    <button
      onClick={() => setRevealed(true)}
      className={className || 'text-primary underline hover:no-underline'}
      title="Click to reveal email"
    >
      {user}[at]{domain}
    </button>
  )
}