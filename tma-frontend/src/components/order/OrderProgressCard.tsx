import {
  getOrderProgressPercent,
  getOrderUiPhase,
  orderUiPhaseDescriptions,
  orderUiPhaseLabels,
  ORDER_PROGRESS_STEPS,
  OrderUiPhase,
} from '../../utils/orderStatus'

function phaseIndex(phase: OrderUiPhase): number {
  const order: OrderUiPhase[] = ['created', 'manager', 'processing', 'done']
  const idx = order.indexOf(phase)
  return idx >= 0 ? idx : 0
}

export function OrderProgressCard({ status }: { status: string }) {
  const phase = getOrderUiPhase(status as any)
  const isCancelled = phase === 'cancelled' || phase === 'refund'
  const isDone = phase === 'done'
  const progress = getOrderProgressPercent(status as any)
  const currentStep = phaseIndex(isDone ? 'done' : phase)

  if (isCancelled) {
    return (
      <div className="rounded-xl bg-[var(--tg-secondary)] p-4">
        <div className="mb-2 flex items-center justify-between">
          <span className="text-xs text-[var(--tg-hint)]">Статус заказа</span>
          <span className="text-xs text-red-400">{orderUiPhaseLabels[phase]}</span>
        </div>
        <p className="text-sm text-[var(--tg-hint)] leading-relaxed">{orderUiPhaseDescriptions[phase]}</p>
      </div>
    )
  }

  return (
    <div className="rounded-xl bg-[var(--tg-secondary)] p-4">
      <div className="mb-3 flex items-center justify-between">
        <span className="text-xs text-[var(--tg-hint)]">Статус заказа</span>
        <span className={`text-xs font-semibold ${isDone ? 'text-green-400' : 'text-[var(--tg-text)]'}`}>
          {orderUiPhaseLabels[phase]}
        </span>
      </div>

      <div className="mb-3 h-2 overflow-hidden rounded-full bg-[var(--tg-card)]">
        <div
          className={`h-full rounded-full transition-all duration-700 ${
            isDone ? 'bg-gradient-to-r from-green-500 to-green-400' : 'bg-[var(--tg-button)]'
          }`}
          style={{ width: `${progress}%` }}
        />
      </div>

      <div className="mb-3 grid grid-cols-4 gap-1">
        {ORDER_PROGRESS_STEPS.map((step, index) => {
          const active = index <= currentStep
          return (
            <div key={step.id} className="text-center">
              <div
                className={`mx-auto mb-1 h-2 w-2 rounded-full ${
                  active ? (isDone ? 'bg-green-400' : 'bg-[var(--tg-button)]') : 'bg-white/15'
                }`}
              />
              <div className={`text-[10px] leading-tight ${active ? 'text-[var(--tg-text)]' : 'text-[var(--tg-hint)]'}`}>
                {step.label}
              </div>
            </div>
          )
        })}
      </div>

      <p className="border-t border-[var(--tg-border)] pt-3 text-xs leading-relaxed text-[var(--tg-hint)]">
        {orderUiPhaseDescriptions[phase]}
      </p>
    </div>
  )
}
