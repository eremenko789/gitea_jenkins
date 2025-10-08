# Gitea-Jenkins Webhook Server

Веб-сервер на Go для интеграции Gitea и Jenkins. Приложение получает webhook от Gitea о создании Pull Request, проверяет существование соответствующих джоб в Jenkins и публикует комментарии с результатами проверки.

## Возможности

- ✅ Обработка webhook от Gitea о создании Pull Request
- ✅ Проверка существования джоб в Jenkins по шаблону
- ✅ Публикация комментариев в Gitea с результатами проверки
- ✅ Обработка таймаутов
- ✅ Конфигурируемые организации и шаблоны джоб через YAML
- ✅ Docker контейнеризация
- ✅ Параллельная обработка webhook

## Структура проекта

```
.
├── main.go                    # Точка входа
├── go.mod                     # Go модули
├── config.yaml               # Конфигурация
├── Dockerfile                # Docker образ
├── docker-compose.yml        # Docker Compose
└── internal/
    ├── config/               # Конфигурация
    ├── handlers/             # HTTP обработчики
    ├── models/               # Модели данных
    └── services/             # Бизнес-логика
```

## Конфигурация

Отредактируйте `config.yaml`:

```yaml
server:
  port: "8080"

jenkins:
  url: "http://jenkins:8080"
  username: "admin"
  token: "your-jenkins-token"

gitea:
  url: "http://gitea:3000"
  token: "your-gitea-token"

# Организации Jenkins, которые отслеживают репозитории
organizations:
  - name: "redkitlab"
    repositories:
      - "scada"
      - "monitoring"
    job_pattern: "{organization}/{repository}/PR-{pr_number}"
    description: "RedKitLab organization jobs"
  
  - name: "devops"
    repositories:
      - "infrastructure"
      - "deployment"
    job_pattern: "{organization}/{repository}/PR-{pr_number}"
    description: "DevOps organization jobs"

# Таймаут для проверки существования джоб
check_timeout: "30s"
```

### Шаблоны джоб

Приложение использует шаблоны для формирования имен джоб:
- `{organization}` - имя организации
- `{repository}` - имя репозитория
- `{pr_number}` - номер Pull Request

Например, для PR #453 в репозитории "scada" организации "redkitlab" будет проверяться джоба: `redkitlab/scada/PR-453`

## Запуск

### Локальная разработка

1. Установите зависимости:
```bash
go mod download
```

2. Запустите приложение:
```bash
go run main.go
```

### Docker

1. Соберите образ:
```bash
docker build -t gitea-jenkins-webhook .
```

2. Запустите контейнер:
```bash
docker run -p 8080:8080 -v $(pwd)/config.yaml:/app/config.yaml gitea-jenkins-webhook
```

### Docker Compose

1. Запустите все сервисы:
```bash
docker-compose up -d
```

Это запустит:
- Webhook сервер на порту 8080
- Jenkins на порту 8081
- Gitea на порту 3000

## Настройка webhook в Gitea

1. Откройте настройки репозитория в Gitea
2. Перейдите в раздел "Webhooks"
3. Добавьте новый webhook:
   - URL: `http://your-server:8080/webhook/gitea`
   - Content Type: `application/json`
   - Events: выберите "Pull Request"

## API Endpoints

- `POST /webhook/gitea` - Обработка webhook от Gitea
- `GET /health` - Проверка состояния сервиса

## Переменные окружения

- `PORT` - Порт сервера (по умолчанию 8080)

## Логирование

Приложение выводит подробные логи о:
- Получении webhook
- Запуске джоб в Jenkins
- Результатах выполнения
- Ошибках

## Безопасность

- Приложение запускается под непривилегированным пользователем
- Использует HTTPS для API запросов
- Поддерживает аутентификацию через токены

## Разработка

### Добавление новых типов событий

1. Расширьте структуры в `internal/models/gitea.go`
2. Обновите обработчик в `internal/handlers/webhook.go`
3. Добавьте логику в `internal/services/webhook.go`

### Добавление новых параметров джоб

1. Обновите `JobConfig` в `internal/config/config.go`
2. Добавьте обработку в `internal/services/webhook.go`

## Лицензия

MIT