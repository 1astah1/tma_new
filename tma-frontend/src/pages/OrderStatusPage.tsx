import { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useOrder, useSendCredentials, useSend2FACode } from '../hooks/useOrders'
import { Header } from '../components/layout/Header'
import { Button, Loader } from '../components/ui/Button'
import { StatusBadge } from '../components/ui/StatusBadge'
import { statusLabels, OrderStatus } from '../types/order'
import { formatPrice } from '../utils/format'
import api from '../services/api'

function ProgressStepper({ status, deliveryMethod }: { status: string; deliveryMethod: string }) {
  const isKey = deliveryMethod === 'key'

  const keySteps = [
    { key: 'WAITING_PAYMENT', label: 'Оплата' },
    { key: 'PAYMENT_VERIFICATION', label: 'Проверка' },
    { key: 'PAID', label: 'Оплачен' },
    { key: 'KEY_ISSUED', label: 'Ключ выдан' },
    { key: 'COMPLETED', label: 'Завершён' },
  ]

  const activationSteps = [
    { key: 'WAITING_PAYMENT', label: 'Оплата' },
    { key: 'PAYMENT_VERIFICATION', label: 'Проверка' },
    { key: 'PAID', label: 'Оплачен' },
    { key: 'WAITING_ACTIVATION', label: 'В очереди' },
    { key: 'AWAITING_CREDENTIALS', label: 'Данные' },
    { key: 'CREDENTIALS_RECEIVED', label: 'Проверка' },
    { key: 'CREDENTIALS_INVALID', label: 'Данные неверны' },
    { key: 'AWAITING_2FA', label: '2FA' },
    { key: 'INVALID_2FA', label: 'Код неверен' },
    { key: 'ACTIVATING', label: 'Активация' },
    { key: 'ACTIVATED', label: 'Готово' },
    { key: 'COMPLETED', label: 'Завершён' },
  ]

  const statusDescriptions: Record<string, { title: string; text: string }> = {
    NEW: { title: 'Заказ создан', text: 'Ваш заказ оформлен. Перейдите к оплате, чтобы продолжить.' },
    WAITING_PAYMENT: { title: 'Ожидает оплаты', text: 'Оплатите заказ, чтобы мы начали его обработку. После оплаты модератор проверит платёж.' },
    PAYMENT_VERIFICATION: { title: 'Проверка платежа', text: 'Модератор проверяет ваш платёж. Обычно это занимает 5–15 минут. После подтверждения заказ будет обработан.' },
    PAID: { title: 'Оплачен', text: isKey
      ? 'Оплата подтверждена. Ожидайте выдачи ключа активации.'
      : 'Оплата подтверждена. Заказ передан модератору для активации.' },
    WAITING_ACTIVATION: { title: 'В очереди на активацию', text: 'Ваш заказ в очереди. Модератор возьмёт его в работу в ближайшее время.' },
    AWAITING_CREDENTIALS: { title: 'Требуются данные аккаунта', text: 'Модератор готов начать активацию. Введите данные вашего аккаунта ниже.' },
    CREDENTIALS_RECEIVED: { title: 'Данные получены', text: 'Модератор проверяет данные аккаунта и готовится к активации. Ожидайте.' },
    CREDENTIALS_INVALID: { title: 'Данные неверны', text: 'Введённые данные не подошли. Проверьте логин и пароль и отправьте их снова.' },
    AWAITING_2FA: { title: 'Ожидает код подтверждения', text: 'Модератор вошёл в аккаунт и ожидает код 2FA. Введите код с почты или приложения-аутентификатора.' },
    INVALID_2FA: { title: 'Код неверен', text: 'Введённый код не подошёл. Проверьте и отправьте правильный код подтверждения.' },
    ACTIVATING: { title: 'Активация в процессе', text: 'Модератор активирует товар на вашем аккаунте. Это может занять несколько минут. Не закрывайте приложение.' },
    ACTIVATED: { title: 'Активация завершена', text: 'Товар успешно активирован на вашем аккаунте. Заказ завершён!' },
    KEY_ISSUED: { title: 'Ключ выдан', text: 'Ваш ключ активации готов. Скопируйте его ниже.' },
    COMPLETED: { title: 'Заказ завершён', text: 'Спасибо за покупку! Оставьте отзыв, если вам понравилось.' },
    CANCELLED: { title: 'Заказ отменён', text: 'Заказ был отменён. Обратитесь в поддержку, если нужна помощь.' },
    REFUND_REQUESTED: { title: 'Запрошен возврат', text: 'Ваш запрос на возврат обрабатывается. Ожидайте ответа модератора.' },
    REFUNDED: { title: 'Возврат оформлен', text: 'Средства возвращены. Спасибо за понимание.' },
  }

  const steps = isKey ? keySteps : activationSteps
  const currentIndex = steps.findIndex(s => s.key === status)
  const progressPercent = currentIndex >= 0 ? Math.min(((currentIndex + 1) / steps.length) * 100, 100) : 0

  const isCompleted = status === 'COMPLETED' || status === 'KEY_ISSUED' || status === 'ACTIVATED'
  const isCancelled = status === 'CANCELLED' || status === 'REFUNDED'
  const desc = statusDescriptions[status] || { title: status, text: '' }

  if (isCancelled) {
    return (
      <div className="bg-[var(--tg-secondary)] rounded-xl p-4">
        <div className="flex items-center justify-between mb-2">
          <span className="text-xs text-[var(--tg-hint)]">Прогресс заказа</span>
          <span className="text-xs text-red-400">Отменён</span>
        </div>
        <div className="w-full h-2 bg-[var(--tg-card)] rounded-full overflow-hidden mb-3">
          <div className="h-full bg-red-500 rounded-full" style={{ width: `${progressPercent}%` }} />
        </div>
        <div className="border-t border-[var(--tg-border)] pt-3">
          <div className="text-sm font-semibold text-red-400 mb-1">{desc.title}</div>
          <div className="text-xs text-[var(--tg-hint)]">{desc.text}</div>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-[var(--tg-secondary)] rounded-xl p-4">
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs text-[var(--tg-hint)]">Прогресс заказа</span>
        <span className="text-xs text-[var(--tg-hint)]">{Math.round(progressPercent)}%</span>
      </div>
      <div className="w-full h-2 bg-[var(--tg-card)] rounded-full overflow-hidden mb-3">
        <div
          className={`h-full rounded-full transition-all duration-700 ease-out ${
            isCompleted
              ? 'bg-gradient-to-r from-green-500 to-green-400'
              : 'bg-gradient-to-r from-[var(--tg-button)] to-[var(--tg-button)]/70'
          }`}
          style={{ width: `${progressPercent}%` }}
        />
      </div>
      <div className="border-t border-[var(--tg-border)] pt-3">
        <div className={`text-sm font-semibold mb-1 ${
          isCompleted ? 'text-green-400' : 'text-[var(--tg-text)]'
        }`}>
          {desc.title}
        </div>
        <div className="text-xs text-[var(--tg-hint)] leading-relaxed">{desc.text}</div>
      </div>
    </div>
  )
}

export function OrderStatusPage() {
  const { id } = useParams<{ id: string }>()
  const nav = useNavigate()
  const { data: order, isLoading } = useOrder(id!)
  const sendCreds = useSendCredentials()
  const sendCode = useSend2FACode()

  const [login, setLogin] = useState('')
  const [password, setPassword] = useState('')
  const [code, setCode] = useState('')

  if (isLoading) return <div className="flex justify-center py-20"><Loader /></div>
  if (!order) return <div className="p-4 text-center text-tg-hint">Заказ не найден</div>

  const isActivation = order.delivery_method === 'activation'
  const completed = order.status === 'COMPLETED' || order.status === 'KEY_ISSUED' || order.status === 'ACTIVATED'
  const cancelled = order.status === 'CANCELLED' || order.status === 'REFUNDED'
  const chatEnabled = !['NEW', 'WAITING_PAYMENT', 'PAYMENT_VERIFICATION'].includes(order.status)

  const handleSendCredentials = async () => {
    if (!login || !password) return
    await sendCreds.mutateAsync({ orderId: order.id, platform: order.product?.platform || 'xbox', login, password })
  }

  const handleSendCode = async () => {
    if (!code) return
    await sendCode.mutateAsync({ orderId: order.id, code })
  }

  return (
    <div className="pb-24">
      <Header title="Статус заказа" onBack={() => nav(-1)} />
      <div className="p-4 space-y-4">
        <div className="bg-[var(--tg-secondary)] rounded-xl p-4">
          <div className="flex justify-between items-center mb-2">
            <span className="text-tg-hint text-xs">Заказ #{order.id.substring(0, 8)}</span>
            <StatusBadge status={order.status} />
          </div>
          <div className="font-semibold text-base mb-1">{order.product?.title || 'Товар'}</div>
          <div className="flex justify-between items-center">
            <div className="text-xs text-[var(--tg-hint)]">
              {isActivation ? '🔐 Активация' : '🔑 Ключ'}
              {order.quantity > 1 && ` • ${order.quantity} шт`}
            </div>
            <span className="font-bold text-[var(--tg-button)] text-lg">
              {order.payment_amount ? formatPrice(order.payment_amount) : ''}
            </span>
          </div>
        </div>

        <ProgressStepper status={order.status} deliveryMethod={order.delivery_method} />

        {order.status === 'WAITING_PAYMENT' && (
          <Button fullWidth onClick={() => nav(`/checkout/${order.id}`)}>
            💳 Перейти к оплате
          </Button>
        )}

        {order.status === 'AWAITING_CREDENTIALS' && (
          <div className="bg-[var(--tg-secondary)] rounded-xl p-4 space-y-3">
            <h3 className="font-bold text-base">🔐 Введите данные аккаунта</h3>
            <p className="text-sm text-[var(--tg-hint)]">
              Эти данные нужны только для активации. Они шифруются и не передаются третьим лицам.
            </p>
            <input
              type="text" placeholder="Логин (email)"
              value={login} onChange={(e) => setLogin(e.target.value)}
              className="w-full px-3 py-2.5 rounded-lg bg-[var(--tg-input-bg)] border border-[var(--tg-input-border)] text-[var(--tg-input-text)] placeholder-[var(--tg-input-placeholder)] text-sm"
            />
            <input
              type="password" placeholder="Пароль"
              value={password} onChange={(e) => setPassword(e.target.value)}
              className="w-full px-3 py-2.5 rounded-lg bg-[var(--tg-input-bg)] border border-[var(--tg-input-border)] text-[var(--tg-input-text)] placeholder-[var(--tg-input-placeholder)] text-sm"
            />
            <Button fullWidth onClick={handleSendCredentials} loading={sendCreds.isPending}>
              📤 Отправить данные
            </Button>
          </div>
        )}

        {order.status === 'CREDENTIALS_INVALID' && (
          <div className="bg-[var(--tg-secondary)] rounded-xl p-4 space-y-3">
            <h3 className="font-bold text-base">🔐 Введите данные аккаунта</h3>
            <div className="bg-[var(--tg-alert-error-bg)] border border-[var(--tg-alert-error-border)] rounded-lg p-3">
              <p className="text-sm text-[var(--tg-alert-error-text)] font-medium">❌ Данные не подошли</p>
              {order.cancelled_reason && (
                <p className="text-xs text-[var(--tg-alert-error-text)]/80 mt-1">Причина: {order.cancelled_reason}</p>
              )}
            </div>
            <p className="text-sm text-[var(--tg-hint)]">
              Проверьте логин и пароль и отправьте их снова.
            </p>
            <input
              type="text" placeholder="Логин (email)"
              value={login} onChange={(e) => setLogin(e.target.value)}
              className="w-full px-3 py-2.5 rounded-lg bg-[var(--tg-input-bg)] border border-[var(--tg-input-border)] text-[var(--tg-input-text)] placeholder-[var(--tg-input-placeholder)] text-sm"
            />
            <input
              type="password" placeholder="Пароль"
              value={password} onChange={(e) => setPassword(e.target.value)}
              className="w-full px-3 py-2.5 rounded-lg bg-[var(--tg-input-bg)] border border-[var(--tg-input-border)] text-[var(--tg-input-text)] placeholder-[var(--tg-input-placeholder)] text-sm"
            />
            <Button fullWidth onClick={handleSendCredentials} loading={sendCreds.isPending}>
              📤 Отправить данные
            </Button>
          </div>
        )}

        {order.status === 'AWAITING_2FA' && (
          <div className="bg-[var(--tg-secondary)] rounded-xl p-4 space-y-3">
            <h3 className="font-bold text-base">🔑 Код подтверждения</h3>
            <p className="text-sm text-[var(--tg-hint)]">
              Модератор ожидает код для входа. Введите код с почты или приложения-аутентификатора.
            </p>
            <input
              type="text" placeholder="Код подтверждения (6 цифр)"
              value={code} onChange={(e) => setCode(e.target.value)}
              className="w-full px-3 py-2.5 rounded-lg bg-[var(--tg-input-bg)] border border-[var(--tg-input-border)] text-[var(--tg-input-text)] placeholder-[var(--tg-input-placeholder)] text-sm text-center text-lg tracking-widest"
              maxLength={6}
            />
            <Button fullWidth onClick={handleSendCode} loading={sendCode.isPending}>
              🔑 Отправить код
            </Button>
          </div>
        )}

        {order.status === 'INVALID_2FA' && (
          <div className="bg-[var(--tg-secondary)] rounded-xl p-4 space-y-3">
            <h3 className="font-bold text-base">🔑 Код подтверждения</h3>
            <div className="bg-[var(--tg-alert-error-bg)] border border-[var(--tg-alert-error-border)] rounded-lg p-3">
              <p className="text-sm text-[var(--tg-alert-error-text)] font-medium">❌ Код не подошёл</p>
              {order.cancelled_reason && (
                <p className="text-xs text-[var(--tg-alert-error-text)]/80 mt-1">Причина: {order.cancelled_reason}</p>
              )}
            </div>
            <p className="text-sm text-[var(--tg-hint)]">
              Введите правильный код с почты или приложения-аутентификатора.
            </p>
            <input
              type="text" placeholder="Код подтверждения (6 цифр)"
              value={code} onChange={(e) => setCode(e.target.value)}
              className="w-full px-3 py-2.5 rounded-lg bg-[var(--tg-input-bg)] border border-[var(--tg-input-border)] text-[var(--tg-input-text)] placeholder-[var(--tg-input-placeholder)] text-sm text-center text-lg tracking-widest"
              maxLength={6}
            />
            <Button fullWidth onClick={handleSendCode} loading={sendCode.isPending}>
              🔑 Отправить код
            </Button>
          </div>
        )}

        {order.status === 'ACTIVATING' && (
          <div className="bg-[var(--tg-alert-warning-bg)] border border-[var(--tg-alert-warning-border)] rounded-xl p-6 text-center space-y-3">
            <div className="text-4xl">⚙️</div>
            <h3 className="font-bold text-[var(--tg-alert-warning-text)]">Активация в процессе</h3>
            <p className="text-sm text-[var(--tg-alert-warning-text)]/80">
              Модератор активирует товар. Это может занять несколько минут.
            </p>
            <Loader />
          </div>
        )}

        {chatEnabled && <OrderChat orderId={order.id} />}

        {completed && (
          <div className="bg-[var(--tg-alert-success-bg)] border border-[var(--tg-alert-success-border)] rounded-xl p-6 text-center space-y-4">
            <div className="text-4xl mb-2">🎉</div>
            <h2 className="text-lg font-bold text-[var(--tg-alert-success-text)] mb-2">Заказ завершён!</h2>
            <p className="text-sm text-[var(--tg-alert-success-text)]/80 mb-4">
              {order.delivery_method === 'key' ? 'Ваш ключ активации готов.' : 'Товар успешно активирован на вашем аккаунте.'}
            </p>

            {order.delivery_method === 'key' && order.key_value && (
              <div className="bg-[var(--tg-secondary)] rounded-lg p-4 space-y-2">
                <p className="text-xs text-[var(--tg-hint)] uppercase tracking-wide">Ваш ключ активации</p>
                <div className="font-mono text-sm bg-[var(--tg-bg)] rounded-lg p-3 break-all select-all">
                  {order.key_value}
                </div>
                <button
                  onClick={() => navigator.clipboard.writeText(order.key_value!)}
                  className="text-xs text-[var(--tg-link)] hover:opacity-80"
                >
                  📋 Скопировать
                </button>
              </div>
            )}

            <div className="bg-[var(--tg-secondary)] rounded-lg p-4 space-y-3 text-left">
              <p className="text-sm font-medium text-[var(--tg-text)]">💬 Оставьте отзыв — это поможет другим покупателям!</p>
              <div className="space-y-2">
                <a href="https://t.me/coin_mint_reviews" target="_blank" rel="noopener noreferrer"
                   className="flex items-center gap-2 text-sm text-[var(--tg-link)] hover:opacity-80">
                  <span>📝</span> Оставить отзыв в канале
                </a>
                <a href="https://t.me/coin_mint_chat" target="_blank" rel="noopener noreferrer"
                   className="flex items-center gap-2 text-sm text-[var(--tg-link)] hover:opacity-80">
                  <span>💬</span> Чат поддержки
                </a>
              </div>
            </div>

            <Button onClick={() => nav('/')}>🛒 Вернуться в магазин</Button>
          </div>
        )}

        {cancelled && (
          <div className="bg-[var(--tg-alert-error-bg)] border border-[var(--tg-alert-error-border)] rounded-xl p-4 text-center">
            <div className="text-4xl mb-2">❌</div>
            <h2 className="text-lg font-bold text-[var(--tg-alert-error-text)] mb-2">Заказ отменён</h2>
            {order.cancelled_reason && (
              <p className="text-sm text-[var(--tg-alert-error-text)]/80">Причина: {order.cancelled_reason}</p>
            )}
            <Button className="mt-4" variant="outline" onClick={() => nav('/')}>
              🛒 В магазин
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}

function OrderChat({ orderId }: { orderId: string }) {
  const [messages, setMessages] = useState<any[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const pollRef = useRef<number>()
  const msgCountRef = useRef(0)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  const fetchMessages = async () => {
    try {
      const { data } = await api.get(`/orders/${orderId}/chat`)
      if (data.length > msgCountRef.current) {
        msgCountRef.current = data.length
        scrollToBottom()
      }
      setMessages(data)
    } catch {}
  }

  useEffect(() => {
    fetchMessages()
    pollRef.current = window.setInterval(fetchMessages, 3000)
    return () => { if (pollRef.current) clearInterval(pollRef.current) }
  }, [orderId])

  const sendMessage = async () => {
    if (!input.trim() || loading) return
    setLoading(true)
    try {
      await api.post(`/orders/${orderId}/chat`, { message: input.trim() })
      setInput('')
      fetchMessages()
    } catch {}
    setLoading(false)
  }

  const formatTime = (date: string) => {
    return new Date(date).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
  }

  const isKeyMessage = (msg: any) => {
    return msg.sender_type === 'admin' && msg.message.includes('Ваш ключ активации:')
  }

  const extractKey = (msg: any) => {
    const match = msg.message.match(/Ваш ключ активации:\s*(.+)/)
    return match ? match[1].trim() : ''
  }

  const copyKey = (key: string) => {
    navigator.clipboard.writeText(key)
  }

  return (
    <div className="bg-[var(--tg-secondary)] rounded-xl overflow-hidden">
      <div className="px-4 py-3 border-b border-[var(--tg-border)]">
        <h3 className="font-bold text-base">💬 Чат поддержки</h3>
      </div>
      <div className="h-64 overflow-y-auto p-4 space-y-3">
        {messages.length === 0 && (
          <p className="text-center text-sm text-[var(--tg-hint)] py-8">Напишите сообщение, если у вас есть вопросы</p>
        )}
        {messages.map((msg) => {
          const isKey = isKeyMessage(msg)
          if (isKey) {
            const key = extractKey(msg)
            return (
              <div key={msg.id} className="flex justify-start">
                <div className="max-w-[90%] rounded-2xl px-3 py-3 bg-[var(--tg-alert-success-bg)] border border-[var(--tg-alert-success-border)] rounded-bl-md">
                  <div className="text-sm font-medium text-[var(--tg-alert-success-text)] mb-2">🔑 Ваш ключ активации:</div>
                  <div className="font-mono text-sm bg-[var(--tg-bg)] rounded-lg p-2 break-all select-all text-[var(--tg-text)]">
                    {key}
                  </div>
                  <button
                    onClick={() => copyKey(key)}
                    className="text-xs text-[var(--tg-link)] hover:opacity-80 mt-2"
                  >
                    📋 Скопировать
                  </button>
                  <div className="text-[10px] mt-1 text-[var(--tg-hint)]">
                    {formatTime(msg.created_at)}
                  </div>
                </div>
              </div>
            )
          }
          return (
            <div key={msg.id} className={`flex ${msg.sender_type === 'user' ? 'justify-end' : 'justify-start'}`}>
              <div className={`max-w-[80%] rounded-2xl px-3 py-2 ${
                msg.sender_type === 'user'
                  ? 'bg-[var(--tg-button)] text-white rounded-br-md'
                  : 'bg-[var(--tg-card)] text-[var(--tg-text)] rounded-bl-md'
              }`}>
                <div className="text-sm">{msg.message}</div>
                <div className={`text-[10px] mt-1 ${msg.sender_type === 'user' ? 'text-white/60' : 'text-[var(--tg-hint)]'}`}>
                  {formatTime(msg.created_at)}
                </div>
              </div>
            </div>
          )
        })}
        <div ref={messagesEndRef} />
      </div>
      <div className="px-4 py-3 border-t border-[var(--tg-border)] flex gap-2">
        <input
          type="text"
          placeholder="Сообщение..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
          className="flex-1 px-3 py-2 rounded-lg bg-[var(--tg-input-bg)] border border-[var(--tg-input-border)] text-[var(--tg-input-text)] placeholder-[var(--tg-input-placeholder)] text-sm"
        />
        <Button onClick={sendMessage} loading={loading} disabled={!input.trim()}>
          📤
        </Button>
      </div>
    </div>
  )
}
