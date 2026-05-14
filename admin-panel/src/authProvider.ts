import { AuthProvider } from 'react-admin'

export const authProvider: AuthProvider = {
  login: async ({ username, password }) => {
    const res = await fetch('/api/v1/admin/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ telegram_id: parseInt(username), password }),
    })
    if (!res.ok) throw new Error('Login failed')
    const data = await res.json()
    localStorage.setItem('token', data.token)
    localStorage.setItem('admin', JSON.stringify(data.admin))
  },
  logout: () => {
    localStorage.removeItem('token')
    localStorage.removeItem('admin')
    return Promise.resolve()
  },
  checkError: ({ status }) => {
    if (status === 401 || status === 403) {
      localStorage.removeItem('token')
      return Promise.reject()
    }
    return Promise.resolve()
  },
  checkAuth: () => {
    return localStorage.getItem('token') ? Promise.resolve() : Promise.reject()
  },
  getPermissions: () => {
    const admin = localStorage.getItem('admin')
    return admin ? Promise.resolve(JSON.parse(admin).roles) : Promise.reject()
  },
}
