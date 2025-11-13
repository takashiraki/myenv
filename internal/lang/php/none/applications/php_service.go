package applications

import (
	"errors"
	"fmt"
	"myenv/internal/config/application"
	"myenv/internal/events"
	"myenv/internal/infrastructure"
	langutils "myenv/internal/lang/utils"
	"myenv/internal/utils"
	"os"
	"path/filepath"
	"time"
)

type (
	PHPService struct {
		container infrastructure.ContainerInterface
		repository infrastructure.RepositoryInterface
		config_service application.ConfigService
	}
)

func NewPHPService(
	container infrastructure.ContainerInterface,
	repository infrastructure.RepositoryInterface,
	config_service application.ConfigService,
) *PHPService {
	return &PHPService{
		container: container,
		repository: repository,
		config_service: config_service,
	}
}

func (s *PHPService) Create(
	eventChan chan<- events.Event,
	containerName string,
	virtualHost string,
	modules []string,
) error {
	eventChan <- events.Event{
		Key: "clone_php_repository",
		Name: "Clone PHP Repository",
		Status: "running",
		Message: "Cloning PHP repository...",
	}

	homeDir, err := os.UserHomeDir()

	if err != nil {
		eventChan <- events.Event{
			Key: "clone_php_repository",
			Name: "Clone PHP Repository",
			Status: "error",
			Message: "Failed to get home directory",
		}
		return err
	}

	targetPath := filepath.Join(homeDir, "dev", containerName)

	if _, err := os.Stat(targetPath); err == nil {
		eventChan <- events.Event{
			Key: "clone_php_repository",
			Name: "Clone PHP Repository",
			Status: "error",
			Message: "Directory already exists",
		}
		return errors.New("directory already exists")
	}

	moduleConfig := application.Project{
		ContainerName: containerName,
		ContainerProxy: virtualHost,
		Path: targetPath,
		Lang: "php",
		Fw: "none",
		Options: map[string]string{
			"type": "new",
		},
		Modules: modules,
	}

	if err := s.config_service.AddProject(moduleConfig); err != nil {
		eventChan <- events.Event{
			Key: "clone_php_repository",
			Name: "Clone PHP Repository",
			Status: "error",
			Message: "Failed to add project to config",
		}
		return err
	}

	targetRepo := "https://github.com/takashiraki/docker_php.git"

	if err := s.repository.CloneRepo(targetRepo, targetPath); err != nil {
		eventChan <- events.Event{
			Key: "clone_php_repository",
			Name: "Clone PHP Repository",
			Status: "error",
			Message: "Failed to clone PHP repository",
		}
		return err
	}

	eventChan <- events.Event{
		Key: "clone_php_repository",
		Name: "Clone PHP Repository",
		Status: "success",
		Message: "PHP repository cloned successfully",
	}

	eventChan <- events.Event{
		Key: "set_up_environment_variables",
		Name: "Set up environment variables",
		Status: "running",
		Message: "Setting up environment variables...",
	}

	if err := utils.CreateEnvFile(targetPath); err != nil {
		eventChan <- events.Event{
			Key: "set_up_environment_variables",
			Name: "Set up environment variables",
			Status: "error",
			Message: "Failed to create .env file",
		}
		return err
	}

	envFilePath := filepath.Join(targetPath, ".env")

	content, err := os.ReadFile(envFilePath)

	if err != nil {
		eventChan <- events.Event{
			Key: "set_up_environment_variables",
			Name: "Set up environment variables",
			Status: "error",
			Message: "Failed to read .env file",
		}
		return err
	}

	updateContent := string(content)

	replacements := map[string]any{
		"REPOSITORY_PATH=": fmt.Sprintf("REPOSITORY_PATH=%s", "src"),
		"CONTAINER_NAME=":  fmt.Sprintf("CONTAINER_NAME=%s", containerName),
		"VIRTUAL_HOST=":       fmt.Sprintf("VIRTUAL_HOST=%s", virtualHost),
		"TZ=": fmt.Sprintf("TZ=%s", time.Now().Location().String()),
	}

	if err := utils.ReplaceAllValue(&updateContent, replacements); err != nil {
		eventChan <- events.Event{
			Key: "set_up_environment_variables",
			Name: "Set up environment variables",
			Status: "error",
			Message: "Failed to update .env file",
		}
		return err
	}

	if err := os.WriteFile(envFilePath, []byte(updateContent), 0644); err != nil {
		eventChan <- events.Event{
			Key: "set_up_environment_variables",
			Name: "Set up environment variables",
			Status: "error",
			Message: "Failed to write .env file",
		}
		return err
	}

	eventChan <- events.Event{
		Key: "set_up_environment_variables",
		Name: "Set up environment variables",
		Status: "success",
		Message: "Environment variables set up successfully",
	}

	eventChan <- events.Event{
		Key: "create_devcontainer_settings",
		Name: "Create devcontainer settings",
		Status: "running",
		Message: "Creating devcontainer settings...",
	}

	devContainerExamplePath := filepath.Join(targetPath, ".devcontainer","devcontainer.json.example")

	if _, err := os.Stat(devContainerExamplePath); os.IsNotExist(err) {
		eventChan <- events.Event{
			Key: "create_devcontainer_settings",
			Name: "Create devcontainer settings",
			Status: "error",
			Message: "devcontainer.json.example does not exist",
		}
		return err
	}

	devContainerPath := filepath.Join(targetPath,".devcontainer", "devcontainer.json")

	if err := utils.CopyFile(devContainerExamplePath, devContainerPath); err != nil {
		eventChan <- events.Event{
			Key: "create_devcontainer_settings",
			Name: "Create devcontainer settings",
			Status: "error",
			Message: "Failed to create devcontainer settings",
		}
		return err
	}

	devContainerContents, err := os.ReadFile(devContainerPath)

	if err != nil {
		eventChan <- events.Event{
			Key: "create_devcontainer_settings",
			Name: "Create devcontainer settings",
			Status: "error",
			Message: "Failed to read devcontainer.json",
		}
		return err
	}

	updateDevContainerContents := string(devContainerContents)

	replacements = map[string]any{
		`"name": "my php",`: fmt.Sprintf(`"name": "%s",`, containerName),
	}

	if err := utils.ReplaceAllValue(&updateDevContainerContents, replacements); err != nil {
		eventChan <- events.Event{
			Key: "create_devcontainer_settings",
			Name: "Create devcontainer settings",
			Status: "error",
			Message: "Failed to update devcontainer.json",
		}
		return err
	}

	if err := os.WriteFile(devContainerPath, []byte(updateDevContainerContents), 0644); err != nil {
		eventChan <- events.Event{
			Key: "create_devcontainer_settings",
			Name: "Create devcontainer settings",
			Status: "error",
			Message: "Failed to write devcontainer.json",
		}
		return err
	}

	eventChan <- events.Event{
		Key: "create_devcontainer_settings",
		Name: "Create devcontainer settings",
		Status: "success",
		Message: "Devcontainer settings created successfully",
	}

	eventChan <- events.Event{
		Key: "resolve_dependencies_container_booting",
		Name: "Resolve dependencies container booting",
		Status: "running",
		Message: "Resolving dependencies container booting...",
	}

	if err := langutils.ResolveDependenciesContainerBooting(s.container, modules, s.config_service); err != nil {
		eventChan <- events.Event{
			Key: "resolve_dependencies_container_booting",
			Name: "Resolve dependencies container booting",
			Status: "error",
			Message: "Failed to resolve dependencies container booting",
		}
		return err
	}

	eventChan <- events.Event{
		Key: "resolve_dependencies_container_booting",
		Name: "Resolve dependencies container booting",
		Status: "success",
		Message: "Resolved dependencies container booting successfully",
	}

	eventChan <- events.Event{
		Key: "start_php_containers",
		Name: "Start PHP containers",
		Status: "running",
		Message: "Starting PHP containers...",
	}

	if err := s.container.CreateContainer(targetPath); err != nil {
		eventChan <- events.Event{
			Key: "start_php_containers",
			Name: "Start PHP containers",
			Status: "error",
			Message: "Failed to start PHP containers",
		}
		return err
	}

	eventChan <- events.Event{
		Key: "start_php_containers",
		Name: "Start PHP containers",
		Status: "success",
		Message: "PHP containers started successfully",
	}

	eventChan <- events.Event{
		Key: "php_setup_completed",
		Name: "PHP setup completed",
		Status: "success",
		Message: "PHP setup completed successfully",
	}

	return nil
}

