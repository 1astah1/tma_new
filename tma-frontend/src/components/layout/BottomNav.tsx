import { useNavigate, useLocation } from 'react-router-dom'

const items = [
  { path: '/', label: 'Магазин', icon: '🏪' },
  { path: '/orders', label: 'Заказы', icon: '📋' },
  { path: '/profile', label: 'Профиль', icon: '👤' },
]

export function BottomNav() {
  const navigate = useNavigate()
  const location = useLocation()

  return (
    <div className="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 z-10">
      <div className="max-w-lg mx-auto flex">
        {items.map((item) => {
          const active = location.pathname === item.path
          return (
            <button
              key={item.path}
              onClick={() => navigate(item.path)}
              className={`flex-1 py-2 flex flex-col items-center text-xs ${
                active ? 'text-tg-button font-medium' : 'text-gray-500'
              }`}
            >
              <span className="text-lg">{item.icon}</span>
              <span>{item.label}</span>
            </button>
          )
        })}
      </div>
    </div>
  )
}
