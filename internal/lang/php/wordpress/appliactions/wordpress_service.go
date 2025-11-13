package appliactions

import (
	"errors"
	"fmt"
	"myenv/internal/config/application"
	"myenv/internal/events"
	"myenv/internal/infrastructure"
	"myenv/internal/utils"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type (
	WordpressService struct {
		container      infrastructure.ContainerInterface
		repository     infrastructure.RepositoryInterface
		config_service application.ConfigService
	}
)

func NewWordpressService(
	container infrastructure.ContainerInterface,
	repository infrastructure.RepositoryInterface,
	config_service application.ConfigService,
) *WordpressService {
	return &WordpressService{
		container:      container,
		repository:     repository,
		config_service: config_service,
	}
}

func (s *WordpressService) Create(
	eventChan chan<- events.Event,
	containerName string,
	virtualHost string,
) error {
	eventChan <- events.Event{
		Key:     "clone_wordpress_repository",
		Name:    "Clone WordPress Repository",
		Status:  "running",
		Message: "Cloning WordPress repository...",
	}

	homeDir, err := os.UserHomeDir()

	if err != nil {
		eventChan <- events.Event{
			Key:     "clone_wordpress_repository",
			Name:    "Clone WordPress Repository",
			Status:  "error",
			Message: "Failed to get home directory",
		}
		return err
	}

	targetPath := filepath.Join(homeDir, "dev", containerName)

	if _, err := os.Stat(targetPath); err == nil {
		eventChan <- events.Event{
			Key:     "clone_wordpress_repository",
			Name:    "Clone WordPress Repository",
			Status:  "error",
			Message: "Directory already exists",
		}
		return errors.New("directory already exists")
	}

	modules := []string{
		"proxy",
		"mysql",
		"mailpit",
	}

	moduleConfig := application.Project{
		ContainerName:  containerName,
		ContainerProxy: virtualHost,
		Path:           targetPath,
		Lang:           "php",
		Fw:             "wordpress",
		Options: map[string]string{
			"type": "new",
		},
		Modules: modules,
	}

	if err := s.config_service.AddProject(moduleConfig); err != nil {
		eventChan <- events.Event{
			Key:     "clone_wordpress_repository",
			Name:    "Clone WordPress Repository",
			Status:  "error",
			Message: "Failed to add project to config",
		}
		return err
	}

	targetRepo := "https://github.com/takashiraki/docker_wordpress.git"

	if err := s.repository.CloneRepo(targetRepo, targetPath); err != nil {
		eventChan <- events.Event{
			Key:     "clone_wordpress_repository",
			Name:    "Clone WordPress Repository",
			Status:  "error",
			Message: "Failed to clone WordPress repository",
		}
		return err
	}

	eventChan <- events.Event{
		Key:     "clone_wordpress_repository",
		Name:    "Clone WordPress Repository",
		Status:  "success",
		Message: "WordPress repository cloned successfully",
	}

	eventChan <- events.Event{
		Key:     "set_up_environment_variables",
		Name:    "Set up environment variables",
		Status:  "running",
		Message: "Setting up environment variables...",
	}

	if err := utils.CreateEnvFile(targetPath); err != nil {
		eventChan <- events.Event{
			Key:     "set_up_environment_variables",
			Name:    "Set up environment variables",
			Status:  "error",
			Message: "Failed to create .env file",
		}
		return err
	}

	envFilePath := filepath.Join(targetPath, ".env")

	content, err := os.ReadFile(envFilePath)

	if err != nil {
		eventChan <- events.Event{
			Key:     "set_up_environment_variables",
			Name:    "Set up environment variables",
			Status:  "error",
			Message: "Failed to read .env file",
		}
		return err
	}

	dbName, err := sanitizeDatabaseName(containerName)
	if err != nil {
		eventChan <- events.Event{
			Key:     "create_wordpress_database",
			Name:    "Create WordPress database",
			Status:  "error",
			Message: "Failed to sanitize database name",
		}
	}

	updateContent := string(content)

	replacements := map[string]any{
		"MY_WORDPRESS_DB=": fmt.Sprintf("MY_WORDPRESS_DB=%s", dbName),
		"CONTAINER_NAME=":  fmt.Sprintf("CONTAINER_NAME=%s", containerName),
		"VIRTUAL_HOST=":    fmt.Sprintf("VIRTUAL_HOST=%s", virtualHost),
	}

	if err := utils.ReplaceAllValue(&updateContent, replacements); err != nil {
		eventChan <- events.Event{
			Key:     "set_up_environment_variables",
			Name:    "Set up environment variables",
			Status:  "error",
			Message: "Failed to update .env file",
		}
		return err
	}

	if err := os.WriteFile(envFilePath, []byte(updateContent), 0644); err != nil {
		eventChan <- events.Event{
			Key:     "set_up_environment_variables",
			Name:    "Set up environment variables",
			Status:  "error",
			Message: "Failed to write .env file",
		}
		return err
	}

	eventChan <- events.Event{
		Key:     "set_up_environment_variables",
		Name:    "Set up environment variables",
		Status:  "success",
		Message: "Environment variables set up successfully",
	}

	eventChan <- events.Event{
		Key:     "create_devcontainer_settings",
		Name:    "Create devcontainer settings",
		Status:  "running",
		Message: "Creating devcontainer settings...",
	}

	devContainerExamplePath := filepath.Join(targetPath, ".devcontainer", "devcontainer.json.example")

	if _, err := os.Stat(devContainerExamplePath); os.IsNotExist(err) {
		eventChan <- events.Event{
			Key:     "create_devcontainer_settings",
			Name:    "Create devcontainer settings",
			Status:  "error",
			Message: "devcontainer.json.example does not exist",
		}
		return err
	}

	devContainerPath := filepath.Join(targetPath, ".devcontainer", "devcontainer.json")

	if err := utils.CopyFile(devContainerExamplePath, devContainerPath); err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_settings",
			Name:    "Create devcontainer settings",
			Status:  "error",
			Message: "Failed to create devcontainer settings",
		}
		return err
	}

	devContainerContents, err := os.ReadFile(devContainerPath)

	if err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_settings",
			Name:    "Create devcontainer settings",
			Status:  "error",
			Message: "Failed to read devcontainer.json",
		}
		return err
	}

	updateDevContainerContents := string(devContainerContents)

	replacements = map[string]any{
		`"name": "my wordpress",`: fmt.Sprintf(`"name": "%s",`, containerName),
	}

	if err := utils.ReplaceAllValue(&updateDevContainerContents, replacements); err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_settings",
			Name:    "Create devcontainer settings",
			Status:  "error",
			Message: "Failed to update devcontainer.json",
		}
		return err
	}

	if err := os.WriteFile(devContainerPath, []byte(updateDevContainerContents), 0644); err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_settings",
			Name:    "Create devcontainer settings",
			Status:  "error",
			Message: "Failed to write devcontainer.json",
		}
		return err
	}

	eventChan <- events.Event{
		Key:     "create_devcontainer_settings",
		Name:    "Create devcontainer settings",
		Status:  "success",
		Message: "Devcontainer settings created successfully",
	}

	eventChan <- events.Event{
		Key:     "Resolve dependencied container booting",
		Name:    "resolve_devcontainer_settings",
		Status:  "running",
		Message: "Resolving dependencied container booting...",
	}

	for _, module := range modules {
		moduleObject, err := s.config_service.GetModule(module)

		if err != nil {
			eventChan <- events.Event{
				Key:     "resolve_devcontainer_settings",
				Name:    "resolve_devcontainer_settings",
				Status:  "error",
				Message: fmt.Sprintf("Failed to get module: %s", module),
			}
			return err
		}

		eventChan <- events.Event{
			Key:     "resolve_devcontainer_settings",
			Name:    "resolve_devcontainer_settings",
			Status:  "running",
			Message: fmt.Sprintf("Resolving %s module...", module),
		}

		if err := s.container.CreateContainer(moduleObject.Path); err != nil {
			eventChan <- events.Event{
				Key:     "resolve_devcontainer_settings",
				Name:    "resolve_devcontainer_settings",
				Status:  "error",
				Message: fmt.Sprintf("Failed to create %s container", module),
			}
			return err
		}
	}

	eventChan <- events.Event{
		Key:     "resolve_devcontainer_settings",
		Name:    "resolve_devcontainer_settings",
		Status:  "success",
		Message: "Resolved dependencied container booting successfully",
	}

	eventChan <- events.Event{
		Key:     "create_wordpress_database",
		Name:    "Create WordPress database",
		Status:  "running",
		Message: "Creating WordPress database...",
	}

	for i := 0; i < 15; i++ {
		if err := s.container.ExecCommand(
			"my_database",
			"mysqladmin",
			"ping",
			"-h", "localhost",
			"-uroot",
			"-prootpw",
		); err == nil {
			break
		}

		time.Sleep(2 * time.Second)
	}

	if err := s.container.ExecCommand(
		"my_database",
		"mysql",
		"-uroot",
		"-prootpw",
		"-e",
		fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName),
	); err != nil {
		eventChan <- events.Event{
			Key:     "create_wordpress_database",
			Name:    "Create WordPress database",
			Status:  "error",
			Message: "Failed to create WordPress database",
		}
		return err
	}

	eventChan <- events.Event{
		Key:     "create_wordpress_database",
		Name:    "Create WordPress database",
		Status:  "success",
		Message: "WordPress database created successfully",
	}

	eventChan <- events.Event{
		Key:     "start_wordpress_containers",
		Name:    "Start WordPress containers",
		Status:  "running",
		Message: "Starting WordPress containers...",
	}

	if err := s.container.CreateContainer(targetPath); err != nil {
		eventChan <- events.Event{
			Key:     "start_wordpress_containers",
			Name:    "Start WordPress containers",
			Status:  "error",
			Message: "Failed to start WordPress containers",
		}
		return err
	}

	eventChan <- events.Event{
		Key:     "start_wordpress_containers",
		Name:    "Start WordPress containers",
		Status:  "success",
		Message: "WordPress containers started successfully",
	}

	eventChan <- events.Event{
		Key:     "wordpress_setup_completed",
		Name:    "WordPress setup completed",
		Status:  "success",
		Message: "WordPress setup completed successfully",
	}

	return nil
}

func sanitizeDatabaseName(name string) (string, error) {
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, "`", "")

	if matched, err := regexp.MatchString("^[a-zA-Z0-9_]+$", name); err != nil || !matched {
		return "", errors.New("invalid database name")
	}
	return name, nil
}
