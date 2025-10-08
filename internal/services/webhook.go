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

	// Находим подходящие организации для этого репозитория
	matchingOrgs := s.findMatchingOrganizations(repo.Name)
	if len(matchingOrgs) == 0 {
		log.Printf("No matching organizations found for repository %s", repo.Name)
		return nil
	}

	log.Printf("Found %d matching organizations", len(matchingOrgs))

	// Проверяем существование джоб асинхронно
	jobResults := make(chan JobResult, len(matchingOrgs))
	
	for _, org := range matchingOrgs {
		go s.checkJobAsync(org, pr, repo, jobResults)
	}

	// Собираем результаты
	var existingJobs []JobResult
	var missingJobs []JobResult
	timeout := time.After(s.config.CheckTimeout)

	for i := 0; i < len(matchingOrgs); i++ {
		select {
		case result := <-jobResults:
			if result.Error != nil {
				missingJobs = append(missingJobs, result)
				log.Printf("Job %s check failed: %v", result.JobName, result.Error)
			} else if result.Exists {
				existingJobs = append(existingJobs, result)
				log.Printf("Job %s exists", result.JobName)
			} else {
				missingJobs = append(missingJobs, result)
				log.Printf("Job %s does not exist", result.JobName)
			}
		case <-timeout:
			log.Printf("Timeout waiting for job checks to complete")
			// Обрабатываем оставшиеся джобы как таймаут
			for j := i; j < len(matchingOrgs); j++ {
				select {
				case result := <-jobResults:
					if result.Error != nil {
						missingJobs = append(missingJobs, result)
					} else if result.Exists {
						existingJobs = append(existingJobs, result)
					} else {
						missingJobs = append(missingJobs, result)
					}
				default:
					// Проверка не завершилась в срок
					missingJobs = append(missingJobs, JobResult{
						JobName: fmt.Sprintf("%s/%s/PR-%d", matchingOrgs[j].Name, repo.Name, pr.Number),
						Error:   fmt.Errorf("timeout"),
					})
				}
			}
			goto done
		}
	}

done:
	// Публикуем комментарий с результатами
	return s.publishResults(pr, repo, existingJobs, missingJobs)
}

// JobResult представляет результат проверки джобы
type JobResult struct {
	JobName  string
	JobURL   string
	Exists   bool
	Error    error
}

// findMatchingOrganizations находит организации, отслеживающие репозиторий
func (s *WebhookService) findMatchingOrganizations(repoName string) []config.OrganizationConfig {
	var matchingOrgs []config.OrganizationConfig

	for _, org := range s.config.Organizations {
		for _, repo := range org.Repositories {
			if repo == repoName {
				matchingOrgs = append(matchingOrgs, org)
				break
			}
		}
	}

	return matchingOrgs
}

// checkJobAsync проверяет существование джобы асинхронно
func (s *WebhookService) checkJobAsync(org config.OrganizationConfig, pr models.PullRequest, repo models.Repository, results chan<- JobResult) {
	// Формируем имя джобы по шаблону
	jobName := s.buildJobName(org.JobPattern, org.Name, repo.Name, pr.Number)
	
	log.Printf("Checking job %s", jobName)

	// Проверяем существование джобы
	exists, err := s.jenkinsService.CheckJobExists(jobName)
	if err != nil {
		results <- JobResult{
			JobName: jobName,
			Error:   fmt.Errorf("failed to check job existence: %w", err),
		}
		return
	}

	jobURL := s.jenkinsService.GetJobURL(jobName)
	results <- JobResult{
		JobName: jobName,
		JobURL:  jobURL,
		Exists:  exists,
	}
}

// buildJobName строит имя джобы по шаблону
func (s *WebhookService) buildJobName(pattern, organization, repository string, prNumber int) string {
	jobName := pattern
	jobName = strings.ReplaceAll(jobName, "{organization}", organization)
	jobName = strings.ReplaceAll(jobName, "{repository}", repository)
	jobName = strings.ReplaceAll(jobName, "{pr_number}", fmt.Sprintf("%d", prNumber))
	return jobName
}

// publishResults публикует комментарий с результатами проверки джоб
func (s *WebhookService) publishResults(pr models.PullRequest, repo models.Repository, existingJobs, missingJobs []JobResult) error {
	owner := repo.Owner.Login
	repoName := repo.Name

	var comment strings.Builder
	comment.WriteString("## 🔍 Jenkins Jobs Check\n\n")

	if len(existingJobs) > 0 {
		comment.WriteString("### ✅ Existing Jobs\n")
		for _, job := range existingJobs {
			comment.WriteString(fmt.Sprintf("- **%s**: [View Job](%s)\n", job.JobName, job.JobURL))
		}
		comment.WriteString("\n")
	}

	if len(missingJobs) > 0 {
		comment.WriteString("### ❌ Missing Jobs\n")
		for _, job := range missingJobs {
			if job.Error != nil {
				comment.WriteString(fmt.Sprintf("- **%s**: %s\n", job.JobName, job.Error.Error()))
			} else {
				comment.WriteString(fmt.Sprintf("- **%s**: Job not found\n", job.JobName))
			}
		}
		comment.WriteString("\n")
	}

	if len(existingJobs) == 0 && len(missingJobs) == 0 {
		comment.WriteString("No organizations are tracking this repository.\n")
	}

	comment.WriteString(fmt.Sprintf("*Checked at %s*", time.Now().Format("2006-01-02 15:04:05 UTC")))

	return s.giteaService.CreateComment(owner, repoName, pr.Number, comment.String())
}