package wordpress

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

	moduleConfig := application.Project{
		ContainerName:  containerName,
		ContainerProxy: virtualHost,
		Path:           targetPath,
		Lang:           "php",
		Fw:             "wordpress",
		Options: map[string]string{
			"type": "new",
		},
		Modules: []string{
			"proxy",
			"mysql",
			"mailpit",
		},
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

	updateContent := string(content)

	replacements := map[string]any{
		"MY_WORDPRESS_DB=": fmt.Sprintf("MY_WORDPRESS_DB=%s", containerName),
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
		Key:     "create_wordpress_database",
		Name:    "Create WordPress database",
		Status:  "running",
		Message: "Creating WordPress database...",
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

	if err := s.container.ExecCommand(
		"my_database",
		fmt.Sprintf(
			"mysql -uroot -prootpw -e 'CREATE DATABASE IF NOT EXISTS `%s`'",
			dbName,
		),
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
