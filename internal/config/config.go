package config

import (
	"encoding/json"
	"myenv/internal/utils"
	"os"
	"path/filepath"
)

type Config struct {
	Lang             string `json:"lang"`
	Version          string `json:"version"`
	ContainerRuntime string `json:"containerRuntime"`
}

func GetConfig(version string) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		panic(err)
	}

	targetPath := filepath.Join(homeDir, ".config", "myenv")

	if !utils.DirIsExists(targetPath) {
		if err := os.Mkdir(targetPath, 0755); err != nil {
			panic(err)
		}
	}

	envFilePath := filepath.Join(targetPath, "config.json")

	if _, err := os.Stat(envFilePath); err == nil {
		return
	}

	if _, err := os.Create(envFilePath); err != nil {
		panic(err)
	}

	defaultConfig := &Config{
		Lang:             "en",
		Version:          version,
		ContainerRuntime: "docker",
	}

	saveConfig(envFilePath, defaultConfig)
}

func saveConfig(configFilePath string, config *Config) {
	data, err := json.MarshalIndent(config, "", "  ")

	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(configFilePath, data, 0644); err != nil {
		panic(err)
	}
}

func CheckConfig() error {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return err
	}

	targetPath := filepath.Join(homeDir, ".config", "myenv")

	envFilePath := filepath.Join(targetPath, "config.json")

	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		return err
	}

	return nil
}
