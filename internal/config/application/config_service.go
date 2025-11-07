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
		Projects         map[string]Project
		Modules          map[string]Module
	}

	Project struct {
		ContainerName string
		ContainerProxy string
		Path          string
		Lang          string
		Fw            string
		Options       map[string]string
		Modules       []string
	}

	Module struct {
		Name string
		Path string
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
		Projects: make(map[string]Project),
		Modules: make(map[string]Module),
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

func (s *ConfigService) GetProject(name string) (Project, error) {
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		errMsg := err.Error()

		switch {
			case strings.Contains(errMsg, "no such file or directory"):
				return Project{}, errors.New("project file does not exist")
		}
	}

	config, err := s.GetConfig()

	if err != nil {
		return Project{}, err
	}

	project, exists := config.Projects[name]

	if !exists {
		return Project{}, errors.New("project not found")
	}

	return project, nil
}

func (s *ConfigService) SaveConfig(config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	
	if err != nil {
		return err
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return err
	}
	
	return nil
}

func (s *ConfigService) AddProject(project Project) error {
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		errMsg := err.Error()

		switch {
			case strings.Contains(errMsg, "no such file or directory"):
				return errors.New("project file does not exist")
		}
	}

	config, err := s.GetConfig()
	
	if err != nil {
		return err
	}

	if _, exists := config.Projects[project.ContainerName]; exists {
		return errors.New("project already exists")
	}
	
	config.Projects[project.ContainerName] = project

	if err := s.SaveConfig(config); err != nil {
		return err
	}

	return nil
}