package applications

import (
	"fmt"
	"myenv/internal/config/application"
	"myenv/internal/infrastructure"
	"myenv/internal/utils"
	"os"
	"path/filepath"
	"time"
)

type (
	MailpitService struct {
		container infrastructure.ContainerInterface
		repository infrastructure.RepositoryInterface
		config_service application.ConfigService
	}
)

func NewMailpitService(
	container infrastructure.ContainerInterface,
	repository infrastructure.RepositoryInterface,
	config_service application.ConfigService,
) *MailpitService {
	return &MailpitService{
		container: container,
		repository: repository,
		config_service: config_service,
	}
}

func (s *MailpitService) Create(events chan<- Event) error {
	events <- Event{
		Key: "clone_mailpit_repository",
		Name: "Clone Mailpit Repository",
		Status: "running",
		Message: "Cloning Mailpit repository...",
	}

	targetRepo := "https://github.com/takashiraki/docker_mailpit.git"

	homeDir, err := os.UserHomeDir()

	if err != nil {
		events <- Event{
			Key: "clone_mailpit_repository",
			Status: "error",
			Message: "Failed to get home directory",
		}
		return err
	}

	targetPath := filepath.Join(homeDir, "dev", "docker_mailpit")

	moduleConfig := application.Module{
		Name: "mailpit",
		Path: targetPath,
	}

	if err := s.config_service.AddModule(moduleConfig); err != nil {
		events <- Event{
			Key: "clone_mailpit_repository",
			Status: "error",
			Message: "Failed to add module to config",
		}
		return err
	}

	if err := s.repository.CloneRepo(targetRepo, targetPath); err != nil {
		events <- Event{
			Key: "clone_mailpit_repository",
			Status: "error",
			Message: "Failed to clone Mailpit repository",
		}
		return err
	}

	events <- Event{
		Key: "clone_mailpit_repository",
		Name: "Clone Mailpit Repository",
		Status: "success",
		Message: "Mailpit repository cloned successfully",
	}

	events <- Event{
		Key: "set_up_environment_variables",
		Name: "Set up environment variables",
		Status: "running",
		Message: "Setting up environment variables...",
	}

	if err := utils.CreateEnvFile(targetPath); err != nil {
		events <- Event{
			Key: "set_up_environment_variables",
			Status: "error",
			Message: "Failed to create .env file",
		}
		return err
	}

	envFilePath := filepath.Join(targetPath, ".env")

	content, err := os.ReadFile(envFilePath)

	if err != nil {
		events <- Event{
			Key: "set_up_environment_variables",
			Status: "error",
			Message: "Failed to read .env file",
		}
		return err
	}

	mp_database := "/data/mailpit.db"
	mp_max_messages := 5000
	mp_smtp_uauth_accept_any := 1
	mp_smtp_auth_allow_insecure := 1
	virtual_host := "mailpit.localhost"
	virtual_port := 1025

	updateContent := string(content)

	replacements := map[string]any{
		"MP_DATABASE=": fmt.Sprintf("MP_DATABASE=%s", mp_database),
		"MP_MAX_MESSAGES=": fmt.Sprintf("MP_MAX_MESSAGES=%d", mp_max_messages),
		"MP_SMTP_UAUTH_ACCEPT_ANY=": fmt.Sprintf("MP_SMTP_UAUTH_ACCEPT_ANY=%d", mp_smtp_uauth_accept_any),
		"MP_SMTP_AUTH_ALLOW_INSECURE=": fmt.Sprintf("MP_SMTP_AUTH_ALLOW_INSECURE=%d", mp_smtp_auth_allow_insecure),
		"VIRTUAL_HOST=": fmt.Sprintf("VIRTUAL_HOST=%s", virtual_host),
		"VIRTUAL_PORT=": fmt.Sprintf("VIRTUAL_PORT=%d", virtual_port),
		"TZ=": fmt.Sprintf("TZ=%s", time.Now().Location().String()),
	}

	if err := utils.ReplaceAllValue(&updateContent, replacements); err != nil {
		events <- Event{
			Key: "set_up_environment_variables",
			Status: "error",
			Message: "Failed to update .env file",
		}
		return err
	}

	if err := os.WriteFile(envFilePath, []byte(updateContent), 0644); err != nil {
		events <- Event{
			Key: "set_up_environment_variables",
			Status: "error",
			Message: "Failed to write .env file",
		}
		return err
	}

	events <- Event{
		Key: "set_up_environment_variables",
		Name: "Set up environment variables",
		Status: "success",
		Message: "Environment variables set up successfully",
	}

	events <- Event{
		Key: "start_mailpit_containers",
		Name: "Start Mailpit containers",
		Status: "running",
		Message: "Starting Mailpit containers...",
	}

	if err := s.container.CreateContainer(targetPath); err != nil {
		events <- Event{
			Key: "start_mailpit_containers",
			Status: "error",
			Message: "Failed to start Mailpit containers",
		}
		return err
	}

	events <- Event{
		Key: "mailpit_setup_completed",
		Name: "Mailpit setup completed",
		Status: "success",
		Message: "Mailpit setup completed successfully",
	}

	return nil
}