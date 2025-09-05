# Service Skeleton

Скелетон для создания Go микросервисов в рамках платформы GenDO.

## Особенности

- ✅ Структурированное логирование с Logrus
- ✅ Трассировка запросов с OpenTelemetry + Jaeger
- ✅ Метрики Prometheus
- ✅ HTTP сервер на Chi с middleware
- ✅ Graceful shutdown
- ✅ Конфигурация через переменные окружения
- ✅ Docker контейнеризация
- ✅ Health check эндпоинты

## Структура проекта

```
.
├── cmd/
│   └── main.go              # Точка входа в приложение
├── internal/
│   ├── config/              # Конфигурация приложения
│   ├── handlers/            # HTTP обработчики
│   ├── models/              # Модели данных
│   └── services/            # Бизнес-логика
├── api/                     # OpenAPI спецификации
├── Dockerfile               # Docker образ
├── env.example              # Пример переменных окружения
├── go.mod                   # Go модули
└── README.md               # Документация

```

## Быстрый старт

1. Скопируйте файл конфигурации:
   ```bash
   cp env.example .env
   ```

2. Установите зависимости:
   ```bash
   go mod download
   ```

3. Запустите сервис:
   ```bash
   go run cmd/main.go
   ```

4. Проверьте работу:
   ```bash
   curl http://localhost:8080/health
   curl http://localhost:8080/api/v1/ping
   ```

## Конфигурация

Основные переменные окружения:

- `PORT` - порт сервера (по умолчанию: 8080)
- `SERVICE_NAME` - имя сервиса для трассировки
- `LOG_LEVEL` - уровень логирования (debug, info, warn, error)
- `JAEGER_ENDPOINT` - эндпоинт Jaeger для трассировки

## Эндпоинты

- `GET /health` - проверка состояния сервиса
- `GET /metrics` - метрики Prometheus
- `GET /api/v1/ping` - тестовый эндпоинт

## Docker

Сборка образа:
```bash
docker build -t service-skeleton .
```

Запуск контейнера:
```bash
docker run -p 8080:8080 --env-file .env service-skeleton
```

## Разработка

Для разработки нового сервиса на основе этого скелетона:

1. Скопируйте директорию скелетона
2. Обновите `go.mod` с правильным именем модуля
3. Обновите `SERVICE_NAME` в конфигурации
4. Добавьте свои обработчики в `internal/handlers/`
5. Реализуйте бизнес-логику в `internal/services/`
6. Опишите API в `api/openapi.yaml`
