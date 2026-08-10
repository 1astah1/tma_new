import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getMyOrders, getOrder, createOrder, createBatchOrder, sendCredentials, send2FACode } from '../services/order.service'

export function useMyOrders(status?: string, page = 1) {
  return useQuery({
    queryKey: ['myOrders', status, page],
    queryFn: () => getMyOrders(status, page),
  })
}

export function useOrder(id: string) {
  return useQuery({
    queryKey: ['order', id],
    queryFn: () => getOrder(id),
    enabled: !!id,
    refetchInterval: 5000,
  })
}

export function useCreateOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ productId, variantId, quantity = 1 }: {
      productId: string
      variantId?: string
      quantity?: number
    }) =>
      createOrder(productId, variantId, quantity),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['myOrders'] }),
  })
}

export function useCreateBatchOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ items, promoCode }: {
      items: { product_id: string; delivery_method: 'activation'; variant_id?: string; quantity: number }[]
      promoCode?: string
    }) => createBatchOrder(items, promoCode),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['myOrders'] }),
  })
}

export function useSendCredentials() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ orderId, platform, login, password }: { orderId: string; platform: string; login: string; password: string }) =>
      sendCredentials(orderId, platform, login, password),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['order'] }),
  })
}

export function useSend2FACode() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ orderId, code }: { orderId: string; code: string }) =>
      send2FACode(orderId, code),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['order'] }),
  })
}
