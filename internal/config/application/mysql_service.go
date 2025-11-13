package application

import (
	"fmt"
	"myenv/internal/infrastructure"
	"myenv/internal/utils"
	"os"
	"path/filepath"
)

type (
	MySQLService struct {
		container      infrastructure.ContainerInterface
		repository     infrastructure.RepositoryInterface
		config_service ConfigService
	}
)

func NewMySQLService(
	container infrastructure.ContainerInterface,
	repository infrastructure.RepositoryInterface,
	config_service ConfigService,
) *MySQLService {
	return &MySQLService{
		container:      container,
		repository:     repository,
		config_service: config_service,
	}
}

func (s *MySQLService) Create(events chan<- Event) error {
	events <- Event{
		Key:     "clone_mysql_repository",
		Name:    "Clone MySQL Repository",
		Status:  "running",
		Message: "Cloning MySQL repository...",
	}

	targetRepo := "https://github.com/takashiraki/docker_mysql.git"

	homeDir, err := os.UserHomeDir()

	if err != nil {
		events <- Event{
			Key:     "clone_mysql_repository",
			Status:  "error",
			Message: "Failed to get home directory",
		}
		return err
	}

	targetPath := filepath.Join(homeDir, "dev", "docker_mysql")

	moduleConfig := Module{
		Name: "mysql",
		Path: targetPath,
	}

	if err := s.config_service.AddModule(moduleConfig); err != nil {
		return err
	}

	if err := s.repository.CloneRepo(targetRepo, targetPath); err != nil {
		events <- Event{
			Key:     "clone_mysql_repository",
			Status:  "error",
			Message: "Failed to clone MySQL repository",
		}
		return err
	}

	events <- Event{
		Key:     "clone_mysql_repository",
		Name:    "Clone MySQL Repository",
		Status:  "success",
		Message: "MySQL repository cloned successfully",
	}

	events <- Event{
		Key:     "set_up_environment_variables",
		Name:    "Set up environment variables",
		Status:  "Pending",
		Message: "Setting up environment variables...",
	}

	if err := utils.CreateEnvFile(targetPath); err != nil {
		events <- Event{
			Key:     "set_up_environment_variables",
			Status:  "error",
			Message: "Failed to create .env file",
		}
		return err
	}

	envFilePath := filepath.Join(targetPath, ".env")

	content, err := os.ReadFile(envFilePath)

	if err != nil {
		events <- Event{
			Key:     "set_up_environment_variables",
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

	if err := utils.ReplaceAllValue(&updateContent, replacements); err != nil {
		events <- Event{
			Key:     "set_up_environment_variables",
			Status:  "error",
			Message: "Failed to update .env file",
		}
		return err
	}

	if err := os.WriteFile(envFilePath, []byte(updateContent), 0644); err != nil {
		events <- Event{
			Key:     "set_up_environment_variables",
			Status:  "error",
			Message: "Failed to write .env file",
		}
		return err
	}

	events <- Event{
		Key:     "set_up_environment_variables",
		Name:    "Set up environment variables",
		Status:  "success",
		Message: "Environment variables set up successfully",
	}

	events <- Event{
		Key:     "start_mysql_containers",
		Name:    "Start MySQL containers",
		Status:  "running",
		Message: "Starting MySQL containers...",
	}

	if err := s.container.CreateContainer(targetPath); err != nil {
		events <- Event{
			Key:     "start_mysql_containers",
			Status:  "error",
			Message: "Failed to start MySQL containers",
		}

		return err
	}

	events <- Event{
		Key:     "mysql_setup_completed",
		Name:    "MySQL setup completed",
		Status:  "success",
		Message: "MySQL setup completed successfully",
	}

	return nil
}
