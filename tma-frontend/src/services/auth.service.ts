import api from './api'

export async function loginTelegram(initData: string) {
  const { data } = await api.post('/auth/telegram', { initData })
  localStorage.setItem('token', data.token)
  localStorage.setItem('user', JSON.stringify(data.user))
  return data
}

export function logout() {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
}

export function getToken(): string | null {
  return localStorage.getItem('token')
}

export function getStoredUser() {
  const u = localStorage.getItem('user')
  return u ? JSON.parse(u) : null
}
