-- Возврат удалённого вопроса FAQ.
INSERT INTO faq_items (question, answer, sort_order, is_active)
SELECT 'Вы продаёте ключи?',
       'Нет, покупка оформляется менеджером на ваш аккаунт PlayStation или Xbox.',
       999,
       true
WHERE NOT EXISTS (SELECT 1 FROM faq_items WHERE question = 'Вы продаёте ключи?');
