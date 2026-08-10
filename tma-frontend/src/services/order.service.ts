import api from './api'
import { Order } from '../types/order'

export async function createOrder(productId: string, variantId?: string, quantity = 1) {
  const { data } = await api.post('/orders', {
    product_id: productId,
    delivery_method: 'activation',
    variant_id: variantId,
    quantity,
  })
  return data as Order
}

export async function createBatchOrder(
  items: { product_id: string; delivery_method: 'activation'; variant_id?: string; quantity: number }[],
  promoCode?: string,
) {
  const { data } = await api.post('/orders/batch', {
    items,
    promo_code: promoCode || undefined,
  })
  return data as { orders: Order[]; total_amount: number; order_ids: string[] }
}

export async function uploadOrderReceipt(orderId: string, paymentMethod: string, file: File) {
  const form = new FormData()
  form.append('payment_method', paymentMethod)
  form.append('receipt', file)
  const { data } = await api.post(`/orders/${orderId}/payment`, form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return data
}

export async function uploadBatchReceipt(orderIds: string[], paymentMethod: string, file: File) {
  const form = new FormData()
  form.append('payment_method', paymentMethod)
  form.append('order_ids', JSON.stringify(orderIds))
  form.append('receipt', file)
  const { data } = await api.post('/orders/batch/payment', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return data
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

export async function sendCredentials(orderId: string, platform: string, login: string, password: string) {
  const { data } = await api.post(`/orders/${orderId}/credentials`, { platform, login, password })
  return data
}

export async function send2FACode(orderId: string, code: string) {
  const { data } = await api.post(`/orders/${orderId}/2fa-code`, { code })
  return data
}
