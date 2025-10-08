package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config представляет основную конфигурацию приложения
type Config struct {
	Server        ServerConfig        `yaml:"server"`
	Jenkins       JenkinsConfig       `yaml:"jenkins"`
	Gitea         GiteaConfig         `yaml:"gitea"`
	Organizations []OrganizationConfig `yaml:"organizations"`
	CheckTimeout  time.Duration       `yaml:"check_timeout"`
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

// OrganizationConfig конфигурация организации Jenkins
type OrganizationConfig struct {
	Name         string   `yaml:"name"`
	Repositories []string `yaml:"repositories"`
	JobPattern   string   `yaml:"job_pattern"`
	Description  string   `yaml:"description"`
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
	if config.CheckTimeout == 0 {
		config.CheckTimeout = 30 * time.Second
	}

	if config.Server.Port == "" {
		config.Server.Port = "8080"
	}

	return &config, nil
}