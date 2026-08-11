-- Миграции 002 и 009 были залиты в UTF-8, перекодированном как cp1251 (мойибейк),
-- поэтому в проде реквизиты и шаблоны чата показываются кракозябрами.
-- Чиним значения по месту, не трогая то, что админы уже поправили руками.

UPDATE settings
SET value = replace(
        replace(value::text, 'РђР»СЊС„Р°-Р‘Р°РЅРє', 'Альфа-Банк'),
        'РћР»РµСЃСЏ Рљ.', 'Олеся К.'
    )::jsonb,
    updated_at = NOW()
WHERE key = 'payment_details'
  AND value::text LIKE '%РђР»СЊС„Р°%';

UPDATE chat_templates SET title = 'Приветствие', message = 'Здравствуйте! Чем могу помочь?'
WHERE title = 'РџСЂРёРІРµС‚СЃС‚РІРёРµ';
UPDATE chat_templates SET title = 'Ключ выдан', message = 'Ваш ключ активации: {key}. Пожалуйста, сохраните его.'
WHERE title = 'РљР»СЋС‡ РІС‹РґР°РЅ';
UPDATE chat_templates SET title = 'Оплата подтверждена', message = 'Ваша оплата подтверждена. Начинаем обработку заказа.'
WHERE title = 'РћРїР»Р°С‚Р° РїРѕРґС‚РІРµСЂР¶РґРµРЅР°';
UPDATE chat_templates SET title = 'Ожидание', message = 'Пожалуйста, подождите. Мы обрабатываем ваш запрос.'
WHERE title = 'РћР¶РёРґР°РЅРёРµ';
UPDATE chat_templates SET title = 'Запрос данных', message = 'Для активации нам нужны данные от вашего аккаунта. Пожалуйста, отправьте логин и пароль.'
WHERE title = 'Р—Р°РїСЂРѕСЃ РґР°РЅРЅС‹С…';
UPDATE chat_templates SET title = 'Проблема решена', message = 'Проблема решена. Если у вас есть ещё вопросы — обращайтесь!'
WHERE title = 'РџСЂРѕР±Р»РµРјР° СЂРµС€РµРЅР°';
UPDATE chat_templates SET title = 'Возврат', message = 'Ваш запрос на возврат принят. Обработка займёт до 24 часов.'
WHERE title = 'Р’РѕР·РІСЂР°С‚';
