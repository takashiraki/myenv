package application

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type (
	Config struct {
		lang             string
		containerRuntime string
	}

	ConfigService struct {
		path string
	}
)

func NewConfigService() (*ConfigService, error) {
	homedir, err := os.UserHomeDir()

	if err != nil {
		return nil, err
	}

	path := filepath.Join(homedir, ".config", "myenv", "config.json")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	return &ConfigService{
		path: path,
	}, nil
}

func (s *ConfigService) GetConfig() (Config, error) {
	data, err := os.ReadFile(s.path)

	if err != nil {
		return Config{}, err
	}

	var config Config

	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}
