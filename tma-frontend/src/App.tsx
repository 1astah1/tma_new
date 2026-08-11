import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useTelegram } from './hooks/useTelegram'
import { useAuth } from './hooks/useAuth'
import { BottomNav } from './components/layout/BottomNav'
import { HomePage } from './pages/HomePage'
import { HomeCategoryPage } from './pages/HomeCategoryPage'
import { CatalogPage } from './pages/CatalogPage'
import { ProductPage } from './pages/ProductPage'
import { OrderStatusPage } from './pages/OrderStatusPage'
import { OrdersHistoryPage } from './pages/OrdersHistoryPage'
import { ProfilePage } from './pages/ProfilePage'
import { SupportPage } from './pages/SupportPage'
import { RulesPage } from './pages/RulesPage'
import { CartPage } from './pages/CartPage'
import { WishlistPage } from './pages/WishlistPage'
import { AdminPage } from './pages/AdminPage'
import { ErrorBoundary } from './components/ui/ErrorBoundary'
import { ToastProvider } from './components/ui/Toast'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30000 },
  },
})

function AppShell() {
  useTelegram()
  const { authError } = useAuth()

  return (
    <div
      className="max-w-lg mx-auto min-h-screen overflow-x-hidden bg-[var(--tg-bg)] text-[var(--tg-text)]"
      style={{ minHeight: 'var(--tg-viewport-height, 100vh)' }}
    >
      {authError ? (
        <div className="px-4 py-2 text-center text-xs text-amber-500 border-b border-amber-500/20">
          {authError}
        </div>
      ) : null}
      <Outlet />
      <BottomNav />
    </div>
  )
}

function AppRoutes() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<HomePage />} />
        <Route path="category/:id" element={<HomeCategoryPage />} />
        <Route path="catalog" element={<CatalogPage />} />
        <Route path="product/:id" element={<ProductPage />} />
        <Route path="order/:id" element={<OrderStatusPage />} />
        {/* Страховка от опечатки в ссылке: /orders/<id> вёл на «*» и молча
            выкидывал на главную уже после того, как заказ создан. */}
        <Route path="orders/:id" element={<OrderStatusPage />} />
        <Route path="orders" element={<OrdersHistoryPage />} />
        <Route path="profile" element={<ProfilePage />} />
        <Route path="support" element={<SupportPage />} />
        <Route path="rules" element={<RulesPage />} />
        <Route path="cart" element={<CartPage />} />
        <Route path="wishlist" element={<WishlistPage />} />
        <Route path="manage" element={<AdminPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <ToastProvider>
          <ErrorBoundary>
            <AppRoutes />
          </ErrorBoundary>
        </ToastProvider>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
