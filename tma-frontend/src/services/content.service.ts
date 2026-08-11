import api from './api'

export interface ShopSettings {
  support_url: string
  reviews_url: string
  shop_rules: string
  /** Личка менеджера, куда уходит заявка на покупку. */
  manager_url: string
}

export interface FaqItem {
  id: string
  question: string
  answer: string
  sort_order: number
}

export async function getShopSettings() {
  const { data } = await api.get('/content/shop-settings')
  return data as ShopSettings
}

export async function getFaq() {
  const { data } = await api.get('/faq')
  return (data.data ?? data) as FaqItem[]
}
