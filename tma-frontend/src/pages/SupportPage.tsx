import { useNavigate } from 'react-router-dom'
import { Header } from '../components/layout/Header'
import { Button } from '../components/ui/Button'

export function SupportPage() {
  const nav = useNavigate()

  return (
    <div className="pb-24">
      <Header title="Поддержка" onBack={() => nav(-1)} />
      <div className="p-4 space-y-4">
        <div className="bg-[var(--tg-secondary)] rounded-xl p-6 text-center">
          <div className="text-4xl mb-4">💬</div>
          <h2 className="text-lg font-bold mb-2">Связаться с поддержкой</h2>
          <p className="text-sm text-[var(--tg-hint)] mb-6">
            Напишите нам в Telegram — ответим в течение 15 минут
          </p>
          <Button fullWidth size="lg" onClick={() => window.open('https://t.me/your_support_bot', '_blank')}>
            ✈️ Написать в Telegram
          </Button>
        </div>

        <div className="bg-[var(--tg-secondary)] rounded-xl p-4">
          <h3 className="font-medium mb-3">❓ Частые вопросы</h3>
          <div className="space-y-3 text-sm">
            <details className="group">
              <summary className="cursor-pointer font-medium list-none flex justify-between">
                <span>Как получить товар после оплаты?</span>
                <span className="text-[var(--tg-hint)] group-open:rotate-180 transition-transform">▼</span>
              </summary>
              <p className="text-[var(--tg-hint)] mt-2">
                После подтверждения оплаты товар выдаётся автоматически. Для ключей — сразу в заказе, для активации — администратор войдёт в ваш аккаунт.
              </p>
            </details>
            <details className="group">
              <summary className="cursor-pointer font-medium list-none flex justify-between">
                <span>Можно ли вернуть товар?</span>
                <span className="text-[var(--tg-hint)] group-open:rotate-180 transition-transform">▼</span>
              </summary>
              <p className="text-[var(--tg-hint)] mt-2">
                Возврат возможен до момента выдачи ключа или начала активации. Обратитесь в поддержку.
              </p>
            </details>
            <details className="group">
              <summary className="cursor-pointer font-medium list-none flex justify-between">
                <span>Безопасно ли передавать данные аккаунта?</span>
                <span className="text-[var(--tg-hint)] group-open:rotate-180 transition-transform">▼</span>
              </summary>
              <p className="text-[var(--tg-hint)] mt-2">
                Все данные шифруются и используются только для активации товара. После завершения данные удаляются.
              </p>
            </details>
          </div>
        </div>
      </div>
    </div>
  )
}
