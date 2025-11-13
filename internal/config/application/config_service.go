package application

import (
	"encoding/json"
	"errors"
	"myenv/internal/infrastructure"
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

	Event struct {
		Key     string
		Name    string
		Status  string
		Message string
	}

	ConfigService struct {
		path string
		container infrastructure.ContainerInterface
	}
)

func NewConfigService(container infrastructure.ContainerInterface) (*ConfigService, error) {
	homedir, err := os.UserHomeDir()

	if err != nil {
		return nil, err
	}

	path := filepath.Join(homedir, ".config", "myenv", "config.json")

	return &ConfigService{
		path: path,
		container: container,
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
	events chan<- Event,
) error {
	//If the config file already exists, return nil.
	if _, err := os.Stat(s.path); err == nil {
		return  nil
	}

	events <- Event{
		Key: "create_config_file",
		Name: "Create config file",
		Status: "running",
		Message: "Creating config file...",
	}

	dir := filepath.Dir(s.path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		events <- Event{
			Key: "create_config_file",
			Status: "error",
			Message: "Failed to create config file",
		}

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
		events <- Event{
			Key: "create_config_file",
			Status: "error",
			Message: "Failed to marshal config file",
		}
		return  err
	}
	if err := os.WriteFile(s.path, data, 0644); err != nil {
		events <- Event{
			Key: "create_config_file",
			Status: "error",
			Message: "Failed to write config file",
		}
		return err
	}

	events <- Event{
		Key: "create_config_file",
		Name: "Create config file",
		Status: "success",
		Message: "Config file created successfully",
	}

	events <- Event{
		Key: "create_my_proxy_network",
		Name: "Create my proxy network",
		Status: "running",
		Message: "Creating my proxy network...",
	}

	if err := s.container.ChechInfraNetworkExists(); err == nil {
		events <- Event{
			Key: "create_my_proxy_network",
			Status: "success",
			Message: "My proxy network already exists",
		}
	} else {
		if err := s.container.CreateProxyNetwork(); err != nil {
			events <- Event{
				Key: "create_my_proxy_network",
				Status: "error",
				Message: "Failed to create my proxy network",
			}
			return err
		}

		events <- Event{
			Key: "create_my_proxy_network",
			Name: "Create my proxy network",
			Status: "success",
			Message: "My proxy network created successfully",
		}
	}

	events <- Event{
		Key: "create_my_infra_network",
		Name: "Create my infra network",
		Status: "running",
		Message: "Creating my infra network...",
	}

	if err := s.container.ChechInfraNetworkExists(); err == nil {
		events <- Event{
			Key: "create_my_infra_network",
			Status: "success",
			Message: "My infra network already exists",
		}
	} else {
		if err := s.container.CreateInfraNetwork(); err != nil {
			events <- Event{
				Key: "create_my_infra_network",
				Status: "error",
				Message: "Failed to create my infra network",
			}
			return err
		}
		
		events <- Event{
			Key: "create_my_infra_network",
			Name: "Create my infra network",
			Status: "success",
			Message: "My infra network created successfully",
		}
	}


	return nil
}

func (s *ConfigService) GetProjects() ([]Project, error) {
	config, err := s.GetConfig()
	if err != nil {
		return nil, err
	}
	
	var projects []Project

	for _,project := range config.Projects {
		projects = append(projects, project)
	}

	return projects, nil
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

func (s *ConfigService) AddModule(module Module) error {
	config, err := s.GetConfig()

	if err != nil {
		return err
	}

	if _, exists := config.Modules[module.Name]; exists {
		return errors.New("module already exists")
	}

	config.Modules[module.Name] = module

	if err := s.SaveConfig(config); err != nil {
		return err
	}

	return nil
}

func (s *ConfigService) UpProject(name string) (Project, error) {

	project, err := s.GetProject(name)

	if err != nil {
		return Project{}, err
	}

	modules := project.Modules
	
	for _, module := range modules {
		targetModulem, err := s.GetModule(module)

		if err != nil {
			return Project{}, err
		}

		if err := s.container.CreateContainer(targetModulem.Path); err != nil {
			return Project{}, err
		}
	}

	if err := s.container.CreateContainer(project.Path); err != nil {
		return Project{}, err
	}
	
	return project, nil
} 

func (s *ConfigService) GetModule(name string) (Module, error) {
	config, err := s.GetConfig()

	if err != nil {
		return Module{}, err
	}

	module, exists := config.Modules[name]

	if !exists {
		return Module{}, errors.New("module not found")
	}

	return module, nil
}