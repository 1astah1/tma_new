type TgWebApp = {
  initData?: string
  ready: () => void
  expand: () => void
  viewportHeight?: number
  viewportStableHeight?: number
}

function getTg(): TgWebApp | null {
  return (window as any).Telegram?.WebApp ?? null
}

/** Telegram Mini App — объект WebApp есть даже до появления initData. */
export function isTelegramWebApp(): boolean {
  return !!getTg()
}

export function initTelegramWebApp(): void {
  const tg = getTg()
  if (!tg) return

  const applyViewport = () => {
    const vh = tg.viewportStableHeight || tg.viewportHeight
    if (vh && vh > 0) {
      document.documentElement.style.setProperty('--tg-viewport-height', `${vh}px`)
    }
  }

  tg.ready()
  tg.expand()
  applyViewport()

  const onViewport = tg as TgWebApp & { onEvent?: (e: string, cb: () => void) => void; offEvent?: (e: string, cb: () => void) => void }
  if (onViewport.onEvent) {
    onViewport.onEvent('viewportChanged', applyViewport)
  }
}

/** HashRouter-остатки и мусорный hash от Telegram → нормальный path для BrowserRouter. */
export function normalizeAppRoute(): void {
  const { hash, pathname, search } = window.location

  if (hash.startsWith('#/')) {
    const target = hash.slice(1) || '/'
    window.history.replaceState(null, '', target + search)
    return
  }

  if (hash && hash !== '#') {
    window.history.replaceState(null, '', '/' + search)
    return
  }

  if (pathname === '' || pathname === '/index.html') {
    window.history.replaceState(null, '', '/' + search)
  }
}

export function getTelegramInitData(): string {
  const tg = getTg()
  if (tg?.initData) return tg.initData
  if (isTelegramWebApp()) return ''
  return 'test'
}

export function expandTelegramApp(): void {
  getTg()?.expand()
}

export function showTelegramBackButton(show: boolean, onClick?: () => void): void {
  const tg = getTg() as any
  const back = tg?.BackButton
  if (!back) return

  if (back.isVisible && !show) {
    back.hide()
  }

  if (onClick) {
    back.offClick(onClick)
    if (show) {
      back.onClick(onClick)
    }
  }

  if (show) back.show()
  else back.hide()
}

export function showTelegramMainButton(text: string, show: boolean, onClick?: () => void): void {
  const tg = getTg() as any
  if (tg?.MainButton) {
    tg.MainButton.setText(text)
    if (show) tg.MainButton.show()
    else tg.MainButton.hide()
    if (onClick) tg.MainButton.onClick(onClick)
  }
}
