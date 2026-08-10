import { useEffect, useRef, useState } from 'react'
import { Button } from '../ui/Button'
import { useToast } from '../ui/Toast'
import api from '../../services/api'

export function OrderChat({ orderId }: { orderId: string }) {
  const toast = useToast()
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
    } catch {
      toast.toast('Не удалось загрузить чат', 'error')
    }
  }

  useEffect(() => {
    fetchMessages()
    pollRef.current = window.setInterval(fetchMessages, 3000)
    return () => {
      if (pollRef.current) clearInterval(pollRef.current)
    }
  }, [orderId])

  const sendMessage = async () => {
    if (!input.trim() || loading) return
    setLoading(true)
    try {
      await api.post(`/orders/${orderId}/chat`, { message: input.trim() })
      setInput('')
      fetchMessages()
    } catch {
      toast.toast('Не удалось отправить сообщение', 'error')
    }
    setLoading(false)
  }

  const formatTime = (date: string) =>
    new Date(date).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })

  return (
    <div className="overflow-hidden rounded-xl bg-[var(--tg-secondary)]">
      <div className="border-b border-[var(--tg-border)] px-4 py-3">
        <h3 className="text-base font-bold">💬 Чат по заказу</h3>
        <p className="mt-1 text-xs text-[var(--tg-hint)]">
          Переписка с менеджером. Основное общение — в Telegram.
        </p>
      </div>
      <div className="h-64 space-y-3 overflow-y-auto p-4">
        {messages.length === 0 ? (
          <p className="py-8 text-center text-sm text-[var(--tg-hint)]">
            Сообщений пока нет. Напишите менеджеру, если есть вопросы по заказу.
          </p>
        ) : null}
        {messages.map((msg) => (
          <div key={msg.id} className={`flex ${msg.sender_type === 'user' ? 'justify-end' : 'justify-start'}`}>
            <div
              className={`max-w-[80%] rounded-2xl px-3 py-2 ${
                msg.sender_type === 'user'
                  ? 'rounded-br-md bg-[var(--tg-button)] text-white'
                  : 'rounded-bl-md bg-[var(--tg-card)] text-[var(--tg-text)]'
              }`}
            >
              <div className="text-sm whitespace-pre-wrap">{msg.message}</div>
              <div className={`mt-1 text-[10px] ${msg.sender_type === 'user' ? 'text-white/60' : 'text-[var(--tg-hint)]'}`}>
                {formatTime(msg.created_at)}
              </div>
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>
      <div className="flex gap-2 border-t border-[var(--tg-border)] px-4 py-3">
        <input
          type="text"
          placeholder="Сообщение..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
          className="flex-1 rounded-lg border border-[var(--tg-input-border)] bg-[var(--tg-input-bg)] px-3 py-2 text-sm text-[var(--tg-input-text)] placeholder-[var(--tg-input-placeholder)]"
        />
        <Button onClick={sendMessage} loading={loading} disabled={!input.trim()}>
          📤
        </Button>
      </div>
    </div>
  )
}
