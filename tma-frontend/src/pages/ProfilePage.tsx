import { useNavigate } from 'react-router-dom'
import { useProfile } from '../hooks/useProfile'
import { Header } from '../components/layout/Header'
import { Card } from '../components/ui/Card'
import { Loader } from '../components/ui/Button'
import { formatDate } from '../utils/format'

export function ProfilePage() {
  const nav = useNavigate()
  const { data: profile, isLoading } = useProfile()

  if (isLoading) return <div className="flex justify-center py-20"><Loader /></div>

  return (
    <div className="pb-24">
      <Header title="Профиль" />
      <div className="p-4 space-y-4">
        <Card>
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 bg-tg-button/20 rounded-full flex items-center justify-center text-2xl">
              👤
            </div>
            <div>
              <div className="font-semibold text-lg">{profile?.username || 'Пользователь'}</div>
              <div className="text-sm text-tg-hint">ID: {profile?.telegram_id}</div>
              <div className="text-xs text-tg-hint">
                Зарегистрирован: {profile?.created_at ? formatDate(profile.created_at) : '-'}
              </div>
            </div>
          </div>
        </Card>

        <div className="space-y-2">
          <button
            onClick={() => nav('/orders')}
            className="w-full flex items-center gap-3 p-4 bg-white rounded-xl border border-gray-100 hover:bg-gray-50"
          >
            <span className="text-xl">📋</span>
            <span className="font-medium">Мои заказы</span>
            <span className="ml-auto text-tg-hint">→</span>
          </button>
          <button className="w-full flex items-center gap-3 p-4 bg-white rounded-xl border border-gray-100 hover:bg-gray-50">
            <span className="text-xl">💬</span>
            <span className="font-medium">Поддержка</span>
            <span className="ml-auto text-tg-hint">→</span>
          </button>
          <button className="w-full flex items-center gap-3 p-4 bg-white rounded-xl border border-gray-100 hover:bg-gray-50">
            <span className="text-xl">📖</span>
            <span className="font-medium">Правила магазина</span>
            <span className="ml-auto text-tg-hint">→</span>
          </button>
        </div>
      </div>
    </div>
  )
}
