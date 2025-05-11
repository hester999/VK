# PubSub Service

Сервис публикации и подписки на сообщения с использованием gRPC.



## Структура проекта

```
.
├── api/
│   └── proto/           # Proto файлы
├── cmd/
│   ├── client/         # Клиентское приложение
│   └── server/         # Серверное приложение
├── config/
│   └── config.json     # Конфигурационный файл
├── internal/
│   ├── app/           # Основная логика приложения
│   ├── config/        # Конфигурация
│   ├── domain/        # Доменная логика
│   ├── entity/        # Сущности
│   ├── grpc/          # gRPC сервисы
│   └── pkg/           # Вспомогательные пакеты
├── Dockerfile.client   # Dockerfile для клиента
├── Dockerfile.server   # Dockerfile для сервера
└── docker-compose.yml  # Docker Compose конфигурация
```

## Сборка и запуск

### Локальная сборка

1. Сборка сервера:
```bash
go build -o server ./cmd/server
```

2. Сборка клиента:
```bash
go build -o client ./cmd/client
```

3. Запуск сервера:
```bash
./server
```

4. Запуск клиента:
```bash
./client
```

### Сборка в Docker

1. Сборка и запуск через Docker Compose:
```bash
docker-compose up --build
```

2. Остановка:
```bash
docker-compose down
```

## Конфигурация

Конфигурация хранится в `config/config.json`:

```json
{
    "server": {
        "addr": "0.0.0.0",
        "port": "50051"
    },
    "client": {
        "default_server": "server:50051",
        "default_topic": "test-topic",
        "topics": [
            "news",
            "weather",
            "sports",
            "tech",
            "music"
        ]
    }
}
```

### Варианты конфигурации в Docker

В `docker-compose.yml` доступны несколько вариантов конфигурации:

1. Через переменную окружения:
```yaml
environment:
  - CONFIG_PATH=/app/config/config.json
```

2. Монтирование конфига через volume:
```yaml
volumes:
  - ./config:/app/config
```

3. Через команду с флагом:
```yaml
command: ["./server", "-config", "/app/config/config.json"]
```

4. Монтирование отдельного конфига:
```yaml
volumes:
  - ./custom-config.json:/app/config/config.json
```

## Логирование

Сервис использует структурированное логирование:
- Сервер: формат `time=... level=INFO msg="..."`
- Клиент: формат `time=... level=INFO msg="..."`

## Тестирование

1. Запуск тестов:
```bash
go test ./...
``` 







# О проекте

Это задание было для меня новым и интересным опытом. До него я работал с REST и GraphQL, но с gRPC столкнулся впервые. Было немного непривычно, но в процессе стало понятнее, как устроены подписки, стримы.

 Раньше я делал только простые задачи на многопоточность и асинхронность, а здесь пришлось реально вникнуть в то, как всё устроено.

В целом, задание помогло мне прокачаться и почувствовать уверенность в теме, с которой раньше не приходилось работать.