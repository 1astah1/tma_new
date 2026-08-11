import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios'
import { API_URL } from '../config/api'
import { getTelegramInitData } from '../utils/telegram'

const api = axios.create({
  baseURL: API_URL,
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const url = config.url || ''
  const isPublicCatalog =
    url.startsWith('/products') ||
    url.startsWith('/content/') ||
    url.startsWith('/platforms') ||
    url.startsWith('/faq')

  if (!isPublicCatalog) {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
  }
  return config
})

/**
 * Токен живёт сутки, а мини-апп у пользователя может висеть открытым дольше.
 * Раньше протухший токен приводил к 401 → localStorage.clear() → reload:
 * пользователь жал «Предзаказ», страница молча перезагружалась, заказ не
 * создавался. Теперь молча перелогиниваемся по свежему initData и повторяем
 * запрос один раз; если и это не вышло — отдаём ошибку наверх, чтобы экран
 * показал внятное сообщение.
 */
let refreshing: Promise<string | null> | null = null

async function refreshToken(): Promise<string | null> {
  if (refreshing) return refreshing

  refreshing = (async () => {
    try {
      const initData = getTelegramInitData()
      if (!initData) return null

      const { data } = await axios.post(
        `${API_URL}/auth/telegram`,
        { initData },
        { headers: { 'Content-Type': 'application/json' } },
      )
      if (!data?.token) return null

      localStorage.setItem('token', data.token)
      if (data.user) localStorage.setItem('user', JSON.stringify(data.user))
      return data.token as string
    } catch {
      return null
    } finally {
      refreshing = null
    }
  })()

  return refreshing
}

type RetriableConfig = InternalAxiosRequestConfig & { _retried?: boolean }

api.interceptors.response.use(
  (res) => res,
  async (err: AxiosError) => {
    const config = err.config as RetriableConfig | undefined
    const isAuthCall = (config?.url || '').startsWith('/auth/')

    if (err.response?.status === 401 && config && !config._retried && !isAuthCall) {
      config._retried = true
      const token = await refreshToken()
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
        return api.request(config)
      }
      localStorage.removeItem('token')
      localStorage.removeItem('user')
    }

    return Promise.reject(err)
  },
)

export default api
