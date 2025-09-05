# API Спецификации

В этой директории размещаются OpenAPI (Swagger) спецификации для генерации кода API.

## Использование

1. Создайте файл `openapi.yaml` с описанием вашего API
2. Запустите `make generate` для генерации кода
3. Сгенерированные файлы появятся в:
   - `internal/generated/api/` - модели данных и интерфейсы
   - `internal/api/` - Chi handlers

## Пример структуры openapi.yaml

```yaml
openapi: 3.0.3
info:
  title: My Service API
  version: 1.0.0

paths:
  /users:
    get:
      operationId: getUsers
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UsersResponse'

components:
  schemas:
    User:
      type: object
      required:
        - id
        - name
      properties:
        id:
          type: string
        name:
          type: string
```

## Валидация запросов

Сгенерированный код автоматически включает валидацию запросов на основе OpenAPI схемы. Примеры подключения валидации смотрите в `cmd/main.go`.

## Полезные команды

- `make generate` - генерация кода из OpenAPI
- `make clean` - очистка сгенерированных файлов
- `make lint` - проверка качества кода
