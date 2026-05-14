import { HashRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useTelegram } from './hooks/useTelegram'
import { useAuth } from './hooks/useAuth'
import { BottomNav } from './components/layout/BottomNav'
import { HomePage } from './pages/HomePage'
import { CatalogPage } from './pages/CatalogPage'
import { ProductPage } from './pages/ProductPage'
import { CheckoutPage } from './pages/CheckoutPage'
import { OrderStatusPage } from './pages/OrderStatusPage'
import { OrdersHistoryPage } from './pages/OrdersHistoryPage'
import { ProfilePage } from './pages/ProfilePage'
import { Loader } from './components/ui/Button'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30000 },
  },
})

function AppContent() {
  useTelegram()
  const { isAuthenticated, loading } = useAuth()

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen bg-tg-bg">
        <div className="text-center">
          <Loader />
          <p className="text-tg-hint text-sm mt-4">Загрузка...</p>
        </div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return (
      <div className="flex items-center justify-center h-screen bg-tg-bg">
        <div className="text-center">
          <p className="text-tg-hint text-sm">Ошибка авторизации</p>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-lg mx-auto min-h-screen bg-tg-bg">
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/catalog" element={<CatalogPage />} />
        <Route path="/product/:id" element={<ProductPage />} />
        <Route path="/checkout/:id" element={<CheckoutPage />} />
        <Route path="/order/:id" element={<OrderStatusPage />} />
        <Route path="/orders" element={<OrdersHistoryPage />} />
        <Route path="/profile" element={<ProfilePage />} />
      </Routes>
      <BottomNav />
    </div>
  )
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <HashRouter>
        <AppContent />
      </HashRouter>
    </QueryClientProvider>
  )
}
