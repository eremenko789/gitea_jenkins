package models

// JenkinsJob представляет джобу в Jenkins
type JenkinsJob struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Buildable   bool   `json:"buildable"`
	Color       string `json:"color"`
	Description string `json:"description"`
}