import api from './api'
import { Order } from '../types/order'

export async function createOrder(productId: string, deliveryMethod: 'key' | 'activation') {
  const { data } = await api.post('/orders', { product_id: productId, delivery_method: deliveryMethod })
  return data as Order
}

export async function getMyOrders(status?: string, page = 1, limit = 20) {
  const params = new URLSearchParams()
  if (status) params.set('status', status)
  params.set('page', String(page))
  params.set('limit', String(limit))
  const { data } = await api.get(`/orders?${params}`)
  return data as { data: Order[]; meta: { page: number; limit: number; total: number } }
}

export async function getOrder(id: string) {
  const { data } = await api.get(`/orders/${id}`)
  return data as Order & { history: any[] }
}

export async function confirmPayment(orderId: string, paymentMethod: string, file: File) {
  const form = new FormData()
  form.append('receipt', file)
  form.append('payment_method', paymentMethod)
  const { data } = await api.post(`/orders/${orderId}/confirm-payment`, form)
  return data
}

export async function sendCredentials(orderId: string, platform: string, login: string, password: string) {
  const { data } = await api.post(`/orders/${orderId}/credentials`, { platform, login, password })
  return data
}

export async function send2FACode(orderId: string, code: string) {
  const { data } = await api.post(`/orders/${orderId}/2fa-code`, { code })
  return data
}
