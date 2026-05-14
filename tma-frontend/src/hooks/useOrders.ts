import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getMyOrders, getOrder, createOrder, confirmPayment, sendCredentials, send2FACode } from '../services/order.service'

export function useMyOrders(status?: string) {
  return useQuery({
    queryKey: ['myOrders', status],
    queryFn: () => getMyOrders(status),
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
    mutationFn: ({ productId, deliveryMethod }: { productId: string; deliveryMethod: 'key' | 'activation' }) =>
      createOrder(productId, deliveryMethod),
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
