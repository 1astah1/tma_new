import { useCallback, useEffect, useState } from 'react'
import { useAppStore } from '../stores/app.store'
import { loginTelegram, logout as logoutService } from '../services/auth.service'
import { getTelegramInitData, isTelegramWebApp } from '../utils/telegram'

const MOCK_USER = {
  id: 'mock-id',
  telegram_id: 123456789,
  username: 'test_user',
  first_name: 'Test',
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  last_interaction: null,
}

function shouldUseMockAuth(): boolean {
  return import.meta.env.DEV && !isTelegramWebApp()
}

async function waitForInitData(timeoutMs = 4000): Promise<string> {
  const deadline = Date.now() + timeoutMs
  while (Date.now() < deadline) {
    const initData = getTelegramInitData()
    if (initData) return initData
    await new Promise((resolve) => setTimeout(resolve, 100))
  }
  return getTelegramInitData()
}

export function useAuth() {
  const { user, token, setAuth, logout: storeLogout } = useAppStore()
  const [authError, setAuthError] = useState<string | null>(null)
  const [authReady, setAuthReady] = useState(
    shouldUseMockAuth() || (!!token && token !== 'mock-token-for-dev'),
  )

  const login = useCallback(async () => {
    setAuthError(null)
    try {
      const initData = await waitForInitData()
      if (!initData) {
        throw new Error('Telegram initData is empty')
      }
      const data = await loginTelegram(initData)
      setAuth(data.user, data.token)
      setAuthReady(true)
    } catch (err: unknown) {
      console.error('Auth failed:', err)
      const message =
        err instanceof Error && 'response' in err
          ? (err as { response?: { data?: { error?: { message?: string } } } }).response?.data?.error
              ?.message
          : null
      if (shouldUseMockAuth()) {
        setAuth(MOCK_USER, 'mock-token-for-dev')
        setAuthReady(true)
      } else {
        setAuthError(
          import.meta.env.DEV && message
            ? `Ошибка входа: ${message}`
            : 'Не удалось войти через Telegram. Перезапустите приложение.',
        )
        storeLogout()
        setAuthReady(false)
      }
    }
  }, [setAuth, storeLogout])

  const logout = useCallback(() => {
    logoutService()
    storeLogout()
    setAuthReady(false)
  }, [storeLogout])

  useEffect(() => {
    if (isTelegramWebApp() && token === 'mock-token-for-dev') {
      storeLogout()
    }

    if (shouldUseMockAuth()) {
      setAuth(MOCK_USER, 'mock-token-for-dev')
      setAuthReady(true)
      return
    }

    login()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return {
    user,
    token,
    isAuthenticated: authReady || !!token || !!user,
    loading: false,
    authError,
    login,
    logout,
  }
}
