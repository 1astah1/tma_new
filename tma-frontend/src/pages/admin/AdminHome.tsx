import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { AdminGuard } from './AdminGuard'
import { useGoBack } from '../../hooks/useGoBack'
import { formatPrice } from '../../utils/format'
import { getAdminStats } from '../../services/tmaAdmin.service'

const SECTIONS = [
  { path: '/manage/orders', icon: '📦', title: 'Заказы', hint: 'Статусы и переписка' },
  { path: '/manage/catalog', icon: '🎮', title: 'Каталог', hint: 'Импорт и обновление цен' },
  { path: '/manage/products', icon: '🏷', title: 'Товары', hint: 'Цены и видимость' },
  { path: '/manage/settings', icon: '⚙️', title: 'Настройки', hint: 'Реквизиты и наценки' },
  { path: '/manage/users', icon: '👥', title: 'Покупатели', hint: 'Блокировки' },
  { path: '/manage/promos', icon: '🎟', title: 'Промокоды', hint: 'Скидки' },
  { path: '/manage/staff', icon: '🧑‍💼', title: 'Команда', hint: 'Админы и менеджеры' },
]

export function AdminHome() {
  const nav = useNavigate()
  const goBack = useGoBack()
  const { data: stats } = useQuery({ queryKey: ['tma-admin-stats'], queryFn: getAdminStats })

  const tiles = [
    { label: 'Заказов сегодня', value: String(stats?.orders_today ?? 0) },
    { label: 'Ждут оплаты', value: String(stats?.waiting_payment ?? 0) },
    { label: 'В работе, ₽', value: formatPrice(stats?.pending_revenue ?? 0) },
    { label: 'Всего заказов', value: String(stats?.total_orders ?? 0) },
    { label: 'Покупателей', value: String(stats?.total_users ?? 0) },
    { label: 'Товаров', value: String(stats?.catalog?.active_game_products ?? 0) },
  ]

  return (
    <AdminGuard title="Управление" onBack={goBack}>
      <div className="space-y-4 p-4">
        <div className="grid grid-cols-3 gap-2">
          {tiles.map((tile) => (
            <div key={tile.label} className="rounded-2xl border border-white/10 bg-[#141414] p-3 text-center">
              <div className="text-base font-black text-amber-200">{tile.value}</div>
              <div className="text-[11px] leading-tight text-white/50">{tile.label}</div>
            </div>
          ))}
        </div>

        <div className="grid grid-cols-2 gap-2">
          {SECTIONS.map((section) => (
            <button
              key={section.path}
              onClick={() => nav(section.path)}
              className="rounded-2xl border border-white/10 bg-[#141414] p-4 text-left transition-colors active:bg-white/5"
            >
              <div className="mb-1 text-2xl">{section.icon}</div>
              <div className="font-semibold">{section.title}</div>
              <div className="text-[11px] text-white/45">{section.hint}</div>
            </button>
          ))}
        </div>
      </div>
    </AdminGuard>
  )
}
