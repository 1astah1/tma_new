import api from './api'
import { Order, OrderStatus } from '../types/order'
import { Product } from '../types/product'

/** Админка внутри мини-аппа: вход по обычному телеграм-токену. */
const base = '/tma-admin'

export type AdminStats = {
  orders_today?: number
  total_orders?: number
  waiting_payment?: number
  pending_orders?: number
  revenue_today?: number
  pending_revenue?: number
  total_users?: number
  catalog?: { active_game_products?: number; total_imports?: number; pending_imports?: number }
}

export type AdminUser = {
  id: string
  telegram_id: number
  username: string | null
  first_name: string | null
  is_banned: boolean
  created_at: string
}

export type AdminPromo = {
  id: string
  code: string
  discount_percent: number
  is_active: boolean
  usage_limit?: number | null
  used_count?: number
}

function unwrap<T>(data: any): T {
  return (data?.data ?? data) as T
}

/** Пустой список приходит как null — без этого экран падал на .map. */
function unwrapList<T>(data: any): T[] {
  const value = data?.data ?? data
  return Array.isArray(value) ? (value as T[]) : []
}

// ─── заказы ───
export async function getAdminOrders(status?: string) {
  const params = new URLSearchParams({ limit: '50' })
  if (status) params.set('status', status)
  const { data } = await api.get(`${base}/orders?${params}`)
  return unwrapList<Order>(data)
}

export async function updateAdminOrderStatus(id: string, status: OrderStatus, comment?: string) {
  const { data } = await api.patch(`${base}/orders/${id}/status`, { status, comment })
  return data
}

export async function getAdminOrderChat(id: string) {
  const { data } = await api.get(`${base}/orders/${id}/chat`)
  return unwrapList<{ id: string; message: string; sender_type: string; created_at: string }>(data)
}

export async function sendAdminOrderMessage(id: string, message: string) {
  const { data } = await api.post(`${base}/orders/${id}/chat`, { message })
  return data
}

// ─── сводка ───
export async function getAdminStats() {
  const { data } = await api.get(`${base}/dashboard`)
  return unwrap<AdminStats>(data)
}

// ─── товары ───
export async function getAdminProducts(search?: string) {
  const params = new URLSearchParams({ limit: '40' })
  if (search) params.set('search', search)
  const { data } = await api.get(`${base}/products?${params}`)
  return unwrapList<Product>(data)
}

export async function updateAdminProduct(id: string, patch: Record<string, unknown>) {
  const { data } = await api.put(`${base}/products/${id}`, patch)
  return data
}

// ─── настройки ───
export async function getSetting(key: string) {
  const { data } = await api.get(`${base}/settings?key=${encodeURIComponent(key)}`)
  return data?.value ?? null
}

export async function saveSetting(key: string, value: unknown) {
  const { data } = await api.put(`${base}/settings`, { key, value })
  return data
}

// ─── пользователи ───
export async function getAdminUsers(search?: string) {
  const params = new URLSearchParams({ limit: '50' })
  if (search) params.set('search', search)
  const { data } = await api.get(`${base}/users?${params}`)
  return unwrapList<AdminUser>(data)
}

export async function setUserBanned(id: string, banned: boolean) {
  const { data } = await api.patch(`${base}/users/${id}`, { is_banned: banned })
  return data
}

// ─── промокоды ───
export async function getAdminPromos() {
  const { data } = await api.get(`${base}/promos?limit=50`)
  return unwrapList<AdminPromo>(data)
}

export async function createPromo(payload: { code: string; discount_percent: number }) {
  const { data } = await api.post(`${base}/promos`, { ...payload, is_active: true })
  return data
}

export async function deletePromo(id: string) {
  await api.delete(`${base}/promos/${id}`)
}

// ─── каталог ───
export async function getCatalogSummary() {
  const { data } = await api.get(`${base}/catalog-imports/summary`)
  return unwrap<Record<string, number>>(data)
}

export async function getParserStatus() {
  const { data } = await api.get(`${base}/catalog-parser/status`)
  return data as { running?: boolean; current_stage?: string; imported?: number; processed?: number; total?: number }
}

export async function runWantedImport() {
  const { data } = await api.post(`${base}/catalog-imports/import-wanted`, {})
  return data
}

export async function runDeduplicate() {
  const { data } = await api.post(`${base}/catalog-imports/deduplicate`, {})
  return data
}

// ─── рассылка ───
export async function sendBroadcast(message: string) {
  const { data } = await api.post(`${base}/broadcast`, { message })
  return data
}
