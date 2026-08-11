import api from './api'
import { Order, OrderStatus } from '../types/order'

/** Админка внутри мини-аппа: вход по обычному телеграм-токену. */
export type AdminStats = {
  orders_today?: number
  orders_total?: number
  revenue_today?: number
  revenue_total?: number
  pending_orders?: number
}

export async function getAdminOrders(status?: string) {
  const params = new URLSearchParams({ limit: '50' })
  if (status) params.set('status', status)
  const { data } = await api.get(`/tma-admin/orders?${params}`)
  return (data.data ?? data) as Order[]
}

export async function getAdminOrder(id: string) {
  const { data } = await api.get(`/tma-admin/orders/${id}`)
  return data
}

export async function updateAdminOrderStatus(id: string, status: OrderStatus, comment?: string) {
  const { data } = await api.patch(`/tma-admin/orders/${id}/status`, { status, comment })
  return data
}

export async function getAdminOrderChat(id: string) {
  const { data } = await api.get(`/tma-admin/orders/${id}/chat`)
  return (data.data ?? data) as { id: string; message: string; sender_type: string; created_at: string }[]
}

export async function sendAdminOrderMessage(id: string, message: string) {
  const { data } = await api.post(`/tma-admin/orders/${id}/chat`, { message })
  return data
}

export async function getAdminStats() {
  const { data } = await api.get('/tma-admin/dashboard')
  return (data.data ?? data) as AdminStats
}
