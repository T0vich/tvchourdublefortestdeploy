# Цепочка обмена

Сервис многосторонних бартерных обменов: пользователь указывает, что отдаёт и что
хочет получить, а сервис подбирает маршрут обмена через чужие объявления.

Кейс 6 хакатона Avito Start. Это **тестовая копия** командного репозитория
`GavrRMANN/xakaton_avito`, доведённая до рабочего состояния и развёрнутая
публично. Перечень отличий от оригинала — в [TEAM-REVIEW.md](TEAM-REVIEW.md).

## Демо

| | |
| --- | --- |
| Веб-приложение | https://tvchbmen-front.vercel.app |
| API | https://tvchbmen-api.vercel.app |
| Swagger UI | https://tvchbmen-api.vercel.app/swagger/index.html |

Демо-пользователи: `user1@test.com`, `user2@test.com`, `user3@test.com`,
`user4@test.com` — пароль у всех `demo1234`.

В базе заранее разложен сценарий с готовой цепочкой:

| Пользователь | Отдаёт | Хочет получить |
| --- | --- | --- |
| user1 | iPhone 15 | Видеокарту |
| user2 | GT Avalanche | Телефон |
| user3 | MacBook Pro | Велосипед |
| user4 | RTX 4080 | Ноутбук |

Войдите как `user1` и запросите цепочку до RTX 4080 — сервис найдёт маршрут
`RTX 4080 → MacBook Pro → GT Avalanche → iPhone 15`.

## Запуск

Нужен только Docker.

```bash
docker compose up --build
```

Готово: фронт на http://localhost:3000, API на http://localhost:8080,
Swagger на http://localhost:8080/swagger/index.html.

Схема базы и демо-данные накатываются автоматически при первом старте.

Полный сброс вместе с данными:

```bash
docker compose down -v
```

### Проверить, что API поднялся

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/products?limit=5
```

Второй запрос отдаёт каталог **без токена** — витрина публичная.

### Пройти сценарий обмена целиком

```bash
# 1. Вход
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user1@test.com","password":"demo1234"}' | jq -r .token)

# 2. Найти id видеокарты
RTX=$(curl -s 'http://localhost:8080/api/v1/products?limit=20' \
  | jq -r '.[] | select(.title | test("RTX")) | .product_id')

# 3. Запросить цепочку обмена
curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/search/chain?target_product_id=$RTX&max_depth=10" \
  | jq '.chain[].title'
```

## Запуск без Docker

Нужны Go 1.26+, Node 22+ и доступный Postgres 16+.

```bash
# Бэкенд
cd back
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/trade_chain?sslmode=disable"
export JWT_SECRET="любая-длинная-строка"
go run ./cmd/migrate     # накатить схему и демо-данные
go run ./cmd/server

# Фронтенд, в другом терминале
cd front
npm ci
echo 'VITE_API_BASE_URL=http://localhost:8080' > .env.local
npm run dev
```

## Переменные окружения

### Бэкенд

| Переменная | Обязательна | По умолчанию | Назначение |
| --- | --- | --- | --- |
| `DATABASE_URL` | да | — | строка подключения к Postgres |
| `JWT_SECRET` | на проде | отладочный ключ | ключ подписи токенов |
| `PORT` | нет | `8080` | порт HTTP-сервера |
| `CORS_ALLOWED_ORIGINS` | нет | `*` | список источников через запятую |
| `DB_MAX_CONNS` | нет | `5` | размер пула соединений |
| `PUBLIC_HOST` | нет | — | домен для Swagger UI на деплое |

### Фронтенд

| Переменная | По умолчанию | Назначение |
| --- | --- | --- |
| `VITE_API_BASE_URL` | `http://localhost:3001` | адрес API без `/api/v1` |

## Стек

Go 1.26 · chi · pgx · Postgres 16 · React 18 · TypeScript · Redux Toolkit Query ·
Vite · Feature-Sliced Design · Docker Compose

## Структура

```
back/
  cmd/server/          точка входа HTTP-сервера
  cmd/migrate/         накат SQL на внешнюю базу
  internal/httpapi/    роутер, обработчики, CORS, проверки прав
  internal/auth/       JWT и middleware авторизации
  internal/service/    бизнес-логика
  internal/repository/ доступ к данным
  internal/search/     подбор цепочки обмена (обход графа)
  internal/domain/     модели
  infrastructure/migrations/  схема и демо-данные
front/
  src/app/             провайдеры, роутер, стор
  src/pages/           страницы
  src/features/        формы и сценарии
  src/entities/        сущности и обращения к API
  src/shared/          переиспользуемый UI и утилиты
docs/API.md            описание целевого контракта API
```

## Как устроен подбор цепочки

У каждого объявления есть список желаний — категории, которые владелец готов
принять взамен. Это задаёт граф: из товара A ведёт ребро в товар B, если
владелец B согласен принять категорию A.

Поиск идёт обходом в ширину от желаемого товара к тому, что есть у пользователя,
с ограничением глубины (`max_depth`). Возвращается сама цепочка и её длина.

Полнотекстовый поиск по каталогу работает на `tsvector` с учётом опечаток через
триграммы `pg_trgm`, поэтому «ноутбук» находится и при неточном вводе.

## Разработка

```bash
cd back
go test ./...                                        # тесты
golangci-lint run --config ../.golangci.yaml ./...   # линтер

cd front
npm run lint
npm run build
npm run storybook
```

CI прогоняет сборку и линт фронта, линт, тесты и сборку бэка, а также сборку
всех образов через `docker compose build`.
