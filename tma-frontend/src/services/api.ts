import axios from 'axios'
import { API_URL } from '../config/api'

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

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      if (!import.meta.env.DEV) {
        window.location.reload()
      }
    }
    return Promise.reject(err)
  }
)

export default api
