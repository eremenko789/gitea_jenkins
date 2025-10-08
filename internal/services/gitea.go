package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gitea-jenkins-webhook/internal/config"
	"gitea-jenkins-webhook/internal/models"
)

// GiteaService предоставляет методы для работы с Gitea API
type GiteaService struct {
	config config.GiteaConfig
	client *http.Client
}

// NewGiteaService создает новый сервис Gitea
func NewGiteaService(cfg config.GiteaConfig) *GiteaService {
	return &GiteaService{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateComment создает комментарий к Pull Request
func (s *GiteaService) CreateComment(owner, repo string, prNumber int, body string) error {
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/issues/%d/comments", 
		s.config.URL, owner, repo, prNumber)

	comment := models.CreateCommentRequest{
		Body: body,
	}

	jsonData, err := json.Marshal(comment)
	if err != nil {
		return fmt.Errorf("failed to marshal comment: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+s.config.Token)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gitea API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetPullRequest получает информацию о Pull Request
func (s *GiteaService) GetPullRequest(owner, repo string, prNumber int) (*models.PullRequest, error) {
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/pulls/%d", 
		s.config.URL, owner, repo, prNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+s.config.Token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gitea API error: %d - %s", resp.StatusCode, string(body))
	}

	var pr models.PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &pr, nil
}