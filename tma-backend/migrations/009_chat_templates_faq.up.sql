-- Chat templates and FAQ
CREATE TABLE IF NOT EXISTS chat_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'general',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS faq_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question VARCHAR(255) NOT NULL,
    answer TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_templates_category ON chat_templates(category);
CREATE INDEX IF NOT EXISTS idx_faq_active ON faq_items(is_active, sort_order);

-- Seed FAQ
INSERT INTO faq_items (question, answer, sort_order) VALUES
('Как получить ключ?', 'После оплаты ключ будет отправлен в этот чат автоматически. Также он отобразится на странице заказа.', 1),
('Как оплатить заказ?', 'Выберите способ оплаты (СБП, карта, криптовалюта) и следуйте инструкциям на экране оплаты.', 2),
('Сколько времени занимает активация?', 'Обычно активация занимает от 5 до 30 минут. В редких случаях может занять до 24 часов.', 3),
('Что делать если ключ не работает?', 'Напишите в этот чат — мы поможем решить проблему или заменим ключ.', 4),
('Как вернуть деньги?', 'Перейдите в заказ и нажмите "Запросить возврат". Мы рассмотрим запрос в течение 24 часов.', 5),
('Можно ли отменить заказ?', 'Да, если заказ ещё не оплачен или не начата активация. Напишите в чат для отмены.', 6);

-- Seed templates
INSERT INTO chat_templates (title, message, category) VALUES
('Приветствие', 'Здравствуйте! Чем могу помочь?', 'greeting'),
('Ключ выдан', 'Ваш ключ активации: {key}. Пожалуйста, сохраните его.', 'order'),
('Оплата подтверждена', 'Ваша оплата подтверждена. Начинаем обработку заказа.', 'order'),
('Ожидание', 'Пожалуйста, подождите. Мы обрабатываем ваш запрос.', 'general'),
('Запрос данных', 'Для активации нам нужны данные от вашего аккаунта. Пожалуйста, отправьте логин и пароль.', 'order'),
('Проблема решена', 'Проблема решена. Если у вас есть ещё вопросы — обращайтесь!', 'general'),
('Возврат', 'Ваш запрос на возврат принят. Обработка займёт до 24 часов.', 'order');
