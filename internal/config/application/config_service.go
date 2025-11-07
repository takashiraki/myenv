package application

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type (
	Config struct {
		Lang             string
		ContainerRuntime string
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

	return &ConfigService{
		path: path,
	}, nil
}

func (s *ConfigService) GetConfig() (Config, error) {
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		errMsg := err.Error()

		switch {
			case strings.Contains(errMsg, "no such file or directory"):
				return Config{}, errors.New("config file does not exist")
		}
	}

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

func (s *ConfigService) CreateConfig(
	lang string,
	containerRuntime string,
) error {
	//If the config file already exists, return nil.
	if _, err := os.Stat(s.path); err == nil {
		return  nil		
	}

	dir := filepath.Dir(s.path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	config := Config{
		Lang: lang,
		ContainerRuntime: containerRuntime,
	};

	data, err := json.MarshalIndent(config, "", "  ")

	if err != nil {
		return  err
	}
	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return err
	}

	return nil
}