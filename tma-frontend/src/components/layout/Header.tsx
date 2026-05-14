import { ReactNode, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { showTelegramBackButton } from '../../utils/telegram'

interface Props {
  title?: string
  children?: ReactNode
  onBack?: () => void
  showLogo?: boolean
}

export function Header({ title, children, onBack, showLogo }: Props) {
  const navigate = useNavigate()

  useEffect(() => {
    showTelegramBackButton(!!onBack)
  }, [onBack])

  const [dark, setDark] = useState(() => localStorage.getItem('tma-theme') !== 'light')

  const toggleTheme = () => {
    const next = !dark
    setDark(next)
    localStorage.setItem('tma-theme', next ? 'dark' : 'light')
    document.documentElement.setAttribute('data-theme', next ? 'dark' : 'light')
  }

  useEffect(() => {
    const saved = localStorage.getItem('tma-theme')
    if (saved === 'dark') document.documentElement.setAttribute('data-theme', 'dark')
    else if (saved === 'light') document.documentElement.setAttribute('data-theme', 'light')
  }, [])

  return (
    <div className="sticky top-0 z-10 bg-[var(--tg-bg)] border-b border-[var(--tg-border)] px-4 py-3 backdrop-blur-lg bg-opacity-80">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {onBack && (
            <button onClick={onBack} className="text-[var(--tg-button)] text-lg">←</button>
          )}
          {showLogo && (
            <div className="flex items-center gap-2 cursor-pointer" onClick={() => navigate('/')}>
              <img src="/logo.png" alt="COIN MINT" className="w-auto h-10" />
              <span className="font-bold text-base text-[var(--tg-text)]">COIN MINT</span>
            </div>
          )}
          {title && <h1 className="text-lg font-semibold text-[var(--tg-text)]">{title}</h1>}
        </div>
        <div className="flex items-center gap-2">
          {children}
          <button
            onClick={toggleTheme}
            className="w-9 h-9 rounded-full bg-[var(--tg-secondary)] flex items-center justify-center text-lg hover:scale-110 transition-transform"
            title={dark ? 'Светлая тема' : 'Тёмная тема'}
          >
            {dark ? '☀️' : '🌙'}
          </button>
        </div>
      </div>
    </div>
  )
}
