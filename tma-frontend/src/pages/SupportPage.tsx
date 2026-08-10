import { useGoBack } from '../hooks/useGoBack'
import { Header } from '../components/layout/Header'
import { Button } from '../components/ui/Button'
import { Loader } from '../components/ui/Button'
import { resolveFaqItems } from '../constants/defaultFaq'
import { useFaq, useShopSettings } from '../hooks/useContent'

export function SupportPage() {
  const goBack = useGoBack('/profile')
  const { data: settings } = useShopSettings()
  const { data: faq, isLoading } = useFaq()
  const faqItems = resolveFaqItems(faq)

  return (
    <div className="pb-page">
      <Header title="Поддержка" onBack={goBack} />
      <div className="p-4 space-y-4">
        <div className="bg-[var(--tg-secondary)] rounded-xl p-6 text-center">
          <div className="text-4xl mb-4">💬</div>
          <h2 className="text-lg font-bold mb-2">Связаться с поддержкой</h2>
          <p className="text-sm text-[var(--tg-hint)] mb-6">
            Напишите нам в Telegram — ответим в течение 15 минут
          </p>
          <Button
            fullWidth
            size="lg"
            onClick={() => window.open(settings?.support_url || 'https://t.me/coin_mint_chat', '_blank')}
          >
            ✈️ Написать в Telegram
          </Button>
        </div>

        <div className="bg-[var(--tg-secondary)] rounded-xl p-4">
          <h3 className="font-medium mb-3">❓ Частые вопросы</h3>
          {isLoading ? (
            <div className="flex justify-center py-6"><Loader /></div>
          ) : (
            <div className="space-y-3 text-sm">
              {faqItems.map((item) => (
                <details key={item.id} className="group">
                  <summary className="cursor-pointer font-medium list-none flex justify-between">
                    <span>{item.question}</span>
                    <span className="text-[var(--tg-hint)] group-open:rotate-180 transition-transform">▼</span>
                  </summary>
                  <p className="text-[var(--tg-hint)] mt-2">{item.answer}</p>
                </details>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
