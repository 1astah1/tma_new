/**
 * Человеческий текст ошибки оформления. Без него пользователь видел молчание:
 * запрос падал, а на экране ничего не менялось.
 */
export function orderErrorMessage(err: unknown): string {
  const response = (err as { response?: { status?: number; data?: { error?: { message?: string } } } })?.response

  if (response?.status === 401) {
    return 'Сессия истекла. Закройте и снова откройте приложение из Telegram.'
  }
  if (response?.status === 429) {
    return 'Слишком много запросов подряд. Подождите немного и попробуйте снова.'
  }

  const message = response?.data?.error?.message
  if (message) return message

  return 'Не удалось оформить заказ. Проверьте связь и попробуйте ещё раз.'
}
