import { ButtonHTMLAttributes, ReactNode } from 'react'

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'danger'
  size?: 'sm' | 'md' | 'lg'
  loading?: boolean
  fullWidth?: boolean
  children: ReactNode
}

export function Button({ variant = 'primary', size = 'md', loading, fullWidth, children, className = '', ...props }: Props) {
  const base = 'rounded-lg font-medium transition-all disabled:opacity-50 inline-flex items-center justify-center gap-2'
  const sizes = { sm: 'px-3 py-1.5 text-sm', md: 'px-4 py-2.5 text-base', lg: 'px-6 py-3 text-lg' }
  const variants = {
    primary: 'bg-tg-button text-tg-button-text hover:opacity-90',
    secondary: 'bg-tg-secondary text-tg-text hover:opacity-80',
    outline: 'border-2 border-tg-button text-tg-button bg-transparent',
    danger: 'bg-red-500 text-white hover:bg-red-600',
  }
  return (
    <button
      className={`${base} ${sizes[size]} ${variants[variant]} ${fullWidth ? 'w-full' : ''} ${className}`}
      disabled={loading || props.disabled}
      {...props}
    >
      {loading && <Loader size="sm" />}
      {children}
    </button>
  )
}

export function Loader({ size = 'md' }: { size?: 'sm' | 'md' | 'lg' }) {
  const sizes = { sm: 'w-4 h-4', md: 'w-6 h-6', lg: 'w-8 h-8' }
  return (
    <div className={`${sizes[size]} border-2 border-tg-button border-t-transparent rounded-full animate-spin`} />
  )
}
