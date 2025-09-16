package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the devbox configuration
type Config struct {
	Projects map[string]*Project `json:"projects"`
}

// Project represents a devbox project
type Project struct {
	Name          string `json:"name"`
	ContainerName string `json:"container_name"`
	BaseImage     string `json:"base_image"`
	WorkspacePath string `json:"workspace_path"`
	Status        string `json:"status,omitempty"`
}

// ConfigManager handles configuration file operations
type ConfigManager struct {
	configPath string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() (*ConfigManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".devbox")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	return &ConfigManager{configPath: configPath}, nil
}

// Load loads the configuration from file
func (cm *ConfigManager) Load() (*Config, error) {
	config := &Config{
		Projects: make(map[string]*Project),
	}

	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// Config file doesn't exist, return empty config
		return config, nil
	}

	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if len(data) == 0 {
		// Empty file, return empty config
		return config, nil
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// Save saves the configuration to file
func (cm *ConfigManager) Save(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddProject adds a new project to the configuration
func (config *Config) AddProject(project *Project) {
	if config.Projects == nil {
		config.Projects = make(map[string]*Project)
	}
	config.Projects[project.Name] = project
}

// RemoveProject removes a project from the configuration
func (config *Config) RemoveProject(name string) {
	if config.Projects != nil {
		delete(config.Projects, name)
	}
}

// GetProject returns a project by name
func (config *Config) GetProject(name string) (*Project, bool) {
	if config.Projects == nil {
		return nil, false
	}
	project, exists := config.Projects[name]
	return project, exists
}

// GetProjects returns all projects
func (config *Config) GetProjects() map[string]*Project {
	if config.Projects == nil {
		return make(map[string]*Project)
	}
	return config.Projects
}
