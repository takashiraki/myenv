package applications

import (
	"myenv/internal/config/application"
	"myenv/internal/events"
	"myenv/internal/infrastructure"
	"os"
	"path/filepath"
)

type (
	LaravelService struct {
		container      infrastructure.ContainerInterface
		repository     infrastructure.RepositoryInterface
		config_service application.ConfigService
	}
)

func NewLaravelService(
	container infrastructure.ContainerInterface,
	repository infrastructure.RepositoryInterface,
	config_service application.ConfigService,
) *LaravelService {
	return &LaravelService{
		container:      container,
		repository:     repository,
		config_service: config_service,
	}
}

func (s *LaravelService) Create(
	eventChan chan<- events.Event,
	containerName string,
	virtualHost string,
) error {
	eventChan <- events.Event{
		Key:     "clone_laravel_repository",
		Name:    "Clone Laravel Repository",
		Status:  "running",
		Message: "Cloning Laravel repository...",
	}

	homeDir, err := os.UserHomeDir()

	if err != nil {
		eventChan <- events.Event{
			Key:     "clone_laravel_repository",
			Name:    "Clone Laravel Repository",
			Status:  "error",
			Message: "Failed to get user home directory: " + err.Error(),
		}

		return err
	}

	targetPath := filepath.Join(homeDir, "dev", containerName)

	if _, err := os.Stat(targetPath); err == nil {
		eventChan <- events.Event{
			Key:     "clone_laravel_repository",
			Name:    "Clone Laravel Repository",
			Status:  "error",
			Message: "Target path already exists: " + targetPath,
		}

		return err
	}

	modules := []string{
		"proxy",
		"mysql",
		"mailpit",
	}

	projectConfig := application.Project{
		ContainerName:  containerName,
		ContainerProxy: virtualHost,
		Path:           targetPath,
		Lang:           "php",
		Fw:             "laravel",
		Options: map[string]string{
			"type": "new",
		},
		Modules: modules,
	}

	if err := s.config_service.AddProject(projectConfig); err != nil {
		eventChan <- events.Event{
			Key:     "clone_laravel_repository",
			Name:    "Clone Laravel Repository",
			Status:  "error",
			Message: "Failed to add project configuration: " + err.Error(),
		}

		return err
	}

	targetRepo := "https://github.com/takashiraki/docker_laravel.git"

	return nil
}
