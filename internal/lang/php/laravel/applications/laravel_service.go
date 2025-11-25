package applications

import (
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

	if err := s.repository.CloneRepo(targetRepo, targetPath); err != nil {
		eventChan <- events.Event{
			Key:     "clone_laravel_repository",
			Name:    "Clone Laravel Repository",
			Status:  "error",
			Message: "Failed to clone repository: " + err.Error(),
		}

		return err
	}

	eventChan <- events.Event{
		Key:     "clone_laravel_repository",
		Name:    "Clone Laravel Repository",
		Status:  "success",
		Message: "Laravel repository cloned successfully",
	}

	eventChan <- events.Event{
		Key:     "set_up_environment_variables",
		Name:    "Set Up Environment Variables",
		Status:  "running",
		Message: "Setting up environment variables...",
	}

	if err := utils.CreateEnvFile(targetPath); err != nil {
		eventChan <- events.Event{
			Key:     "set_up_environment_variables",
			Name:    "Set Up Environment Variables",
			Status:  "error",
			Message: "Failed to set up environment variables: " + err.Error(),
		}

		return err
	}

	envFilePath := filepath.Join(targetPath, ".env")

	content, err := os.ReadFile(envFilePath)

	if err != nil {
		eventChan <- events.Event{
			Key:     "set_up_environment_variables",
			Name:    "Set Up Environment Variables",
			Status:  "error",
			Message: "Failed to read .env file: " + err.Error(),
		}

		return err
	}

	updateContent := string(content)

	replacements := map[string]any{
		"CONTAINER_NAME=": fmt.Sprintf("CONTAINER_NAME=%s", containerName),
		"REPOSITORY=":     "REPOSITORY=src",
		"DOCKER_PATH=":    "DOCKER_PATH=Infra/php",
		"VIRTUAL_HOST=":   fmt.Sprintf("VIRTUAL_HOST=%s", virtualHost),
		"TZ=":             fmt.Sprintf("TZ=%s", time.Now().Location().String()),
	}

	if err := utils.ReplaceAllValue(&updateContent, replacements); err != nil {
		eventChan <- events.Event{
			Key:     "set_up_environment_variables",
			Name:    "Set Up Environment Variables",
			Status:  "error",
			Message: "Failed to replace values in .env file: " + err.Error(),
		}

		return err
	}

	if err := os.WriteFile(envFilePath, []byte(updateContent), 0644); err != nil {
		eventChan <- events.Event{
			Key:     "set_up_environment_variables",
			Name:    "Set Up Environment Variables",
			Status:  "error",
			Message: "Failed to write .env file: " + err.Error(),
		}

		return err
	}

	eventChan <- events.Event{
		Key:     "set_up_environment_variables",
		Name:    "Set Up Environment Variables",
		Status:  "success",
		Message: "Environment variables set up successfully",
	}

	eventChan <- events.Event{
		Key:     "resolve_dependencies_container_booting",
		Name:    "Resolve Dependencies & Container Booting",
		Status:  "running",
		Message: "Resolving dependencies and booting container...",
	}

	if err := langutils.ResolveDependenciesContainerBooting(
		s.container,
		modules,
		s.config_service,
	); err != nil {
		eventChan <- events.Event{
			Key:     "resolve_dependencies_container_booting",
			Name:    "Resolve Dependencies & Container Booting",
			Status:  "error",
			Message: "Failed to resolve dependencies and boot container: " + err.Error(),
		}

		return err
	}

	eventChan <- events.Event{
		Key:     "resolve_dependencies_container_booting",
		Name:    "Resolve Dependencies & Container Booting",
		Status:  "success",
		Message: "Dependencies resolved and container booted successfully",
	}

	eventChan <- events.Event{
		Key:     "start_laravel_container",
		Name:    "Start Laravel Container",
		Status:  "running",
		Message: "Starting Laravel container...",
	}

	if err := s.container.CreateContainer(targetPath); err != nil {
		eventChan <- events.Event{
			Key:     "start_laravel_container",
			Name:    "Start Laravel Container",
			Status:  "error",
			Message: "Failed to start Laravel container: " + err.Error(),
		}

		return err
	}

	if _, err := s.container.ExecCommand(
		containerName,
		"laravel",
		"new",
		containerName,
		"--no-interaction",
		"--phpunit",
		"--database=sqlite",
	); err != nil {
		eventChan <- events.Event{
			Key:     "start_laravel_container",
			Name:    "Start Laravel Container",
			Status:  "error",
			Message: "Failed to create new Laravel application: " + err.Error(),
		}

		return err
	}

	replacements = map[string]any{
		"REPOSITORY=src": fmt.Sprintf("REPOSITORY=src/%s", containerName),
	}

	if err := utils.ReplaceAllValue(&updateContent, replacements); err != nil {
		eventChan <- events.Event{
			Key:     "start_laravel_container",
			Name:    "Start Laravel Container",
			Status:  "error",
			Message: "Failed to replace values in .env file: " + err.Error(),
		}

		return err
	}

	if err := os.WriteFile(envFilePath, []byte(updateContent), 0644); err != nil {
		eventChan <- events.Event{
			Key:     "start_laravel_container",
			Name:    "Start Laravel Container",
			Status:  "error",
			Message: "Failed to write .env file: " + err.Error(),
		}

		return err
	}

	if err := s.container.CreateContainer(targetPath); err != nil {
		eventChan <- events.Event{
			Key:     "start_laravel_container",
			Name:    "Start Laravel Container",
			Status:  "error",
			Message: "Failed to restart Laravel container: " + err.Error(),
		}

		return err
	}

	if _, err := s.container.ExecCommand(
		containerName,
		"composer",
		"install",
	); err != nil {
		eventChan <- events.Event{
			Key:     "start_laravel_container",
			Name:    "Start Laravel Container",
			Status:  "error",
			Message: "Failed to create new Laravel application: " + err.Error(),
		}

		return err
	}

	eventChan <- events.Event{
		Key:     "start_laravel_container",
		Name:    "Start Laravel Container",
		Status:  "success",
		Message: "Laravel container started successfully",
	}

	eventChan <- events.Event{
		Key:     "create_devcontainer_settings",
		Name:    "Create DevContainer Settings",
		Status:  "running",
		Message: "Creating DevContainer settings...",
	}

	devcontainerExamplePath := filepath.Join(targetPath, ".devcontainer", "devcontainer.json.example")

	if _, err := os.Stat(devcontainerExamplePath); os.IsNotExist(err) {
		eventChan <- events.Event{
			Key:     "create_devcontainer_settings",
			Name:    "Create DevContainer Settings",
			Status:  "error",
			Message: "DevContainer example file does not exist: " + devcontainerExamplePath,
		}

		return err
	}

	devcontainerPath := filepath.Join(targetPath, ".devcontainer", "devcontainer.json")

	if err := utils.CopyFile(devcontainerExamplePath, devcontainerPath); err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_settings",
			Name:    "Create DevContainer Settings",
			Status:  "error",
			Message: "Failed to create DevContainer settings: " + err.Error(),
		}

		return err
	}

	devContainerContents, err := os.ReadFile(devcontainerPath)

	if err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_settings",
			Name:    "Create DevContainer Settings",
			Status:  "error",
			Message: "Failed to read DevContainer file: " + err.Error(),
		}

		return err
	}

	updateDevContainerContents := string(devContainerContents)

	replacements = map[string]any{
		`"name": "project_repository",`: fmt.Sprintf(`"name": "%s",`, containerName),
	}

	if err := utils.ReplaceAllValue(&updateDevContainerContents, replacements); err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_settings",
			Name:    "Create DevContainer Settings",
			Status:  "error",
			Message: "Failed to replace values in DevContainer file: " + err.Error(),
		}

		return err
	}

	if err := os.WriteFile(devcontainerPath, []byte(updateDevContainerContents), 0644); err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_settings",
			Name:    "Create DevContainer Settings",
			Status:  "error",
			Message: "Failed to write DevContainer file: " + err.Error(),
		}

		return err
	}

	eventChan <- events.Event{
		Key:     "create_devcontainer_settings",
		Name:    "Create DevContainer Settings",
		Status:  "success",
		Message: "DevContainer settings created successfully",
	}

	eventChan <- events.Event{
		Key:     "laravel_setup_complete",
		Name:    "Laravel Setup Complete",
		Status:  "info",
		Message: "Laravel application setup is complete.",
	}

	return nil
}

