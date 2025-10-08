package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gitea-jenkins-webhook/internal/config"
	"gitea-jenkins-webhook/internal/models"
)

// JenkinsService предоставляет методы для работы с Jenkins API
type JenkinsService struct {
	config config.JenkinsConfig
	client *http.Client
}

// NewJenkinsService создает новый сервис Jenkins
func NewJenkinsService(cfg config.JenkinsConfig) *JenkinsService {
	return &JenkinsService{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CheckJobExists проверяет существование джобы в Jenkins
func (s *JenkinsService) CheckJobExists(jobName string) (bool, error) {
	url := fmt.Sprintf("%s/job/%s/api/json", s.config.URL, jobName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(s.config.Username, s.config.Token)

	resp, err := s.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// 404 означает, что джоба не существует
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	// 200 означает, что джоба существует
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	// Другие коды ошибок
	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Errorf("jenkins API error: %d - %s", resp.StatusCode, string(body))
}

// GetJobInfo получает информацию о джобе
func (s *JenkinsService) GetJobInfo(jobName string) (*models.JenkinsJob, error) {
	url := fmt.Sprintf("%s/job/%s/api/json", s.config.URL, jobName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(s.config.Username, s.config.Token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jenkins API error: %d - %s", resp.StatusCode, string(body))
	}

	var job models.JenkinsJob
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &job, nil
}

// GetJobURL возвращает URL джобы в Jenkins
func (s *JenkinsService) GetJobURL(jobName string) string {
	return fmt.Sprintf("%s/job/%s", s.config.URL, jobName)
}