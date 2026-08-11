export interface ProductVariant {
  id: string
  name: string
  price: number
  stock: number
}

export interface ProductEditionOption {
  id: string
  name: string
  price: number
  discount_label?: string
}

export type ProductEditionCatalog = {
  ps_tr?: ProductEditionOption[]
  ps_ua?: ProductEditionOption[]
  xbox?: ProductEditionOption[]
  xbox_tr?: ProductEditionOption[]
}

export interface ProductPrices {
  tr?: number
  ua?: number
  xbox?: number
  xbox_tr?: number
  us?: number
  edition_catalog?: ProductEditionCatalog
}

export interface Product {
  id: string
  title: string
  title_key?: string
  description: string | null
  platform: 'ps4' | 'ps5' | 'xbox' | 'pc'
  type: 'game' | 'currency' | 'subscription'
  game_section?: '' | 'game' | 'new' | 'preorder'
  release_date?: string | null
  price: number
  discount_percent: number
  variants: ProductVariant[]
  image_url: string | null
  delivery_methods: ('key' | 'activation')[]
  prices?: ProductPrices
  status: 'active' | 'inactive'
  created_at: string
  updated_at: string
  platform_variants?: ProductPlatformVariant[]
}

export interface ProductPlatformVariant {
  id: string
  platform: Product['platform']
  price: number
  discount_percent: number
  image_url: string | null
}

export const platformLabels: Record<string, string> = {
  ps4: 'PlayStation 4',
  ps5: 'PlayStation 5',
  xbox: 'Xbox',
  pc: 'PC',
}

export const platformShortLabels: Record<string, string> = {
  ps4: 'PS4',
  ps5: 'PS5',
  xbox: 'Xbox',
  pc: 'PC',
}

export const typeLabels: Record<string, string> = {
  game: 'Игра',
  currency: 'Валюта',
  subscription: 'Подписка',
}

export const platformColors: Record<string, string> = {
  ps4: 'bg-blue-500/20 text-blue-300',
  ps5: 'bg-indigo-500/20 text-indigo-300',
  xbox: 'bg-green-500/20 text-green-300',
  pc: 'bg-cyan-500/20 text-cyan-300',
}
