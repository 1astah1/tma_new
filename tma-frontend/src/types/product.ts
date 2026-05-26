export interface ProductVariant {
  id: string
  name: string
  price: number
  stock: number
}

export interface Product {
  id: string
  title: string
  description: string | null
  platform: 'ps4' | 'ps5' | 'xbox'
  type: 'game' | 'currency' | 'subscription'
  price: number
  discount_percent: number
  variants: ProductVariant[]
  image_url: string | null
  delivery_methods: ('key' | 'activation')[]
  status: 'active' | 'inactive'
  created_at: string
  updated_at: string
}

export const platformLabels: Record<string, string> = {
  ps4: 'PlayStation 4',
  ps5: 'PlayStation 5',
  xbox: 'Xbox',
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
}
