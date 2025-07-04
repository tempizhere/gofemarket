# Gofemarket

Gofemarket — это веб-приложение на Go для управления системой лояльности, позволяющее пользователям регистрировать заказы, получать начисления и списывать баллы. Проект реализует REST API с авторизацией, взаимодействием с базой данных PostgreSQL и внешней системой начислений.

## Требования

- Go 1.23 или выше
- PostgreSQL
- Бинарник системы начислений (`accrual_linux_amd64`)

## Установка

1. Склонируйте репозиторий:
   ```bash
   git clone https://github.com/tempizhere/gofemarket.git
   cd gofemarket
   ```

2. Инициализируйте модуль Go:
   ```bash
   go mod init github.com/tempizhere/gofemarket
   ```

3. Установите зависимости:
   ```bash
   go mod tidy
   ```

4. Создайте файл `.env` в корне проекта с конфигурацией:
   ```
   RUN_ADDRESS=localhost:8080
   DATABASE_URI=postgresql://postgres:postgres@localhost:5432/praktikum?sslmode=disable
   ACCRUAL_SYSTEM_ADDRESS=http://localhost:8000
   ```

## Запуск

1. Скомпилируйте приложение:
   ```bash
   go build -o cmd/gophermart/gophermart ./cmd/gophermart
   ```

2. Запустите сервер:
   ```bash
   ./cmd/gophermart/gophermart
   ```

3. Запустите систему начислений (если требуется):
   ```bash
   ./cmd/accrual/accrual_linux_amd64 -a localhost:8000
   ```

## Структура проекта

- `cmd/gophermart/main.go` — точка входа приложения.
- `internal/api` — обработчики HTTP-запросов, middleware и интерфейсы.
- `internal/client` — клиент для взаимодействия с системой начислений.
- `internal/config` — загрузка конфигурации из переменных окружения.
- `internal/logger` — настройка логирования с использованием `zap`.
- `internal/model` — структуры данных и ошибки.
- `internal/repository` — работа с базой данных PostgreSQL.
- `internal/service` — бизнес-логика приложения.

## Основные функции

- **Регистрация и авторизация**: Пользователи могут регистрироваться и входить в систему, получая JWT-токен в cookie.
- **Управление заказами**: Загрузка номеров заказов с проверкой по алгоритму Луна и получение списка заказов.
- **Управление балансом**: Просмотр текущего баланса, списание баллов и просмотр истории списаний.
- **Интеграция с системой начислений**: Фоновая обработка заказов с запросами к внешнему сервису для получения начислений.

## Эндпоинты API

| Метод | Эндпоинт                     | Описание                              | Аутентификация |
|-------|------------------------------|---------------------------------------|----------------|
| POST  | `/api/user/register`         | Регистрация пользователя              | Не требуется    |
| POST  | `/api/user/login`            | Аутентификация пользователя           | Не требуется    |
| POST  | `/api/user/orders`           | Загрузка номера заказа                | Требуется      |
| GET   | `/api/user/orders`           | Получение списка заказов              | Требуется      |
| GET   | `/api/user/balance`          | Получение текущего баланса            | Требуется      |
| POST  | `/api/user/balance/withdraw` | Списание баллов                       | Требуется      |
| GET   | `/api/user/withdrawals`      | Получение истории списаний            | Требуется      |

### Примечания
- Для аутентификации используется JWT-токен, передаваемый в cookie `token` или заголовке `Authorization: Bearer <token>`.
- Номер заказа должен соответствовать алгоритму Луна.
- Формат тела запросов и ответов — JSON, за исключением `POST /api/user/orders` (текст).

## Конфигурация

- **RUN_ADDRESS**: Адрес и порт сервера (по умолчанию `localhost:8080`).
- **DATABASE_URI**: Строка подключения к PostgreSQL.
- **ACCRUAL_SYSTEM_ADDRESS**: Адрес системы начислений (по умолчанию `http://localhost:8000`).

## Тестирование

Для запуска юнит-тестов выполните:
```bash
go test -v ./internal/...
```

## Зависимости

- `github.com/gorilla/mux` — маршрутизация HTTP-запросов.
- `github.com/joho/godotenv` — загрузка переменных окружения.
- `github.com/lib/pq` — драйвер PostgreSQL.
- `go.uber.org/zap` — логирование.
- `github.com/dgrijalva/jwt-go` — работа с JWT.
- `golang.org/x/crypto/bcrypt` — хеширование паролей.
- `golang.org/x/sync/errgroup` — управление горутинами.