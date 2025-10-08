package models

// JenkinsJob представляет джобу в Jenkins
type JenkinsJob struct {
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Buildable   bool              `json:"buildable"`
	Color       string            `json:"color"`
	Description string            `json:"description"`
	LastBuild   *JenkinsBuild     `json:"lastBuild"`
	LastCompletedBuild *JenkinsBuild `json:"lastCompletedBuild"`
	LastSuccessfulBuild *JenkinsBuild `json:"lastSuccessfulBuild"`
}

// JenkinsBuild представляет сборку в Jenkins
type JenkinsBuild struct {
	Number     int    `json:"number"`
	URL        string `json:"url"`
	Result     string `json:"result"`
	Building   bool   `json:"building"`
	Duration   int    `json:"duration"`
	Timestamp  int64  `json:"timestamp"`
	Description string `json:"description"`
}

// BuildRequest представляет запрос на запуск сборки
type BuildRequest struct {
	Parameter []BuildParameter `json:"parameter"`
}

// BuildParameter представляет параметр сборки
type BuildParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// BuildResponse представляет ответ на запрос запуска сборки
type BuildResponse struct {
	QueueID int `json:"queueId"`
}

// QueueItem представляет элемент в очереди Jenkins
type QueueItem struct {
	ID       int    `json:"id"`
	Task     Task   `json:"task"`
	Why      string `json:"why"`
	Blocked  bool   `json:"blocked"`
	Buildable bool  `json:"buildable"`
	InQueueSince int64 `json:"inQueueSince"`
	Stuck    bool   `json:"stuck"`
}

// Task представляет задачу в очереди
type Task struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}