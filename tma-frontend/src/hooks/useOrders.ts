import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getMyOrders, getOrder, createOrder, createBatchOrder, confirmPayment, confirmBatchPayment, sendCredentials, send2FACode } from '../services/order.service'

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
  })
}

export function useCreateOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ productId, deliveryMethod, variantId, quantity = 1 }: {
      productId: string
      deliveryMethod: 'key' | 'activation'
      variantId?: string
      quantity?: number
    }) =>
      createOrder(productId, deliveryMethod, variantId, quantity),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['myOrders'] }),
  })
}

export function useCreateBatchOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (items: { product_id: string; delivery_method: 'key' | 'activation'; variant_id?: string; quantity: number }[]) =>
      createBatchOrder(items),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['myOrders'] }),
  })
}

export function useConfirmPayment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ orderId, paymentMethod, file }: { orderId: string; paymentMethod: string; file: File }) =>
      confirmPayment(orderId, paymentMethod, file),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['order'] }),
  })
}

export function useConfirmBatchPayment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ orderIds, paymentMethod, file }: { orderIds: string[]; paymentMethod: string; file: File }) =>
      confirmBatchPayment(orderIds, paymentMethod, file),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['myOrders'] })
      qc.invalidateQueries({ queryKey: ['order'] })
    },
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
