export type FaqItem = {
  id: string
  question: string
  answer: string
  sort_order: number
}

/** Актуальный FAQ — fallback, если API ещё не обновлён. */
export const DEFAULT_FAQ: FaqItem[] = [
  {
    id: 'default-faq-1',
    question: 'Как купить игру или подписку?',
    answer:
      'Выберите товар в каталоге, нажмите «Оформить с менеджером» и напишите в Telegram. Менеджер подтвердит заказ, подскажет способ оплаты и оформит покупку на ваш аккаунт PlayStation или Xbox.',
    sort_order: 1,
  },
  {
    id: 'default-faq-2',
    question: 'Актуальны ли цены на сайте?',
    answer:
      'Цены в каталоге ориентировочные и могут меняться из‑за курса валют и акций в PlayStation Store / Xbox Store. Перед оплатой обязательно уточните итоговую стоимость у менеджера.',
    sort_order: 2,
  },
  {
    id: 'default-faq-3',
    question: 'Какие регионы PlayStation поддерживаются?',
    answer: 'PlayStation — Турция 🇹🇷 и Украина 🇺🇦. Для украинского региона итоговую цену уточняет менеджер.',
    sort_order: 3,
  },
  {
    id: 'default-faq-4',
    question: 'Какой регион у игр Xbox?',
    answer: 'Xbox — регион США 🇺🇸 (US).',
    sort_order: 4,
  },
  {
    id: 'default-faq-5',
    question: 'Как происходит оплата?',
    answer:
      'После подтверждения заказа менеджер пришлёт реквизиты и инструкцию в чате Telegram. Оплата только после согласования с менеджером.',
    sort_order: 5,
  },
  {
    id: 'default-faq-6',
    question: 'Нужно ли передавать данные аккаунта?',
    answer:
      'Да, для покупки на ваш аккаунт PSN или Microsoft менеджер запросит необходимые данные. Передавайте их только в официальном чате поддержки магазина.',
    sort_order: 6,
  },
  {
    id: 'default-faq-7',
    question: 'Сколько ждать ответ менеджера?',
    answer: 'Обычно отвечаем в течение 15–30 минут в рабочее время. В пиковые часы — чуть дольше.',
    sort_order: 7,
  },
  {
    id: 'default-faq-8',
    question: 'Что такое предзаказ?',
    answer:
      'Это заказ игры до официального выхода. Менеджер зафиксирует заказ и оформит покупку после релиза или в день выхода игры.',
    sort_order: 8,
  },
  {
    id: 'default-faq-9',
    question: 'Можно ли отменить заказ или вернуть деньги?',
    answer:
      'Да, если покупка ещё не оформлена на ваш аккаунт. Напишите менеджеру в Telegram — рассмотрим запрос в течение 24 часов.',
    sort_order: 9,
  },
]

function isLegacyFaq(items: FaqItem[]): boolean {
  return items.some((item) => /ключ/i.test(item.question) || /ключ/i.test(item.answer))
}

function withoutRemovedFaq(items: FaqItem[]): FaqItem[] {
  return items.filter((item) => !/продаёте ключи/i.test(item.question))
}

export function resolveFaqItems(items?: FaqItem[] | null): FaqItem[] {
  if (!items?.length || isLegacyFaq(items)) {
    return DEFAULT_FAQ
  }
  return withoutRemovedFaq(items)
}
