# Пример использования Gitea-Jenkins Webhook Server

## Быстрый старт

### 1. Настройка конфигурации

Скопируйте `config.yaml` и отредактируйте под ваши нужды:

```bash
cp config.yaml my-config.yaml
# Отредактируйте my-config.yaml
```

### 2. Запуск локально

```bash
# Установка зависимостей
go mod download

# Запуск приложения
go run main.go
```

### 3. Запуск через Docker

```bash
# Сборка образа
docker build -t gitea-jenkins-webhook .

# Запуск контейнера
docker run -p 8080:8080 -v $(pwd)/config.yaml:/app/config.yaml:ro gitea-jenkins-webhook
```

### 4. Запуск через Docker Compose

```bash
# Запуск всех сервисов (включая Jenkins и Gitea для тестирования)
docker-compose up -d

# Просмотр логов
docker-compose logs -f webhook-server
```

## Настройка webhook в Gitea

1. Откройте ваш репозиторий в Gitea
2. Перейдите в Settings → Webhooks
3. Нажмите "Add Webhook"
4. Заполните форму:
   - **Payload URL**: `http://your-server:8080/webhook/gitea`
   - **Content Type**: `application/json`
   - **Secret**: (опционально)
   - **Which events would you like to trigger this webhook**: выберите "Pull Request"
5. Нажмите "Add Webhook"

## Тестирование

### 1. Проверка здоровья сервиса

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:
```json
{"status":"healthy"}
```

### 2. Тестирование webhook

Создайте тестовый webhook payload:

```bash
curl -X POST http://localhost:8080/webhook/gitea \
  -H "Content-Type: application/json" \
  -d '{
    "action": "opened",
    "number": 1,
    "pull_request": {
      "id": 1,
      "number": 1,
      "title": "Test PR",
      "body": "Test PR description",
      "state": "open",
      "html_url": "http://gitea:3000/test/repo/pulls/1",
      "head": {
        "ref": "main",
        "sha": "abc123",
        "repo": {
          "name": "test-repo",
          "full_name": "test/test-repo",
          "html_url": "http://gitea:3000/test/test-repo",
          "clone_url": "http://gitea:3000/test/test-repo.git",
          "owner": {
            "login": "test",
            "id": 1
          }
        }
      },
      "base": {
        "ref": "develop",
        "repo": {
          "name": "test-repo",
          "full_name": "test/test-repo",
          "owner": {
            "login": "test"
          }
        }
      },
      "user": {
        "login": "testuser",
        "id": 2
      }
    },
    "repository": {
      "id": 1,
      "name": "test-repo",
      "full_name": "test/test-repo",
      "html_url": "http://gitea:3000/test/test-repo",
      "clone_url": "http://gitea:3000/test/test-repo.git",
      "owner": {
        "login": "test",
        "id": 1
      }
    },
    "sender": {
      "login": "testuser",
      "id": 2
    }
  }'
```

## Мониторинг

### Логи приложения

Приложение выводит подробные логи:

```
2024/01/15 10:30:00 Processing PR #1: Test PR
2024/01/15 10:30:00 Found 2 matching jobs
2024/01/15 10:30:01 Starting job build-frontend
2024/01/15 10:30:01 Starting job build-backend
2024/01/15 10:30:05 Job build-frontend completed successfully
2024/01/15 10:30:08 Job build-backend completed successfully
```

### Мониторинг через Docker

```bash
# Просмотр логов
docker-compose logs -f webhook-server

# Статистика контейнеров
docker-compose ps

# Использование ресурсов
docker stats
```

## Устранение неполадок

### 1. Проблемы с подключением к Jenkins

Проверьте:
- URL Jenkins в конфигурации
- Токен аутентификации
- Доступность Jenkins из контейнера

```bash
# Тест подключения к Jenkins
curl -u username:token http://jenkins:8080/api/json
```

### 2. Проблемы с подключением к Gitea

Проверьте:
- URL Gitea в конфигурации
- Токен аутентификации
- Доступность Gitea из контейнера

```bash
# Тест подключения к Gitea
curl -H "Authorization: token YOUR_TOKEN" http://gitea:3000/api/v1/user
```

### 3. Проблемы с джобами

Проверьте:
- Имена джоб в Jenkins
- Параметры джоб
- Соответствие репозитория и ветки

### 4. Таймауты

Если джобы не завершаются в срок:
- Увеличьте значение `timeout` в конфигурации
- Проверьте производительность Jenkins
- Убедитесь, что джобы не зависают

## Производственное развертывание

### 1. Использование переменных окружения

```yaml
jenkins:
  url: "${JENKINS_URL}"
  username: "${JENKINS_USERNAME}"
  token: "${JENKINS_TOKEN}"

gitea:
  url: "${GITEA_URL}"
  token: "${GITEA_TOKEN}"
```

### 2. Использование Kubernetes

Создайте ConfigMap и Secret:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: webhook-config
data:
  config.yaml: |
    server:
      port: "8080"
    timeout: "15m"
    jobs:
      - name: "build"
        repository: "company/app"
        branch: "main"
        jenkins_job: "app-build"
        parameters:
          BRANCH_NAME: "main"
---
apiVersion: v1
kind: Secret
metadata:
  name: webhook-secrets
type: Opaque
stringData:
  jenkins-token: "your-jenkins-token"
  gitea-token: "your-gitea-token"
```

### 3. Мониторинг и алерты

Настройте мониторинг:
- HTTP health check на `/health`
- Логи ошибок
- Метрики времени выполнения джоб
- Алерты при недоступности сервисов