func (s *PHPService) Clone(
	eventChan chan<- events.Event,
	containerName string,
	virtualHost string,
	repoUrl string,
	modules []string,
) error {
	eventChan <- events.Event{
		Key: "clone_php_repository",
		Name: "Clone PHP Repository",
		Status: "running",
		Message: "Cloning PHP repository...",
	}

	homeDir, err := os.UserHomeDir()

	if err != nil {
		eventChan <- events.Event{
			Key: "clone_php_repository",
			Name: "Clone PHP Repository",
			Status: "error",
			Message: "Failed to get home directory",
		}
		return err
	}

	targetPath := filepath.Join(homeDir, "dev", containerName)

	if _, err := os.Stat(targetPath); err == nil {
		eventChan <- events.Event{
			Key: "clone_php_repository",
			Name: "Clone PHP Repository",
			Status: "error",
			Message: "Directory already exists",
		}
		return errors.New("directory already exists")
	}

	moduleConfig := application.Project{
		ContainerName: containerName,
		ContainerProxy: virtualHost,
		Path: targetPath,
		Lang: "php",
		Fw: "none",
		Options: map[string]string{
			"type": "clone",
			"repo": repoUrl,
		},
		Modules: modules,
	}

	if err := s.config_service.AddProject(moduleConfig); err != nil {
		eventChan <- events.Event{
			Key: "clone_php_repository",
			Name: "Clone PHP Repository",
			Status: "error",
			Message: "Failed to add project to config",
		}
		return err
	}

	targetRepo := "https://github.com/takashiraki/docker_php.git"

	if err := s.repository.CloneRepo(targetRepo, targetPath); err != nil {
		eventChan <- events.Event{
			Key: "clone_php_repository",
			Name: "Clone PHP Repository",
			Status: "error",
			Message: "Failed to clone PHP repository",
		}
		return err
	}

	eventChan <- events.Event{
		Key: "clone_php_repository",
		Name: "Clone PHP Repository",
		Status: "success",
		Message: "PHP repository cloned successfully",
	}

	eventChan <- events.Event{
		Key: "set_up_environment_variables",
		Name: "Set up environment variables",
		Status: "running",
		Message: "Setting up environment variables...",
	}

	if err := utils.CreateEnvFile(targetPath); err != nil {
		eventChan <- events.Event{
			Key: "set_up_environment_variables",
			Name: "Set up environment variables",
			Status: "error",
			Message: "Failed to create .env file",
		}
		return err
	}

	envFilePath := filepath.Join(targetPath, ".env")

	content, err := os.ReadFile(envFilePath)

	if err != nil {
		eventChan <- events.Event{
			Key: "set_up_environment_variables",
			Name: "Set up environment variables",
			Status: "error",
			Message: "Failed to read .env file",
		}
		return err
	}

	updateContent := string(content)

	replacements := map[string]any{
		"REPOSITORY_PATH=": fmt.Sprintf("REPOSITORY_PATH=%s", "src/" + containerName),
		"CONTAINER_NAME=":  fmt.Sprintf("CONTAINER_NAME=%s", containerName),
		"VIRTUAL_HOST=":       fmt.Sprintf("VIRTUAL_HOST=%s", virtualHost),
		"TZ=": fmt.Sprintf("TZ=%s", time.Now().Location().String()),
	}

	if err := utils.ReplaceAllValue(&updateContent, replacements); err != nil {
		eventChan <- events.Event{
			Key: "set_up_environment_variables",
			Name: "Set up environment variables",
			Status: "error",
			Message: "Failed to update .env file",
		}
		return err
	}

	if err := os.WriteFile(envFilePath, []byte(updateContent), 0644); err != nil {
		eventChan <- events.Event{
			Key: "set_up_environment_variables",
			Name: "Set up environment variables",
			Status: "error",
			Message: "Failed to write .env file",
		}
		return err
	}

	eventChan <- events.Event{
		Key: "set_up_environment_variables",
		Name: "Set up environment variables",
		Status: "success",
		Message: "Environment variables set up successfully",
	}

	eventChan <- events.Event{
		Key: "create_devcontainer_settings",
		Name: "Create devcontainer settings",
		Status: "running",
		Message: "Creating devcontainer settings...",
	}

	devContainerExamplePath := filepath.Join(targetPath, ".devcontainer","devcontainer.json.example")

	if _, err := os.Stat(devContainerExamplePath); os.IsNotExist(err) {
		eventChan <- events.Event{
			Key: "create_devcontainer_settings",
			Name: "Create devcontainer settings",
			Status: "error",
			Message: "devcontainer.json.example does not exist",
		}
		return err
	}

	devContainerPath := filepath.Join(targetPath,".devcontainer", "devcontainer.json")

	if err := utils.CopyFile(devContainerExamplePath, devContainerPath); err != nil {
		eventChan <- events.Event{
			Key: "create_devcontainer_settings",
			Name: "Create devcontainer settings",
			Status: "error",
			Message: "Failed to create devcontainer settings",
		}
		return err
	}

	devContainerContents, err := os.ReadFile(devContainerPath)

	if err != nil {
		eventChan <- events.Event{
			Key: "create_devcontainer_settings",
			Name: "Create devcontainer settings",
			Status: "error",
			Message: "Failed to read devcontainer.json",
		}
		return err
	}

	updateDevContainerContents := string(devContainerContents)

	replacements = map[string]any{
		`"name": "my php",`: fmt.Sprintf(`"name": "%s",`, containerName),
	}

	if err := utils.ReplaceAllValue(&updateDevContainerContents, replacements); err != nil {
		eventChan <- events.Event{
			Key: "create_devcontainer_settings",
			Name: "Create devcontainer settings",
			Status: "error",
			Message: "Failed to update devcontainer.json",
		}
		return err
	}

	if err := os.WriteFile(devContainerPath, []byte(updateDevContainerContents), 0644); err != nil {
		eventChan <- events.Event{
			Key: "create_devcontainer_settings",
			Name: "Create devcontainer settings",
			Status: "error",
			Message: "Failed to write devcontainer.json",
		}
		return err
	}

	eventChan <- events.Event{
		Key: "create_devcontainer_settings",
		Name: "Create devcontainer settings",
		Status: "success",
		Message: "Devcontainer settings created successfully",
	}

	eventChan <- events.Event{
		Key: "resolve_dependencies_container_booting",
		Name: "Resolve dependencies container booting",
		Status: "running",
		Message: "Resolving dependencies container booting...",
	}

	if err := langutils.ResolveDependenciesContainerBooting(s.container, modules, s.config_service); err != nil {
		eventChan <- events.Event{
			Key: "resolve_dependencies_container_booting",
			Name: "Resolve dependencies container booting",
			Status: "error",
			Message: "Failed to resolve dependencies container booting",
		}
		return err
	}

	eventChan <- events.Event{
		Key: "resolve_dependencies_container_booting",
		Name: "Resolve dependencies container booting",
		Status: "success",
		Message: "Resolved dependencies container booting successfully",
	}
	
	eventChan <- events.Event{
		Key: "clone_project_repository",
		Name: "Clone project repository",
		Status: "running",
		Message: "Cloning project repository...",
	}

	srcPath := filepath.Join(targetPath, "src", containerName)

	if err := s.repository.CloneRepo(repoUrl, srcPath); err != nil {
		eventChan <- events.Event{
			Key: "clone_project_repository",
			Name: "Clone project repository",
			Status: "error",
			Message: "Failed to clone project repository",
		}
		return err
	}

	eventChan <- events.Event{
		Key: "clone_project_repository",
		Name: "Clone project repository",
		Status: "success",
		Message: "Project repository cloned successfully",
	}

	eventChan <- events.Event{
		Key: "start_php_containers",
		Name: "Start PHP containers",
		Status: "running",
		Message: "Starting PHP containers...",
	}

	if err := s.container.CreateContainer(targetPath); err != nil {
		eventChan <- events.Event{
			Key: "start_php_containers",
			Name: "Start PHP containers",
			Status: "error",
			Message: "Failed to start PHP containers",
		}
		return err
	}

	eventChan <- events.Event{
		Key: "start_php_containers",
		Name: "Start PHP containers",
		Status: "success",
		Message: "PHP containers started successfully",
	}

	eventChan <- events.Event{
		Key: "php_setup_completed",
		Name: "PHP setup completed",
		Status: "success",
		Message: "PHP setup completed successfully",
	}

	return nil
}