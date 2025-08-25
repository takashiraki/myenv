package config

import (
	"encoding/json"
	"myenv/internal/utils"
	"os"
	"path/filepath"
)

type Project struct {
	ContainerName string `json:"container_name"`
	ContainerPort int `json:"container_port"`
	Path string `json:"path"`
	Lang string `json:"lang"`
	Fw string `json:"framework"`
	Options []string `json:"options"`
}

type Config struct {
	Lang             string `json:"lang"`
	Version          string `json:"version"`
	ContainerRuntime string `json:"containerRuntime"`
	Projects map[string]Project `json:"projects"`
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

func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return nil, err
	}

	envFilePath := filepath.Join(homeDir, ".config","myenv","config.json")
	
	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		return nil, err
	}

	data, err := os.ReadFile(envFilePath)
	if err != nil {
		return nil, err
	}

	var config Config

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveConfig(config *Config)error {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		panic(err)
	}

	envFilePath := filepath.Join(homeDir, ".config","myenv","config.json")

	data, err := json.MarshalIndent(config, "", "  ")

	if err != nil {
		return err
	}

	if err := os.WriteFile(envFilePath, data, 0644); err != nil {
		return err
	}

	return nil
}

func AddProjectConfig(containerName string, containerPort int, path string, lang string, fw string, options []string)error {
	config, err := LoadConfig()

	if err != nil {
		return err
	}

	if config.Projects == nil {
		config.Projects = make(map[string]Project)
	}

	config.Projects[containerName] = Project{
		ContainerName: containerName,
		ContainerPort: containerPort,
		Path: path,
		Lang: lang,
		Fw: fw,
		Options: options,
	}

	if err := SaveConfig(config); err != nil {
		return err
	}

	return nil
}