import { useEffect } from 'react'
import { expandTelegramApp } from '../utils/telegram'

export function useTelegram() {
  useEffect(() => {
    expandTelegramApp()
  }, [])
}
