import { useGoBack } from '../hooks/useGoBack'
import { Header } from '../components/layout/Header'
import { Loader } from '../components/ui/Button'
import { useShopSettings } from '../hooks/useContent'

const defaultRules = [
  'Оформление заказа происходит через Telegram-чат с менеджером.',
  'Оплата производится по реквизитам, указанным менеджером.',
  'Возврат возможен до оформления покупки на аккаунте. Напишите менеджеру в Telegram.',
  'Данные аккаунта используются только для активации и шифруются.',
  'Срок обработки заказа зависит от типа товара и загрузки менеджеров.',
  'По спорным вопросам обращайтесь в поддержку.',
]

export function RulesPage() {
  const goBack = useGoBack('/profile')
  const { data: settings, isLoading } = useShopSettings()

  const rules = settings?.shop_rules
    ? settings.shop_rules.split('\n').map((line) => line.trim()).filter(Boolean)
    : defaultRules

  return (
    <div className="pb-page">
      <Header title="Правила магазина" onBack={goBack} />
      <div className="p-4 space-y-3">
        {isLoading ? (
          <div className="flex justify-center py-8"><Loader /></div>
        ) : (
          rules.map((rule, index) => (
            <div key={index} className="bg-[var(--tg-secondary)] rounded-xl p-4 text-sm text-[var(--tg-text)]">
              {rule}
            </div>
          ))
        )}
      </div>
    </div>
  )
}
