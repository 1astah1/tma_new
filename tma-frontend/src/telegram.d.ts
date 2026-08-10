interface TelegramWebApp {
  initData: string
  expand(): void
  openTelegramLink(url: string): void
  BackButton: { show(): void; hide(): void }
  MainButton: { show(): void; hide(): void; setText(text: string): void; onClick(fn: () => void): void }
}

interface Window {
  Telegram?: { WebApp: TelegramWebApp }
}
