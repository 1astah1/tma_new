import { useEffect } from 'react'
import { initTelegramWebApp, expandTelegramApp } from '../utils/telegram'

export function useTelegram() {
  useEffect(() => {
    initTelegramWebApp()
    expandTelegramApp()
  }, [])
}
