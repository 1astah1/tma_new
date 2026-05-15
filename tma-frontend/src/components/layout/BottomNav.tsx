import { useNavigate, useLocation } from 'react-router-dom'

const items = [
  { path: '/', label: 'Магазин', icon: 'cart', color: '#38bdf8' },
  { path: '/orders', label: 'Заказы', icon: 'orders', color: '#a78bfa' },
  { path: '/profile', label: 'Профиль', icon: 'profile', color: '#fb923c' },
]

const CartIcon = ({ active, color }: { active: boolean; color: string }) => (
  <svg width="26" height="26" viewBox="0 0 24 24" fill="none" style={{ filter: active ? `drop-shadow(0 0 8px ${color})` : 'none' }}>
    <path d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 100 4 2 2 0 000-4z" stroke={active ? color : '#64748b'} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

const OrdersIcon = ({ active, color }: { active: boolean; color: string }) => (
  <svg width="26" height="26" viewBox="0 0 24 24" fill="none" style={{ filter: active ? `drop-shadow(0 0 8px ${color})` : 'none' }}>
    <path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" stroke={active ? color : '#64748b'} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

const ProfileIcon = ({ active, color }: { active: boolean; color: string }) => (
  <svg width="26" height="26" viewBox="0 0 24 24" fill="none" style={{ filter: active ? `drop-shadow(0 0 8px ${color})` : 'none' }}>
    <path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2M12 11a4 4 0 100-8 4 4 0 000 8z" stroke={active ? color : '#64748b'} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

const getIcon = (icon: string, active: boolean, color: string) => {
  switch (icon) {
    case 'cart': return <CartIcon active={active} color={color} />
    case 'orders': return <OrdersIcon active={active} color={color} />
    case 'profile': return <ProfileIcon active={active} color={color} />
    default: return null
  }
}

export function BottomNav() {
  const navigate = useNavigate()
  const location = useLocation()

  return (
    <div className="fixed bottom-3 left-3 right-3 z-50">
      <div className="max-w-md mx-auto bg-[#1a1035]/90 backdrop-blur-2xl rounded-2xl border border-white/10 shadow-[0_8px_32px_rgba(0,0,0,0.4)] overflow-hidden">
        <div className="flex items-center justify-around py-2 px-2">
          {items.map((item) => {
            const active = location.pathname === item.path
            return (
              <button
                key={item.path}
                onClick={() => navigate(item.path)}
                className="flex flex-col items-center gap-1 py-2 px-4 rounded-xl transition-all duration-300 relative"
                style={{
                  background: active ? `radial-gradient(ellipse at center, ${item.color}15 0%, transparent 70%)` : 'transparent',
                }}
              >
                {getIcon(item.icon, active, item.color)}
                <span
                  className="text-[11px] font-medium transition-colors duration-300"
                  style={{ color: active ? item.color : '#64748b' }}
                >
                  {item.label}
                </span>
                {active && (
                  <div
                    className="absolute bottom-0.5 w-6 h-0.5 rounded-full"
                    style={{
                      backgroundColor: item.color,
                      boxShadow: `0 0 10px ${item.color}, 0 0 20px ${item.color}40`,
                    }}
                  />
                )}
              </button>
            )
          })}
        </div>
      </div>
    </div>
  )
}
