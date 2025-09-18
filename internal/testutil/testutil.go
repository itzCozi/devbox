package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"devbox/internal/config"
)

func CreateTempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

func CreateTestConfig() *config.Config {
	return &config.Config{
		Projects: make(map[string]*config.Project),
		Settings: &config.GlobalSettings{
			DefaultBaseImage:    "ubuntu:22.04",
			DefaultEnvironment:  map[string]string{"PATH": "/usr/local/bin:/usr/bin:/bin"},
			ConfigTemplatesPath: "",
			AutoUpdate:          true,
			AutoStopOnExit:      true,
		},
	}
}

func CreateTestProject(name string) *config.Project {
	return &config.Project{
		Name:          name,
		BoxName:       name + "-box",
		BaseImage:     "ubuntu:22.04",
		WorkspacePath: filepath.Join("/home/user/devbox", name),
		Status:        "stopped",
		ConfigFile:    filepath.Join("/home/user/devbox", name, "devbox.json"),
	}
}

func CreateTestProjectConfig(name string) *config.ProjectConfig {
	return &config.ProjectConfig{
		Name:      name,
		BaseImage: "ubuntu:22.04",
		SetupCommands: []string{
			"apt update",
			"apt install -y curl git",
		},
		Environment: map[string]string{
			"ENV":  "development",
			"USER": "developer",
		},
		Ports: []string{
			"8080:8080",
			"3000:3000",
		},
		Volumes: []string{
			"/workspace/data:/data",
		},
		WorkingDir: "/workspace",
		Shell:      "/bin/bash",
		User:       "root",
	}
}

func WriteJSONFile(t *testing.T, path string, data interface{}) {
	t.Helper()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	err = os.WriteFile(path, jsonData, 0644)
	if err != nil {
		t.Fatalf("Failed to write JSON file: %v", err)
	}
}

func ReadJSONFile(t *testing.T, path string, dest interface{}) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}

	err = json.Unmarshal(data, dest)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
}

func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func AssertError(t *testing.T, err error, expectedMessage string) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected error but got none")
	}

	if expectedMessage != "" && !Contains(err.Error(), expectedMessage) {
		t.Fatalf("Expected error to contain %q, got %q", expectedMessage, err.Error())
	}
}

func AssertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

func AssertNotNil(t *testing.T, value interface{}) {
	t.Helper()
	if value == nil {
		t.Fatal("Expected value to not be nil")
	}
}

func AssertNil(t *testing.T, value interface{}) {
	t.Helper()
	if value != nil {
		t.Fatalf("Expected value to be nil, got %v", value)
	}
}

func Contains(s, substr string) bool {
	return len(s) >= len(substr) && containsAtIndex(s, substr)
}

func containsAtIndex(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func CreateConfigManager(t *testing.T) (*config.ConfigManager, string) {
	t.Helper()

	tempDir := CreateTempDir(t)

	cm, err := config.NewConfigManager()
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	return cm, tempDir
}

func ValidateProjectName(name string) bool {
	if name == "" {
		return false
	}

	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}

	return true
}
