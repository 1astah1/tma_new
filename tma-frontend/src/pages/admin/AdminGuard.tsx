import { ReactNode, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Header } from '../../components/layout/Header'
import { useProfile } from '../../hooks/useProfile'

/**
 * Пускает в управление только тех, кто есть в списке админов. Сервер и так
 * отвечает 403, но и пустой каркас экрана покупателю показывать незачем.
 */
export function AdminGuard({ title, onBack, children }: { title: string; onBack?: () => void; children: ReactNode }) {
  const nav = useNavigate()
  const { data: profile, isLoading } = useProfile()
  const isAdmin = !!profile?.is_admin

  useEffect(() => {
    if (!isLoading && profile && !isAdmin) nav('/', { replace: true })
  }, [isLoading, profile, isAdmin, nav])

  if (!isAdmin) {
    return (
      <div className="pb-page">
        <Header title="" onBack={onBack} />
        <div className="p-10 text-center text-[var(--tg-hint)]">
          {isLoading ? 'Загрузка…' : 'Раздел доступен только менеджерам'}
        </div>
      </div>
    )
  }

  return (
    <div className="pb-page">
      <Header title={title} onBack={onBack} />
      {children}
    </div>
  )
}
