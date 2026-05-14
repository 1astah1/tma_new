import { useQuery } from '@tanstack/react-query'
import api from '../services/api'

export function useProfile() {
  return useQuery({
    queryKey: ['profile'],
    queryFn: async () => {
      const { data } = await api.get('/profile')
      return data
    },
  })
}
