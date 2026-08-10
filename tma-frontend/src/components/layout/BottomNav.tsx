import { useNavigate, useLocation } from 'react-router-dom'
import { useCart } from '../../stores/cartStore'

const items = [
  { path: '/', label: 'Главная', icon: 'home', color: '#c9a84c' },
  { path: '/catalog', label: 'Каталог', icon: 'catalog', color: '#c9a84c' },
  { path: '/cart', label: 'Корзина', icon: 'cart', color: '#c9a84c' },
  { path: '/profile', label: 'Профиль', icon: 'profile', color: '#c9a84c' },
]

const HomeIcon = ({ active, color }: { active: boolean; color: string }) => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" style={{ filter: active ? `drop-shadow(0 0 6px ${color})` : 'none' }}>
    <path d="M3 10.5L12 3l9 7.5V20a1 1 0 01-1 1h-5v-6H9v6H4a1 1 0 01-1-1v-9.5z" stroke={active ? color : '#64748b'} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

const CatalogIcon = ({ active, color }: { active: boolean; color: string }) => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" style={{ filter: active ? `drop-shadow(0 0 6px ${color})` : 'none' }}>
    <path d="M4 6h7v7H4V6zm9 0h7v7h-7V6zM4 15h7v3H4v-3zm9 0h7v3h-7v-3z" stroke={active ? color : '#64748b'} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

const CartIcon = ({ active, color }: { active: boolean; color: string }) => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" style={{ filter: active ? `drop-shadow(0 0 6px ${color})` : 'none' }}>
    <path d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 100 4 2 2 0 000-4z" stroke={active ? color : '#64748b'} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

const ProfileIcon = ({ active, color }: { active: boolean; color: string }) => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" style={{ filter: active ? `drop-shadow(0 0 6px ${color})` : 'none' }}>
    <path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2M12 11a4 4 0 100-8 4 4 0 000 8z" stroke={active ? color : '#64748b'} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

const getIcon = (icon: string, active: boolean, color: string) => {
  switch (icon) {
    case 'home': return <HomeIcon active={active} color={color} />
    case 'catalog': return <CatalogIcon active={active} color={color} />
    case 'cart': return <CartIcon active={active} color={color} />
    case 'profile': return <ProfileIcon active={active} color={color} />
    default: return null
  }
}

function isNavActive(path: string, pathname: string): boolean {
  if (path === '/') return pathname === '/'
  if (path === '/catalog') return pathname === '/catalog'
  if (path === '/cart') return pathname === '/cart'
  if (path === '/profile') {
    return ['/profile', '/orders', '/wishlist', '/support', '/rules'].includes(pathname)
  }
  return pathname === path
}

export function BottomNav() {
  const navigate = useNavigate()
  const location = useLocation()
  const itemCount = useCart((s) => s.getItemCount())

  return (
    <div className="fixed bottom-[max(0.5rem,var(--app-safe-bottom))] left-2 right-2 z-50">
      <div className="mx-auto max-w-lg bg-[#161616]/95 backdrop-blur-xl rounded-xl border border-white/10 shadow-lg">
        <div className="flex items-center justify-around py-1.5 px-1">
          {items.map((item) => {
            const active = isNavActive(item.path, location.pathname)
            const isCart = item.path === '/cart'
            return (
              <button
                key={item.path}
                onClick={() => navigate(item.path)}
                className="flex flex-col items-center gap-0.5 py-1 px-2.5 rounded-lg transition-all duration-200 relative min-w-0 flex-1"
                style={{
                  background: active ? `radial-gradient(ellipse at center, ${item.color}10 0%, transparent 70%)` : 'transparent',
                }}
              >
                <div className="relative">
                  {getIcon(item.icon, active, item.color)}
                  {isCart && itemCount > 0 && (
                    <span className="absolute -top-1 -right-1 bg-red-500 text-white text-[9px] font-bold w-3.5 h-3.5 rounded-full flex items-center justify-center">
                      {itemCount > 9 ? '9+' : itemCount}
                    </span>
                  )}
                </div>
                <span
                  className="text-[10px] font-medium transition-colors duration-200 truncate max-w-full"
                  style={{ color: active ? item.color : '#64748b' }}
                >
                  {item.label}
                </span>
                {active && (
                  <div
                    className="absolute bottom-0 w-4 h-0.5 rounded-full"
                    style={{
                      backgroundColor: item.color,
                      boxShadow: `0 0 6px ${item.color}`,
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
