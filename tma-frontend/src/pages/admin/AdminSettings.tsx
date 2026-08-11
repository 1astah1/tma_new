import { useEffect, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { AdminGuard } from './AdminGuard'
import { Button } from '../../components/ui/Button'
import { useToast } from '../../components/ui/Toast'
import { useGoBack } from '../../hooks/useGoBack'
import { useProfile } from '../../hooks/useProfile'
import { getSetting, saveSetting } from '../../services/tmaAdmin.service'

type PaymentDetails = {
  sbp?: { phone?: string; bank?: string; receiver?: string }
  card?: { number?: string; bank?: string }
  crypto?: { binance?: string; bybit?: string; trc20?: string }
}

type Pricing = {
  turkey_markup?: number
  ukraine_markup?: number
  xbox_usd_multiplier?: number
  min_price_rub?: number
}

function Field({
  label,
  value,
  onChange,
  hint,
  numeric,
}: {
  label: string
  value: string
  onChange: (v: string) => void
  hint?: string
  numeric?: boolean
}) {
  return (
    <label className="block">
      <span className="mb-1 block text-xs text-white/50">{label}</span>
      <input
        value={value}
        inputMode={numeric ? 'decimal' : undefined}
        onChange={(e) => onChange(e.target.value)}
        className="w-full rounded-xl border border-white/10 bg-[#141414] px-3 py-2 text-sm outline-none focus:border-amber-500/50"
      />
      {hint ? <span className="mt-1 block text-[11px] text-white/35">{hint}</span> : null}
    </label>
  )
}

export function AdminSettings() {
  const goBack = useGoBack()
  const toast = useToast()
  const { data: profile } = useProfile()
  const isAdmin = !!profile?.is_admin

  const [payment, setPayment] = useState<PaymentDetails>({})
  const [pricing, setPricing] = useState<Pricing>({})
  const [manager, setManager] = useState('')
  const [support, setSupport] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!isAdmin) return
    Promise.all([
      getSetting('payment_details'),
      getSetting('pricing_formulas'),
      getSetting('manager_url'),
      getSetting('support_url'),
    ])
      .then(([pay, price, mgr, sup]) => {
        const parse = (v: unknown) => (typeof v === 'string' ? JSON.parse(v || '{}') : v || {})
        setPayment(parse(pay))
        setPricing(parse(price))
        setManager(typeof mgr === 'string' ? mgr : '')
        setSupport(typeof sup === 'string' ? sup : '')
      })
      .catch(() => toast.toast('Не удалось загрузить настройки', 'error'))
      .finally(() => setLoading(false))
  }, [isAdmin])

  const save = useMutation({
    mutationFn: async () => {
      await saveSetting('payment_details', payment)
      await saveSetting('pricing_formulas', {
        turkey_markup: Number(pricing.turkey_markup) || 2.2,
        ukraine_markup: Number(pricing.ukraine_markup) || 2.3,
        xbox_usd_multiplier: Number(pricing.xbox_usd_multiplier) || 80,
        min_price_rub: Number(pricing.min_price_rub) || 149,
      })
      if (manager) await saveSetting('manager_url', manager)
      if (support) await saveSetting('support_url', support)
    },
    onSuccess: () => toast.toast('Сохранено. Цены пересчитаются при следующем обновлении каталога', 'success'),
    onError: () => toast.toast('Не сохранилось', 'error'),
  })

  return (
    <AdminGuard title="Настройки" onBack={goBack}>
      <div className="space-y-5 p-4">
        {loading ? (
          <div className="py-10 text-center text-white/50">Загрузка…</div>
        ) : (
          <>
            <section className="space-y-2">
              <h2 className="text-sm font-bold text-amber-200">Реквизиты для оплаты</h2>
              <Field
                label="СБП — телефон"
                value={payment.sbp?.phone ?? ''}
                onChange={(v) => setPayment({ ...payment, sbp: { ...payment.sbp, phone: v } })}
              />
              <Field
                label="СБП — банк"
                value={payment.sbp?.bank ?? ''}
                onChange={(v) => setPayment({ ...payment, sbp: { ...payment.sbp, bank: v } })}
              />
              <Field
                label="СБП — получатель"
                value={payment.sbp?.receiver ?? ''}
                onChange={(v) => setPayment({ ...payment, sbp: { ...payment.sbp, receiver: v } })}
              />
              <Field
                label="Карта — номер"
                value={payment.card?.number ?? ''}
                onChange={(v) => setPayment({ ...payment, card: { ...payment.card, number: v } })}
              />
              <Field
                label="Карта — банк"
                value={payment.card?.bank ?? ''}
                onChange={(v) => setPayment({ ...payment, card: { ...payment.card, bank: v } })}
              />
            </section>

            <section className="space-y-2">
              <h2 className="text-sm font-bold text-amber-200">Наценки</h2>
              <Field
                numeric
                label="PlayStation Турция"
                hint="номинал карты в лирах × коэффициент"
                value={String(pricing.turkey_markup ?? 2.2)}
                onChange={(v) => setPricing({ ...pricing, turkey_markup: Number(v) })}
              />
              <Field
                numeric
                label="PlayStation Украина"
                hint="гривны × коэффициент"
                value={String(pricing.ukraine_markup ?? 2.3)}
                onChange={(v) => setPricing({ ...pricing, ukraine_markup: Number(v) })}
              />
              <Field
                numeric
                label="Xbox США"
                hint="доллары × коэффициент"
                value={String(pricing.xbox_usd_multiplier ?? 80)}
                onChange={(v) => setPricing({ ...pricing, xbox_usd_multiplier: Number(v) })}
              />
              <Field
                numeric
                label="Минимальная цена, ₽"
                value={String(pricing.min_price_rub ?? 149)}
                onChange={(v) => setPricing({ ...pricing, min_price_rub: Number(v) })}
              />
            </section>

            <section className="space-y-2">
              <h2 className="text-sm font-bold text-amber-200">Связь</h2>
              <Field label="Личка менеджера" hint="сюда уходят заявки на покупку" value={manager} onChange={setManager} />
              <Field label="Чат поддержки" value={support} onChange={setSupport} />
            </section>

            <Button size="md" className="w-full" onClick={() => save.mutate()} loading={save.isPending}>
              Сохранить
            </Button>
          </>
        )}
      </div>
    </AdminGuard>
  )
}
