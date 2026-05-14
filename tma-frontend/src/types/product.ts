export interface Product {
  id: string
  title: string
  description: string | null
  platform: 'ps4' | 'ps5' | 'xbox'
  type: 'game' | 'currency' | 'subscription'
  price: number
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
  ps4: 'bg-blue-100 text-blue-800',
  ps5: 'bg-indigo-100 text-indigo-800',
  xbox: 'bg-green-100 text-green-800',
}
