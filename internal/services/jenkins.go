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

// TriggerBuild запускает сборку в Jenkins
func (s *JenkinsService) TriggerBuild(jobName string, parameters map[string]string) (*models.BuildResponse, error) {
	url := fmt.Sprintf("%s/job/%s/buildWithParameters", s.config.URL, jobName)
	
	// Создаем параметры для сборки
	var buildParams []models.BuildParameter
	for key, value := range parameters {
		buildParams = append(buildParams, models.BuildParameter{
			Name:  key,
			Value: value,
		})
	}

	buildRequest := models.BuildRequest{
		Parameter: buildParams,
	}

	jsonData, err := json.Marshal(buildRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal build request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.config.Username, s.config.Token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jenkins API error: %d - %s", resp.StatusCode, string(body))
	}

	// Jenkins возвращает 201 Created без тела ответа при успешном запуске
	// Мы можем получить queue ID из заголовка Location
	location := resp.Header.Get("Location")
	if location != "" {
		// Извлекаем queue ID из URL
		// Формат: /queue/item/12345/
		var queueID int
		if _, err := fmt.Sscanf(location, "/queue/item/%d/", &queueID); err == nil {
			return &models.BuildResponse{QueueID: queueID}, nil
		}
	}

	return &models.BuildResponse{}, nil
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

// GetBuildInfo получает информацию о сборке
func (s *JenkinsService) GetBuildInfo(jobName string, buildNumber int) (*models.JenkinsBuild, error) {
	url := fmt.Sprintf("%s/job/%s/%d/api/json", s.config.URL, jobName, buildNumber)

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

	var build models.JenkinsBuild
	if err := json.NewDecoder(resp.Body).Decode(&build); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &build, nil
}

// WaitForBuildCompletion ждет завершения сборки
func (s *JenkinsService) WaitForBuildCompletion(jobName string, queueID int, timeout time.Duration) (*models.JenkinsBuild, error) {
	start := time.Now()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Проверяем очередь
			queueItem, err := s.getQueueItem(queueID)
			if err != nil {
				return nil, fmt.Errorf("failed to get queue item: %w", err)
			}

			if queueItem == nil {
				// Элемент больше не в очереди, получаем последнюю сборку
				job, err := s.GetJobInfo(jobName)
				if err != nil {
					return nil, fmt.Errorf("failed to get job info: %w", err)
				}

				if job.LastBuild != nil {
					build, err := s.GetBuildInfo(jobName, job.LastBuild.Number)
					if err != nil {
						return nil, fmt.Errorf("failed to get build info: %w", err)
					}
					return build, nil
				}
			}

			if time.Since(start) > timeout {
				return nil, fmt.Errorf("timeout waiting for build completion")
			}
		}
	}
}

// getQueueItem получает информацию об элементе очереди
func (s *JenkinsService) getQueueItem(queueID int) (*models.QueueItem, error) {
	url := fmt.Sprintf("%s/queue/item/%d/api/json", s.config.URL, queueID)

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

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Элемент больше не в очереди
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jenkins API error: %d - %s", resp.StatusCode, string(body))
	}

	var item models.QueueItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &item, nil
}