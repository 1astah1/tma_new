# TMA Shop

Магазин игр, внутриигровой валюты и подписок (PS4/PS5/Xbox/PC) в виде Telegram Mini App.

## Состав

| Каталог | Что это | Стек |
|---|---|---|
| `tma-backend` | API, парсер каталога, Telegram-бот | Go 1.25, chi, sqlx, PostgreSQL |
| `tma-frontend` | Мини-апп для покупателя | React 18, Vite, Tailwind, zustand, react-query |
| `admin-panel` | Админка | react-admin, MUI |
| `db-backup` | Периодический `pg_dump` с ротацией | postgres:15-alpine |
| `monitoring` | Конфиг Prometheus | — |

## Локальный запуск

```bash
docker compose up -d --build
```

- TMA — http://localhost
- Админка — http://localhost:8081
- API — http://localhost:8080
- Prometheus — http://localhost:9090, Grafana — http://localhost:3000

Без Docker:

```bash
cp tma-backend/.env.example tma-backend/.env   # заполнить JWT_SECRET и ACCOUNT_ENCRYPTION_KEY
migrate -path tma-backend/migrations -database "$DATABASE_URL" up
cd tma-backend && go run ./cmd/server

cd tma-frontend && npm ci && npm run dev
cd admin-panel  && npm ci && npm run dev
```

## Миграции

`golang-migrate`, каталог `tma-backend/migrations`, пары `NNN_name.up.sql` / `NNN_name.down.sql`.
В Docker миграции накатываются автоматически из `entrypoint.sh`.

Перед коммитом новой миграции:

```bash
./scripts/check-migrations.sh
```

Скрипт валит сборку при дублях версий, отсутствующем `down` и файлах без направления —
на всём этом `golang-migrate` падает целиком.

Разовые сиды лежат в `tma-backend/seeds/` и применяются вручную.

## Наполнение каталога по списку

`tma-backend/seeds/wanted_games.csv` — список игр, которые мы хотим продавать
(выгрузка таблицы). Список задаёт **только намерение**: какая игра и под какие
платформы. Всё остальное берётся из магазинов на момент запуска:

| Что | Откуда | Как считается |
|---|---|---|
| Цена PS Турция | PS Store TR | `TurkeyNominalPrice`: округление TL вверх до номинала × наценка (по умолчанию 2.2) |
| Цена PS Украина | PS Store UA | `UkrainePrice`: UAH × наценка (2.3) |
| Цена Xbox/PC | Xbox Store US | `XboxUSAPrice`: USD × множитель (80) |
| Описание | PS RU → PS UA(ru) → PS US(en) / Xbox ru-RU | служебная преамбула стора вырезается |
| Картинка | обложка стора | — |
| Раздел | дата релиза и флаг предзаказа из стора | предзаказ / новинка (90 дней) / каталог |

Наценки и минимальная цена живут в настройках админки (`pricing_formulas`),
импорт берёт их оттуда, а не из констант.

```bash
docker exec tma_backend ./import-wanted -file seeds/wanted_games.csv -report /app/uploads/wanted_report.json
docker exec tma_backend ./import-wanted -file seeds/wanted_games.csv -limit 20   # обкатать на срезе
docker exec tma_backend ./import-wanted -file seeds/wanted_games.csv -dry        # только разбор списка
```

Одна игра → одна карточка: все платформы получают общий `title_key` и общее
название, а страница товара склеивает их в один `edition_catalog`
с ценами по каждому гео. Апгрейды между поколениями, дополнения, пропуски и
валюта отсеиваются — их легко принять за саму игру.

## Проверка перед пушем

```bash
cd tma-backend && gofmt -l . && go vet ./... && go test ./...
cd tma-frontend && npm run build
cd admin-panel  && npm run build
./scripts/check-migrations.sh
```

## Деплой

Backend и frontend — Railway (`railway.json`, билдер RAILPACK). Для frontend есть также `vercel.json`.
CI (`.github/workflows/ci.yml`) гоняет форматирование, vet, тесты и сборку обоих фронтов.

### Свой сервер (51.38.154.221)

Стек живёт в `/opt/tma`, конфиг — `docker-compose.prod.yml` (postgres + backend + frontend + admin,
без Prometheus/Grafana — на сервере всего 2 ГБ RAM). Секреты — в `/opt/tma/.env`.

```bash
rsync -az --exclude .git --exclude node_modules --exclude dist ./ root@51.38.154.221:/opt/tma/
ssh root@51.38.154.221 /opt/tma/redeploy.sh
```

- TMA — https://tma.happ.xin (Caddy, конфиг в `/etc/caddy/tma.caddy`, сертификат Let's Encrypt)
- Бот — [@CoinMintShopBot](https://t.me/CoinMintShopBot), long polling, кнопка меню открывает мини-апп
- Админка — только через SSH-туннель: `ssh -L 8091:127.0.0.1:18091 root@51.38.154.221`

`VITE_BOT_USERNAME` подставляется на этапе сборки фронта (build-arg из `BOT_USERNAME` в `.env`) —
после смены бота фронт надо пересобрать, рестарта мало.

Миграции накатываются самим сервером при старте, отдельный шаг не нужен.
