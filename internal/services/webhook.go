package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gitea-jenkins-webhook/internal/config"
	"gitea-jenkins-webhook/internal/models"
)

// WebhookService координирует обработку webhook
type WebhookService struct {
	jenkinsService *JenkinsService
	giteaService   *GiteaService
	config         *config.Config
}

// NewWebhookService создает новый сервис webhook
func NewWebhookService(jenkinsService *JenkinsService, giteaService *GiteaService, cfg *config.Config) *WebhookService {
	return &WebhookService{
		jenkinsService: jenkinsService,
		giteaService:   giteaService,
		config:         cfg,
	}
}

// ProcessPullRequest обрабатывает создание Pull Request
func (s *WebhookService) ProcessPullRequest(event models.PullRequestEvent) error {
	pr := event.PullRequest
	repo := event.Repository

	log.Printf("Processing PR #%d in repository %s", pr.Number, repo.FullName)

	// Находим подходящие джобы для этого репозитория и ветки
	matchingJobs := s.findMatchingJobs(repo.FullName, pr.Head.Ref)
	if len(matchingJobs) == 0 {
		log.Printf("No matching jobs found for repository %s and branch %s", repo.FullName, pr.Head.Ref)
		return nil
	}

	log.Printf("Found %d matching jobs", len(matchingJobs))

	// Запускаем джобы асинхронно
	jobResults := make(chan JobResult, len(matchingJobs))
	
	for _, job := range matchingJobs {
		go s.runJobAsync(job, pr, repo, jobResults)
	}

	// Собираем результаты
	var successfulJobs []JobResult
	var failedJobs []JobResult
	timeout := time.After(s.config.Timeout)

	for i := 0; i < len(matchingJobs); i++ {
		select {
		case result := <-jobResults:
			if result.Error != nil {
				failedJobs = append(failedJobs, result)
				log.Printf("Job %s failed: %v", result.JobName, result.Error)
			} else {
				successfulJobs = append(successfulJobs, result)
				log.Printf("Job %s completed successfully", result.JobName)
			}
		case <-timeout:
			log.Printf("Timeout waiting for jobs to complete")
			// Обрабатываем оставшиеся джобы как таймаут
			for j := i; j < len(matchingJobs); j++ {
				select {
				case result := <-jobResults:
					if result.Error != nil {
						failedJobs = append(failedJobs, result)
					} else {
						successfulJobs = append(successfulJobs, result)
					}
				default:
					// Джоба не завершилась в срок
					failedJobs = append(failedJobs, JobResult{
						JobName: matchingJobs[j].Name,
						Error:   fmt.Errorf("timeout"),
					})
				}
			}
			goto done
		}
	}

done:
	// Публикуем комментарий с результатами
	return s.publishResults(pr, repo, successfulJobs, failedJobs)
}

// JobResult представляет результат выполнения джобы
type JobResult struct {
	JobName   string
	BuildURL  string
	BuildID   int
	Error     error
}

// findMatchingJobs находит джобы, подходящие для репозитория и ветки
func (s *WebhookService) findMatchingJobs(repoFullName, branch string) []config.JobConfig {
	var matchingJobs []config.JobConfig

	for _, job := range s.config.Jobs {
		if job.Repository == repoFullName && (job.Branch == branch || job.Branch == "*") {
			matchingJobs = append(matchingJobs, job)
		}
	}

	return matchingJobs
}

// runJobAsync запускает джобу асинхронно
func (s *WebhookService) runJobAsync(job config.JobConfig, pr models.PullRequest, repo models.Repository, results chan<- JobResult) {
	log.Printf("Starting job %s", job.Name)

	// Подготавливаем параметры для джобы
	parameters := make(map[string]string)
	for key, value := range job.Parameters {
		parameters[key] = value
	}

	// Добавляем специфичные для PR параметры
	parameters["PR_NUMBER"] = fmt.Sprintf("%d", pr.Number)
	parameters["PR_TITLE"] = pr.Title
	parameters["PR_URL"] = pr.URL
	parameters["PR_HEAD_SHA"] = pr.Head.SHA
	parameters["PR_HEAD_REF"] = pr.Head.Ref
	parameters["PR_BASE_REF"] = pr.Base.Ref

	// Запускаем джобу
	buildResp, err := s.jenkinsService.TriggerBuild(job.JenkinsJob, parameters)
	if err != nil {
		results <- JobResult{
			JobName: job.Name,
			Error:   fmt.Errorf("failed to trigger build: %w", err),
		}
		return
	}

	log.Printf("Job %s triggered with queue ID %d", job.Name, buildResp.QueueID)

	// Ждем завершения сборки
	build, err := s.jenkinsService.WaitForBuildCompletion(job.JenkinsJob, buildResp.QueueID, s.config.Timeout)
	if err != nil {
		results <- JobResult{
			JobName: job.Name,
			Error:   fmt.Errorf("failed to wait for build completion: %w", err),
		}
		return
	}

	results <- JobResult{
		JobName:  job.Name,
		BuildURL: build.URL,
		BuildID:  build.Number,
	}
}

// publishResults публикует комментарий с результатами выполнения джоб
func (s *WebhookService) publishResults(pr models.PullRequest, repo models.Repository, successfulJobs, failedJobs []JobResult) error {
	owner := repo.Owner.Login
	repoName := repo.Name

	var comment strings.Builder
	comment.WriteString("## 🚀 Jenkins Jobs Status\n\n")

	if len(successfulJobs) > 0 {
		comment.WriteString("### ✅ Successful Jobs\n")
		for _, job := range successfulJobs {
			comment.WriteString(fmt.Sprintf("- **%s**: [Build #%d](%s)\n", job.JobName, job.BuildID, job.BuildURL))
		}
		comment.WriteString("\n")
	}

	if len(failedJobs) > 0 {
		comment.WriteString("### ❌ Failed Jobs\n")
		for _, job := range failedJobs {
			if job.Error != nil {
				comment.WriteString(fmt.Sprintf("- **%s**: %s\n", job.JobName, job.Error.Error()))
			} else {
				comment.WriteString(fmt.Sprintf("- **%s**: Unknown error\n", job.JobName))
			}
		}
		comment.WriteString("\n")
	}

	if len(successfulJobs) == 0 && len(failedJobs) == 0 {
		comment.WriteString("No jobs were triggered for this PR.\n")
	}

	comment.WriteString(fmt.Sprintf("*Processed at %s*", time.Now().Format("2006-01-02 15:04:05 UTC")))

	return s.giteaService.CreateComment(owner, repoName, pr.Number, comment.String())
}