func (s *LaravelService) Clone(
	eventChan chan<- events.Event,
	containerName string,
	virtualHost string,
	repoUrl string,
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
			"type": "clone",
			"repo": repoUrl,
		},
		Modules: modules,
	}

	if err := s.config_service.AddProject(
		projectConfig,
	); err != nil {
		eventChan <- events.Event{
			Key:     "clone_laravel_repository",
			Name:    "Clone Laravel Repository",
			Status:  "error",
			Message: "Failed to add project configuration: " + err.Error(),
		}

		return err
	}

	targetRepo := "https://github.com/takashiraki/docker_laravel.git"

	if err := s.repository.CloneRepo(
		targetRepo,
		targetPath,
	); err != nil {
		eventChan <- events.Event{
			Key:     "clone_laravel_repository",
			Name:    "Clone Laravel Repository",
			Status:  "error",
			Message: "Failed to clone repository: " + err.Error(),
		}

		return err
	}

	eventChan <- events.Event{
		Key:     "clone_laravel_repository",
		Name:    "Clone Laravel Repository",
		Status:  "success",
		Message: "Laravel repository cloned successfully",
	}

	eventChan <- events.Event{
		Key:     "set_up_environment_variables",
		Name:    "Set Up Environment Variables",
		Status:  "info",
		Message: "Please set up environment variables manually after cloning.",
	}

	if err := utils.CreateEnvFile(targetPath); err != nil {
		eventChan <- events.Event{
			Key:     "set_up_environment_variables",
			Name:    "Set Up Environment Variables",
			Status:  "error",
			Message: "Failed to set up environment variables: " + err.Error(),
		}

		return err
	}

	envFilePath := filepath.Join(targetPath, ".env")

	content, err := os.ReadFile(envFilePath)

	if err != nil {
		eventChan <- events.Event{
			Key:     "set_up_environment_variables",
			Name:    "Set Up Environment Variables",
			Status:  "error",
			Message: "Failed to read .env file: " + err.Error(),
		}

		return err
	}

	updateContent := string(content)

	replacements := map[string]any{
		"CONTAINER_NAME=": fmt.Sprintf("CONTAINER_NAME=%s", containerName),
		"REPOSITORY=":     fmt.Sprintf("REPOSITORY=src/%s", containerName),
		"DOCKER_PATH=":    "DOCKER_PATH=Infra/php",
		"VIRTUAL_HOST=":   fmt.Sprintf("VIRTUAL_HOST=%s", virtualHost),
		"TZ=":             fmt.Sprintf("TZ=%s", time.Now().Location().String()),
	}

	if err := utils.ReplaceAllValue(&updateContent, replacements); err != nil {
		eventChan <- events.Event{
			Key:     "set_up_environment_variables",
			Name:    "Set Up Environment Variables",
			Status:  "error",
			Message: "Failed to replace values in .env file: " + err.Error(),
		}

		return err
	}

	if err := os.WriteFile(envFilePath, []byte(updateContent), 0644); err != nil {
		eventChan <- events.Event{
			Key:     "set_up_environment_variables",
			Name:    "Set Up Environment Variables",
			Status:  "error",
			Message: "Failed to write .env file: " + err.Error(),
		}

		return err
	}

	eventChan <- events.Event{
		Key:     "set_up_environment_variables",
		Name:    "Set Up Environment Variables",
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
			Message: "DevContainer example file does not exist: " + devContainerExamplePath,
		}

		return err
	}

	devContainerPath := filepath.Join(targetPath, ".devcontainer", "devcontainer.json")

	if err := utils.CopyFile(devContainerExamplePath, devContainerPath); err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_settings",
			Name:    "Create devcontainer settings",
			Status:  "error",
			Message: "Failed to create devcontainer settings: " + err.Error(),
		}

		return err
	}

	devContainerContents, err := os.ReadFile(devContainerPath)

	if err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_settings",
			Name:    "Create devcontainer settings",
			Status:  "error",
			Message: "Failed to read devcontainer file: " + err.Error(),
		}

		return err
	}

	updateDevContainerContents := string(devContainerContents)

	replacements = map[string]any{
		`"name": "project_repository",`: fmt.Sprintf(`"name": "%s",`, containerName),
	}

	if err := utils.ReplaceAllValue(&updateDevContainerContents, replacements); err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_settings",
			Name:    "Create devcontainer settings",
			Status:  "error",
			Message: "Failed to replace values in devcontainer file: " + err.Error(),
		}

		return err
	}

	if err := os.WriteFile(devContainerPath, []byte(updateDevContainerContents), 0644); err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_settings",
			Name:    "Create devcontainer settings",
			Status:  "error",
			Message: "Failed to write devcontainer file: " + err.Error(),
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
		Key:     "resolve_dependencies_container_booting",
		Name:    "Resolve dependencies container booting",
		Status:  "running",
		Message: "Resolving dependencies container booting...",
	}

	if err := langutils.ResolveDependenciesContainerBooting(
		s.container,
		modules,
		s.config_service,
	); err != nil {
		eventChan <- events.Event{
			Key:     "resolve_dependencies_container_booting",
			Name:    "Resolve dependencies container booting",
			Status:  "error",
			Message: "Failed to resolve dependencies and boot container: " + err.Error(),
		}

		return err
	}

	eventChan <- events.Event{
		Key:     "resolve_dependencies_container_booting",
		Name:    "Resolve dependencies container booting",
		Status:  "success",
		Message: "Dependencies resolved and container booted successfully",
	}

	eventChan <- events.Event{
		Key:     "clone_project_repository",
		Name:    "Clone project repository",
		Status:  "running",
		Message: "Cloning project repository...",
	}

	srcPath := filepath.Join(targetPath, "src", containerName)

	if err := s.repository.CloneRepo(
		repoUrl,
		srcPath,
	); err != nil {
		eventChan <- events.Event{
			Key:     "clone_project_repository",
			Name:    "Clone project repository",
			Status:  "error",
			Message: "Failed to clone project repository: " + err.Error(),
		}

		return err
	}

	eventChan <- events.Event{
		Key:     "clone_project_repository",
		Name:    "Clone project repository",
		Status:  "success",
		Message: "Project repository cloned successfully",
	}

	eventChan <- events.Event{
		Key:     "start_laravel_container",
		Name:    "Start Laravel Container",
		Status:  "running",
		Message: "Starting Laravel container...",
	}

	if err := s.container.CreateContainer(targetPath); err != nil {
		eventChan <- events.Event{
			Key:     "start_laravel_container",
			Name:    "Start Laravel Container",
			Status:  "error",
			Message: "Failed to start Laravel container: " + err.Error(),
		}

		return err
	}

	if _, err := s.container.ExecCommand(
		containerName,
		"composer",
		"install",
	); err != nil {
		eventChan <- events.Event{
			Key:     "start_laravel_container",
			Name:    "Start Laravel Container",
			Status:  "error",
			Message: "Failed to install composer dependencies: " + err.Error(),
		}

		return err
	}

	if _, err := s.container.ExecCommand(
		containerName,
		"php",
		"-r",
		"file_exists('.env') || copy('.env.example', '.env');",
	); err != nil {
		eventChan <- events.Event{
			Key:     "start_laravel_container",
			Name:    "Start Laravel Container",
			Status:  "error",
			Message: "Failed to set up .env file: " + err.Error(),
		}

		return err
	}

	if _, err := s.container.ExecCommand(
		containerName,
		"php",
		"artisan",
		"key:generate",
	); err != nil {
		eventChan <- events.Event{
			Key:     "start_laravel_container",
			Name:    "Start Laravel Container",
			Status:  "error",
			Message: "Failed to generate application key: " + err.Error(),
		}

		return err
	}

	eventChan <- events.Event{
		Key:     "start_laravel_container",
		Name:    "Start Laravel Container",
		Status:  "success",
		Message: "Laravel container started successfully",
	}

	eventChan <- events.Event{
		Key:     "laravel_setup_complete",
		Name:    "Laravel Setup Complete",
		Status:  "info",
		Message: "Laravel application setup is complete.",
	}

	return nil
}
