import { AuthProvider } from 'react-admin'

export const authProvider: AuthProvider = {
  login: async ({ username, password }) => {
    // Support both telegram_id and username "admin"/"superadmin"
    let telegramId: number
    if (username === 'admin' || username === 'superadmin') {
      telegramId = 111111
    } else {
      telegramId = parseInt(username)
      if (isNaN(telegramId)) {
        throw new Error('Введите Telegram ID (например: 111111)')
      }
    }

    const res = await fetch('/api/v1/admin/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ telegram_id: telegramId, password }),
    })
    if (!res.ok) {
      const error = await res.json().catch(() => ({ message: 'Login failed' }))
      throw new Error(error.message || 'Неверные учетные данные')
    }
    const data = await res.json()
    localStorage.setItem('token', data.token)
    localStorage.setItem('admin', JSON.stringify(data.admin))
  },
  logout: () => {
    localStorage.removeItem('token')
    localStorage.removeItem('admin')
    return Promise.resolve()
  },
  checkError: (error) => {
    const status = error.status || (error.response && error.response.status)
    if (status === 401 || status === 403) {
      localStorage.removeItem('token')
      localStorage.removeItem('admin')
      return Promise.reject({ redirectTo: '/login' })
    }
    return Promise.resolve()
  },
  checkAuth: () => {
    const token = localStorage.getItem('token')
    if (!token) {
      return Promise.reject({ redirectTo: '/login' })
    }
    return Promise.resolve()
  },
  getPermissions: () => {
    const admin = localStorage.getItem('admin')
    return admin ? Promise.resolve(JSON.parse(admin).roles) : Promise.reject()
  },
}
