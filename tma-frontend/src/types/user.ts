export interface User {
  id: string
  telegram_id: number
  username: string | null
  first_name: string | null
  last_interaction: string | null
  created_at: string
  updated_at: string
}
