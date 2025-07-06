'use client'

interface ObfuscatedEmailProps {
  user: string
  domain: string
  className?: string
}

export function ObfuscatedEmail({ user, domain, className }: ObfuscatedEmailProps) {
  const email = `${user}@${domain}`
  const reversedEmail = email.split('').reverse().join('')
  
  const handleClick = () => {
    window.location.href = `mailto:${email}`
  }
  
  return (
    <span 
      className={className}
      style={{
        unicodeBidi: 'bidi-override',
        direction: 'rtl',
        textDecoration: 'underline',
        cursor: 'pointer'
      }}
      onClick={handleClick}
      title="Click to email"
    >
      {reversedEmail}
    </span>
  )
}