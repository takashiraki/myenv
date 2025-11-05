package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var (
	createConfigFlag bool
)

func TestMain(m *testing.M) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	configPath := filepath.Join(homeDir, ".config", "myenv")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.MkdirAll(configPath, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
			os.Exit(1)
		}
	}

	configFilePath := filepath.Join(configPath, "config.json")

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		createConfigFlag = true

		defaultConfig := &Config{
			Lang:             "en",
			Version:          "test-version",
			ContainerRuntime: "docker",
		}

		saveConfig(configFilePath, defaultConfig)
	}

	containerName := "myenv_test_container"
	containerProxy := "myapp.localhost"
	path := "./testdata/docker-compose"
	lang := "php"
	fw := "none"
	options := map[string]string{
		"type": "clone",
	}

	AddProjectConfig(containerName, containerProxy, path, lang, fw, options)

	exitCode := m.Run()

	if createConfigFlag {
		if err := os.Remove(configFilePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing config file: %v\n", err)
			os.Exit(1)
		}
	}

	os.Exit(exitCode)
}

func Test_DeleteProjectConfig(t *testing.T) {
	projectName := "myenv_test_container"

	beforConfig, err := LoadConfig()

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if _, exists := beforConfig.Projects[projectName]; !exists {
		t.Fatalf("Project %s does not exist in config before deletion", projectName)
	}

	err = DeleteProjectConfig(projectName)

	if err != nil {
		t.Fatalf("Failed to delete project config: %v", err)
	}

	afterConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if _, exists := afterConfig.Projects[projectName]; exists {
		t.Fatalf("Project %s still exists in config after deletion", projectName)
	}
}
