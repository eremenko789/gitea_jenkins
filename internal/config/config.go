package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config представляет основную конфигурацию приложения
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Jenkins  JenkinsConfig  `yaml:"jenkins"`
	Gitea    GiteaConfig    `yaml:"gitea"`
	Jobs     []JobConfig    `yaml:"jobs"`
	Timeout  time.Duration  `yaml:"timeout"`
}

// ServerConfig конфигурация сервера
type ServerConfig struct {
	Port string `yaml:"port"`
}

// JenkinsConfig конфигурация Jenkins
type JenkinsConfig struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
}

// GiteaConfig конфигурация Gitea
type GiteaConfig struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

// JobConfig конфигурация джобы
type JobConfig struct {
	Name        string            `yaml:"name"`
	Repository  string            `yaml:"repository"`
	Branch      string            `yaml:"branch"`
	JenkinsJob  string            `yaml:"jenkins_job"`
	Parameters  map[string]string `yaml:"parameters"`
	Description string            `yaml:"description"`
}

// LoadConfig загружает конфигурацию из YAML файла
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Устанавливаем значения по умолчанию
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Minute
	}

	if config.Server.Port == "" {
		config.Server.Port = "8080"
	}

	return &config, nil
}