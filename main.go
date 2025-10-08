package main

import (
	"log"
	"os"

	"gitea-jenkins-webhook/internal/config"
	"gitea-jenkins-webhook/internal/handlers"
	"gitea-jenkins-webhook/internal/services"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создаем сервисы
	jenkinsService := services.NewJenkinsService(cfg.Jenkins)
	giteaService := services.NewGiteaService(cfg.Gitea)
	webhookService := services.NewWebhookService(jenkinsService, giteaService, cfg)

	// Создаем обработчик HTTP запросов
	handler := handlers.NewWebhookHandler(webhookService)

	// Запускаем сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := handler.Start(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}