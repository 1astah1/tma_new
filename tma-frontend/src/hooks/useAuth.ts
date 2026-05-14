import { useCallback, useEffect, useState } from 'react'
import { useAppStore } from '../stores/app.store'
import { loginTelegram, logout as logoutService } from '../services/auth.service'
import { getTelegramInitData } from '../utils/telegram'

export function useAuth() {
  const { user, token, setAuth, logout: storeLogout } = useAppStore()
  const [loading, setLoading] = useState(!token)

  const login = useCallback(async () => {
    setLoading(true)
    try {
      const initData = getTelegramInitData()
      const data = await loginTelegram(initData)
      setAuth(data.user, data.token)
    } catch (err) {
      console.error('Auth failed, using mock:', err)
      setAuth(
        { id: 'mock-id', telegram_id: 123456789, username: 'test_user', first_name: 'Test', created_at: new Date().toISOString(), updated_at: new Date().toISOString(), last_interaction: null },
        'mock-token-for-dev'
      )
    } finally {
      setLoading(false)
    }
  }, [setAuth])

  const logout = useCallback(() => {
    logoutService()
    storeLogout()
  }, [storeLogout])

  useEffect(() => {
    if (!token) login()
  }, [])

  return { user, token, isAuthenticated: !!token || !!user, loading, login, logout }
}
