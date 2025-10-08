package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitea-jenkins-webhook/internal/models"
	"gitea-jenkins-webhook/internal/services"
)

// WebhookHandler обрабатывает HTTP запросы
type WebhookHandler struct {
	webhookService *services.WebhookService
}

// NewWebhookHandler создает новый обработчик webhook
func NewWebhookHandler(webhookService *services.WebhookService) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
	}
}

// Start запускает HTTP сервер
func (h *WebhookHandler) Start(port string) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Middleware для логирования
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Маршруты
	r.POST("/webhook/gitea", h.handleGiteaWebhook)
	r.GET("/health", h.healthCheck)

	return r.Run(":" + port)
}

// handleGiteaWebhook обрабатывает webhook от Gitea
func (h *WebhookHandler) handleGiteaWebhook(c *gin.Context) {
	var event models.PullRequestEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		log.Printf("Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Обрабатываем только события создания PR
	if event.Action != "opened" {
		log.Printf("Ignoring event action: %s", event.Action)
		c.JSON(http.StatusOK, gin.H{"message": "Event ignored"})
		return
	}

	log.Printf("Processing PR #%d: %s", event.Number, event.PullRequest.Title)

	// Асинхронно обрабатываем webhook
	go func() {
		if err := h.webhookService.ProcessPullRequest(event); err != nil {
			log.Printf("Failed to process PR #%d: %v", event.Number, err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Webhook received and processing started"})
}

// healthCheck проверяет состояние сервиса
func (h *WebhookHandler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}