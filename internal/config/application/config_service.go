package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"myenv/internal/infrastructure"
	CommonUtils "myenv/internal/utils"
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
		ContainerName  string
		ContainerProxy string
		Path           string
		Lang           string
		Fw             string
		Options        map[string]string
		Modules        []string
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
		path       string
		container  infrastructure.ContainerInterface
		repository infrastructure.RepositoryInterface
	}
)

func NewConfigService(container infrastructure.ContainerInterface, repository infrastructure.RepositoryInterface) (*ConfigService, error) {
	homedir, err := os.UserHomeDir()

	if err != nil {
		return nil, err
	}

	path := filepath.Join(homedir, ".config", "myenv", "config.json")

	return &ConfigService{
		path:       path,
		container:  container,
		repository: repository,
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
	quick bool,
) error {
	//If the config file already exists, return nil.
	if _, err := os.Stat(s.path); err == nil {
		return nil
	}

	events <- Event{
		Key:     "create_config_file",
		Name:    "Create config file",
		Status:  "running",
		Message: "Creating config file...",
	}

	dir := filepath.Dir(s.path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		events <- Event{
			Key:     "create_config_file",
			Status:  "error",
			Message: "Failed to create config file",
		}

		return err
	}

	config := Config{
		Lang:             lang,
		ContainerRuntime: containerRuntime,
		Projects:         make(map[string]Project),
		Modules:          make(map[string]Module),
	}

	data, err := json.MarshalIndent(config, "", "  ")

	if err != nil {
		events <- Event{
			Key:     "create_config_file",
			Status:  "error",
			Message: "Failed to marshal config file",
		}
		return err
	}
	if err := os.WriteFile(s.path, data, 0644); err != nil {
		events <- Event{
			Key:     "create_config_file",
			Status:  "error",
			Message: "Failed to write config file",
		}
		return err
	}

	events <- Event{
		Key:     "create_config_file",
		Name:    "Create config file",
		Status:  "success",
		Message: "Config file created successfully",
	}

	events <- Event{
		Key:     "create_my_proxy_network",
		Name:    "Create my proxy network",
		Status:  "running",
		Message: "Creating my proxy network...",
	}

	if err := s.container.ChechInfraNetworkExists(); err == nil {
		events <- Event{
			Key:     "create_my_proxy_network",
			Status:  "success",
			Message: "My proxy network already exists",
		}
	} else {
		if err := s.container.CreateProxyNetwork(); err != nil {
			events <- Event{
				Key:     "create_my_proxy_network",
				Status:  "error",
				Message: "Failed to create my proxy network",
			}
			return err
		}

		events <- Event{
			Key:     "create_my_proxy_network",
			Name:    "Create my proxy network",
			Status:  "success",
			Message: "My proxy network created successfully",
		}
	}

	events <- Event{
		Key:     "create_my_infra_network",
		Name:    "Create my infra network",
		Status:  "running",
		Message: "Creating my infra network...",
	}

	if err := s.container.ChechInfraNetworkExists(); err == nil {
		events <- Event{
			Key:     "create_my_infra_network",
			Status:  "success",
			Message: "My infra network already exists",
		}
	} else {
		if err := s.container.CreateInfraNetwork(); err != nil {
			events <- Event{
				Key:     "create_my_infra_network",
				Status:  "error",
				Message: "Failed to create my infra network",
			}
			return err
		}

		events <- Event{
			Key:     "create_my_infra_network",
			Name:    "Create my infra network",
			Status:  "success",
			Message: "My infra network created successfully",
		}
	}

	if !quick {
		return nil
	}

	homeDir, err := os.UserHomeDir()

	if err != nil {
		return err
	}

	proxyDir := filepath.Join(homeDir, "dev", "docker_proxy_network")

	if _, err := os.Stat(proxyDir); os.IsNotExist(err) {
		events <- Event{
			Key:     "create_proxy_container",
			Name:    "Create proxy container",
			Status:  "running",
			Message: "Cloning proxy repository...",
		}

		proxyRepo := "https://github.com/takashiraki/docker_proxy_network.git"

		if err := s.repository.CloneRepo(proxyRepo, proxyDir); err != nil {
			events <- Event{
				Key:     "create_proxy_container",
				Name:    "Create proxy container",
				Status:  "error",
				Message: "Failed to clone proxy repository",
			}
			return err
		}

		if err := s.container.CreateContainer(proxyDir); err != nil {
			events <- Event{
				Key:     "create_proxy_container",
				Name:    "Create proxy container",
				Status:  "error",
				Message: "Failed to create proxy container",
			}
			return err
		}

		moduleConfig := Module{
			Name: "proxy",
			Path: proxyDir,
		}

		if err := s.AddModule(moduleConfig); err != nil {
			events <- Event{
				Key:     "create_proxy_container",
				Name:    "Create proxy container",
				Status:  "error",
				Message: "Failed to add proxy module to config",
			}
			return err
		}

		events <- Event{
			Key:     "create_proxy_container",
			Name:    "Create proxy container",
			Status:  "success",
			Message: "Proxy module added to config successfully",
		}
	}

	mysqlDir := filepath.Join(homeDir, "dev", "docker_mysql")

	if _, err := os.Stat(mysqlDir); os.IsNotExist(err) {
		events <- Event{
			Key:     "create_mysql_container",
			Name:    "Create mysql container",
			Status:  "running",
			Message: "Cloning mysql repository...",
		}

		mysqlRepo := "https://github.com/takashiraki/docker_mysql.git"

		if err := s.repository.CloneRepo(mysqlRepo, mysqlDir); err != nil {
			events <- Event{
				Key:     "create_mysql_container",
				Name:    "Create mysql container",
				Status:  "error",
				Message: "Failed to clone mysql repository",
			}
			return err
		}

		if err := CommonUtils.CreateEnvFile(mysqlDir); err != nil {
			events <- Event{
				Key:     "create_mysql_container",
				Name:    "Create mysql container",
				Status:  "error",
				Message: "Failed to create .env file",
			}
			return err
		}

		envFilePath := filepath.Join(mysqlDir, ".env")

		content, err := os.ReadFile(envFilePath)

		if err != nil {
			events <- Event{
				Key:     "create_mysql_container",
				Name:    "Create mysql container",
				Status:  "error",
				Message: "Failed to read .env file",
			}

			return err
		}

		mysql_host := "my_database"
		mysql_user := "myenv"
		mysql_password := "myenv"
		mysql_root_password := "rootpw"
		db_port := "3306"

		updateContent := string(content)

		replacements := map[string]any{
			"MYSQL_HOST=":          fmt.Sprintf("MYSQL_HOST=%s", mysql_host),
			"MYSQL_USER=":          fmt.Sprintf("MYSQL_USER=%s", mysql_user),
			"MYSQL_PASSWORD=":      fmt.Sprintf("MYSQL_PASSWORD=%s", mysql_password),
			"MYSQL_ROOT_PASSWORD=": fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", mysql_root_password),
			"DB_PORT=":             fmt.Sprintf("DB_PORT=%s", db_port),
		}

		if err := CommonUtils.ReplaceAllValue(&updateContent, replacements); err != nil {
			events <- Event{
				Key:     "create_mysql_container",
				Name:    "Create mysql container",
				Status:  "error",
				Message: "Failed to update .env file",
			}
			return err
		}

		if err := os.WriteFile(envFilePath, []byte(updateContent), 0644); err != nil {
			events <- Event{
				Key:     "create_mysql_container",
				Name:    "Create mysql container",
				Status:  "error",
				Message: "Failed to write .env file",
			}
			return err
		}

		if err := s.container.CreateContainer(mysqlDir); err != nil {
			events <- Event{
				Key:     "create_mysql_container",
				Name:    "Create mysql container",
				Status:  "error",
				Message: "Failed to create mysql container",
			}
			return err
		}

		moduleConfig := Module{
			Name: "mysql",
			Path: mysqlDir,
		}

		if err := s.AddModule(moduleConfig); err != nil {
			events <- Event{
				Key:     "create_mysql_container",
				Name:    "Create mysql container",
				Status:  "error",
				Message: "Failed to add mysql module to config",
			}
			return err
		}

		events <- Event{
			Key:     "create_mysql_container",
			Name:    "Create mysql container",
			Status:  "success",
			Message: "Mysql module added to config successfully",
		}
	}

	mailpitDir := filepath.Join(homeDir, "dev", "docker_mailpit")

	if _, err := os.Stat(mailpitDir); os.IsNotExist(err) {
		events <- Event{
			Key:     "create_mailpit_container",
			Name:    "Create mailpit container",
			Status:  "running",
			Message: "Cloning mailpit repository...",
		}

		mailpitRepo := "https://github.com/takashiraki/docker_mailpit.git"
		if err := s.repository.CloneRepo(mailpitRepo, mailpitDir); err != nil {
			events <- Event{
				Key:     "create_mailpit_container",
				Name:    "Create mailpit container",
				Status:  "error",
				Message: "Failed to clone mailpit repository",
			}
			return err
		}

		if err := CommonUtils.CreateEnvFile(mailpitDir); err != nil {
			events <- Event{
				Key:     "create_mailpit_container",
				Name:    "Create mailpit container",
				Status:  "error",
				Message: "Failed to create .env file",
			}
			return err
		}

		envFilePath := filepath.Join(mailpitDir, ".env")

		content, err := os.ReadFile(envFilePath)

		if err != nil {
			events <- Event{
				Key:     "create_mailpit_container",
				Name:    "Create mailpit container",
				Status:  "error",
				Message: "Failed to read .env file",
			}
			return err
		}

		mp_database := "/data/mailpit.db"
		mp_max_messages := 5000
		mp_smtp_uauth_accept_any := 1
		mp_smtp_auth_allow_insecure := 1
		virtual_host := "mailpit.localhost"
		virtual_port := 8025

		updateContent := string(content)

		replacements := map[string]any{
			"MP_DATABASE=":                 fmt.Sprintf("MP_DATABASE=%s", mp_database),
			"MP_MAX_MESSAGES=":             fmt.Sprintf("MP_MAX_MESSAGES=%d", mp_max_messages),
			"MP_SMTP_UAUTH_ACCEPT_ANY=":    fmt.Sprintf("MP_SMTP_UAUTH_ACCEPT_ANY=%d", mp_smtp_uauth_accept_any),
			"MP_SMTP_AUTH_ALLOW_INSECURE=": fmt.Sprintf("MP_SMTP_AUTH_ALLOW_INSECURE=%d", mp_smtp_auth_allow_insecure),
			"VIRTUAL_HOST=":                fmt.Sprintf("VIRTUAL_HOST=%s", virtual_host),
			"VIRTUAL_PORT=":                fmt.Sprintf("VIRTUAL_PORT=%d", virtual_port),
		}

		if err := CommonUtils.ReplaceAllValue(&updateContent, replacements); err != nil {
			events <- Event{
				Key:     "create_mailpit_container",
				Name:    "Create mailpit container",
				Status:  "error",
				Message: "Failed to update .env file",
			}
			return err
		}

		if err := os.WriteFile(envFilePath, []byte(updateContent), 0644); err != nil {
			events <- Event{
				Key:     "create_mailpit_container",
				Name:    "Create mailpit container",
				Status:  "error",
				Message: "Failed to write .env file",
			}
			return err
		}

		if err := s.container.CreateContainer(mailpitDir); err != nil {
			events <- Event{
				Key:     "create_mailpit_container",
				Name:    "Create mailpit container",
				Status:  "error",
				Message: "Failed to create mailpit container",
			}
			return err
		}

		moduleConfig := Module{
			Name: "mailpit",
			Path: mailpitDir,
		}

		if err := s.AddModule(moduleConfig); err != nil {
			events <- Event{
				Key:     "create_mailpit_container",
				Name:    "Create mailpit container",
				Status:  "error",
				Message: "Failed to add mailpit module to config",
			}
			return err
		}

		events <- Event{
			Key:     "create_mailpit_container",
			Name:    "Create mailpit container",
			Status:  "success",
			Message: "Mailpit module added to config successfully",
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

	for _, project := range config.Projects {
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
