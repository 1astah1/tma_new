import { useQuery } from '@tanstack/react-query'
import { getFaq, getShopSettings } from '../services/content.service'

export function useShopSettings() {
  return useQuery({
    queryKey: ['shop-settings'],
    queryFn: getShopSettings,
    staleTime: 60000,
  })
}

export function useFaq() {
  return useQuery({
    queryKey: ['faq'],
    queryFn: getFaq,
    staleTime: 60000,
  })
}
