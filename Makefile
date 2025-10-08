# Makefile для gitea-jenkins-webhook

.PHONY: build run test clean docker-build docker-run docker-compose-up docker-compose-down

# Переменные
BINARY_NAME=webhook-server
DOCKER_IMAGE=gitea-jenkins-webhook
CONFIG_FILE=config.yaml

# Сборка приложения
build:
	go build -o $(BINARY_NAME) .

# Запуск приложения
run: build
	./$(BINARY_NAME)

# Запуск с конфигурацией
run-dev: build
	CONFIG_FILE=$(CONFIG_FILE) ./$(BINARY_NAME)

# Тестирование
test:
	go test -v ./...

# Очистка
clean:
	go clean
	rm -f $(BINARY_NAME)

# Сборка Docker образа
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Запуск Docker контейнера
docker-run: docker-build
	docker run -p 8080:8080 -v $(PWD)/$(CONFIG_FILE):/app/$(CONFIG_FILE):ro $(DOCKER_IMAGE)

# Запуск через Docker Compose
docker-compose-up:
	docker-compose up -d

# Остановка Docker Compose
docker-compose-down:
	docker-compose down

# Просмотр логов
logs:
	docker-compose logs -f webhook-server

# Перезапуск сервиса
restart:
	docker-compose restart webhook-server

# Проверка статуса
status:
	docker-compose ps

# Форматирование кода
fmt:
	go fmt ./...

# Проверка линтера
lint:
	golangci-lint run

# Установка зависимостей для разработки
install-dev:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  build              - Сборка приложения"
	@echo "  run                - Запуск приложения"
	@echo "  run-dev            - Запуск с конфигурацией"
	@echo "  test               - Запуск тестов"
	@echo "  clean              - Очистка"
	@echo "  docker-build       - Сборка Docker образа"
	@echo "  docker-run         - Запуск Docker контейнера"
	@echo "  docker-compose-up  - Запуск через Docker Compose"
	@echo "  docker-compose-down - Остановка Docker Compose"
	@echo "  logs               - Просмотр логов"
	@echo "  restart            - Перезапуск сервиса"
	@echo "  status             - Проверка статуса"
	@echo "  fmt                - Форматирование кода"
	@echo "  lint               - Проверка линтера"
	@echo "  install-dev        - Установка зависимостей для разработки"
	@echo "  help               - Показать эту справку"