import { useNavigate } from 'react-router-dom'
import { Header } from '../components/layout/Header'

export function RulesPage() {
  const nav = useNavigate()

  const rules = [
    { icon: '🔒', title: 'Конфиденциальность', text: 'Все данные пользователей шифруются и не передаются третьим лицам.' },
    { icon: '💰', title: 'Оплата', text: 'Оплата производится через СБП, банковскую карту или криптовалюту. После оплаты прикрепите чек.' },
    { icon: '📦', title: 'Выдача товара', text: 'Ключи выдаются автоматически. Активация выполняется администратором в течение 30 минут.' },
    { icon: '🔄', title: 'Возврат', text: 'Возврат возможен до выдачи ключа или начала активации. Обратитесь в поддержку.' },
    { icon: '⚠️', title: 'Ответственность', text: 'Мы не несём ответственности за блокировку аккаунтов, связанных с нарушением правил платформ.' },
    { icon: '🛡️', title: 'Гарантия', text: 'Гарантия на товар — 24 часа с момента выдачи. При проблемах обратитесь в поддержку.' },
  ]

  return (
    <div className="pb-24">
      <Header title="Правила магазина" onBack={() => nav(-1)} />
      <div className="p-4 space-y-3">
        {rules.map((rule, i) => (
          <div key={i} className="bg-[var(--tg-secondary)] rounded-xl p-4">
            <div className="flex items-start gap-3">
              <span className="text-2xl">{rule.icon}</span>
              <div>
                <h3 className="font-medium mb-1">{rule.title}</h3>
                <p className="text-sm text-[var(--tg-hint)]">{rule.text}</p>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
