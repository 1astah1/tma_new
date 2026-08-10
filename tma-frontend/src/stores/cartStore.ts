import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export interface CartItem {
  productId: string
  variantId?: string
  quantity: number
  title: string
  price: number
  image?: string
  deliveryMethod: 'activation'
}

interface CartStore {
  items: CartItem[]
  promoCode: string | null
  promoDiscount: number
  addItem: (item: CartItem) => void
  removeItem: (productId: string, variantId?: string) => void
  updateQuantity: (productId: string, variantId: string | undefined, quantity: number) => void
  clearCart: () => void
  setPromoCode: (code: string | null) => void
  setPromoDiscount: (discount: number) => void
  getTotal: () => number
  getSubtotal: () => number
  getItemCount: () => number
}

export const useCart = create<CartStore>()(
  persist(
    (set, get) => ({
      items: [],
      promoCode: null,
      promoDiscount: 0,
      addItem: (item) =>
        set((state) => {
          const existing = state.items.find(
            (i) => i.productId === item.productId && i.variantId === item.variantId
          )
          if (existing) {
            return {
              items: state.items.map((i) =>
                i.productId === item.productId && i.variantId === item.variantId
                  ? { ...i, quantity: i.quantity + item.quantity }
                  : i
              ),
            }
          }
          return { items: [...state.items, item] }
        }),
      removeItem: (productId, variantId) =>
        set((state) => ({
          items: state.items.filter(
            (i) => !(i.productId === productId && i.variantId === variantId)
          ),
        })),
      updateQuantity: (productId, variantId, quantity) =>
        set((state) => ({
          items:
            quantity <= 0
              ? state.items.filter(
                  (i) => !(i.productId === productId && i.variantId === variantId)
                )
              : state.items.map((i) =>
                  i.productId === productId && i.variantId === variantId
                    ? { ...i, quantity }
                    : i
                ),
        })),
      clearCart: () => set({ items: [], promoCode: null, promoDiscount: 0 }),
      setPromoCode: (code) => set({ promoCode: code }),
      setPromoDiscount: (discount) => set({ promoDiscount: discount }),
      getSubtotal: () => get().items.reduce((sum, i) => sum + i.price * i.quantity, 0),
      getTotal: () => get().getSubtotal() - get().promoDiscount,
      getItemCount: () => get().items.reduce((sum, i) => sum + i.quantity, 0),
    }),
    { name: 'cart-storage' }
  )
)
