export function getTelegramInitData(): string {
  const tg = (window as any).Telegram?.WebApp
  if (tg?.initData) return tg.initData
  return 'test'
}

export function isTelegramWebApp(): boolean {
  return !!(window as any).Telegram?.WebApp
}

export function expandTelegramApp(): void {
  const tg = (window as any).Telegram?.WebApp
  if (tg) tg.expand()
}

export function showTelegramBackButton(show: boolean): void {
  const tg = (window as any).Telegram?.WebApp
  if (tg) {
    if (show) tg.BackButton.show()
    else tg.BackButton.hide()
  }
}

export function showTelegramMainButton(text: string, show: boolean, onClick?: () => void): void {
  const tg = (window as any).Telegram?.WebApp
  if (tg) {
    tg.MainButton.setText(text)
    if (show) tg.MainButton.show()
    else tg.MainButton.hide()
    if (onClick) tg.MainButton.onClick(onClick)
  }
}
