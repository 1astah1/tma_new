import { ReactNode, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { showTelegramBackButton } from '../../utils/telegram'
import { useCart } from '../../stores/cartStore'

interface Props {
  title?: string
  children?: ReactNode
  onBack?: () => void
  showLogo?: boolean
  showCart?: boolean
}

function BackIcon() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden>
      <path
        d="M15 18l-6-6 6-6"
        stroke="currentColor"
        strokeWidth="2.25"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}

export function Header({ title, children, onBack, showLogo, showCart }: Props) {
  const navigate = useNavigate()
  const itemCount = useCart((s) => s.getItemCount())

  useEffect(() => {
    if (!onBack) {
      showTelegramBackButton(false)
      return
    }
    showTelegramBackButton(true, onBack)
    return () => showTelegramBackButton(false, onBack)
  }, [onBack])

  return (
    <div className="sticky top-0 z-40 bg-[var(--tg-bg)]/95 px-4 py-3 backdrop-blur-lg">
      <div className="flex min-w-0 items-center justify-between gap-2">
        <div className="flex min-w-0 flex-1 items-center gap-2">
          {onBack ? (
            <button
              type="button"
              onClick={onBack}
              aria-label="Назад"
              className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-amber-500/25 bg-amber-500/10 text-[var(--tg-button)] transition-colors hover:bg-amber-500/20 active:scale-95"
            >
              <BackIcon />
            </button>
          ) : null}
          {showLogo && (
            <div className="flex shrink-0 cursor-pointer items-center gap-2" onClick={() => navigate('/')}>
              <img src="/wiht_logo.png" alt="COIN MINT" className="h-10 w-auto max-h-12 drop-shadow-[0_0_12px_rgba(201,168,76,0.6)] sm:h-12" />
            </div>
          )}
          {title ? <h1 className="truncate text-lg font-semibold text-[var(--tg-text)]">{title}</h1> : null}
        </div>
        <div className="flex items-center gap-2">
          {showCart && (
            <button
              type="button"
              onClick={() => navigate('/cart')}
              aria-label="Корзина"
              className="relative flex h-9 w-9 items-center justify-center rounded-full border border-amber-500/30 bg-amber-500/10 text-lg transition-colors hover:bg-amber-500/20"
            >
              🛒
              {itemCount > 0 && (
                <span className="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-red-500 px-0.5 text-[9px] font-bold text-white">
                  {itemCount > 9 ? '9+' : itemCount}
                </span>
              )}
            </button>
          )}
          {children}
        </div>
      </div>
    </div>
  )
}
