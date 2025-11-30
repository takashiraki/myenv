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
)

type (
	NuxtService struct {
		container      infrastructure.ContainerInterface
		repository     infrastructure.RepositoryInterface
		config_service application.ConfigService
	}
)

func NewNuxtService(
	container infrastructure.ContainerInterface,
	repository infrastructure.RepositoryInterface,
	config_service application.ConfigService,
) *NuxtService {
	return &NuxtService{
		container:      container,
		repository:     repository,
		config_service: config_service,
	}
}

func (s *NuxtService) Create(
	containerName string,
	virtualHost string,
	framework string,
	eventChan chan<- events.Event,
	modules []string,
) error {

	eventChan <- events.Event{
		Key:     "clone_node_repository",
		Name:    "Clone Node Repository",
		Status:  "running",
		Message: "Cloning Node.js repository...",
	}

	homeDir, err := os.UserHomeDir()

	if err != nil {
		eventChan <- events.Event{
			Key:     "clone_node_repository",
			Name:    "Clone Node Repository",
			Status:  "error",
			Message: "Failed to get user home directory: " + err.Error(),
		}

		return err
	}

	targetPath := filepath.Join(homeDir, "dev", containerName)

	if _, err := os.Stat(targetPath); err == nil {
		eventChan <- events.Event{
			Key:     "clone_node_repository",
			Name:    "Clone Node Repository",
			Status:  "error",
			Message: "Target path already exists: " + targetPath,
		}

		return err
	}

	projectConfig := application.Project{
		ContainerName:  containerName,
		ContainerProxy: virtualHost,
		Path:           targetPath,
		Lang:           "node",
		Fw:             framework,
		Options: map[string]string{
			"type": "new",
		},
		Modules: modules,
	}

	if err := s.config_service.AddProject(projectConfig); err != nil {
		eventChan <- events.Event{
			Key:     "clone_node_repository",
			Name:    "Clone Node Repository",
			Status:  "error",
			Message: "Failed to add project configuration: " + err.Error(),
		}

		return err
	}

	targetRepo := "https://github.com/takashiraki/docker_nodejs.git"

	if err := s.repository.CloneRepo(targetRepo, targetPath); err != nil {
		eventChan <- events.Event{
			Key:     "clone_node_repository",
			Name:    "Clone Node Repository",
			Status:  "error",
			Message: "Failed to clone repository: " + err.Error(),
		}

		return err
	}

	eventChan <- events.Event{
		Key:     "clone_node_repository",
		Name:    "Clone Node Repository",
		Status:  "success",
		Message: "Node repository cloned successfully",
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
			Message: "Failed to create .env file: " + err.Error(),
		}
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
		"DOCKER_PATH=":    "DOCKER_PATH=Infra",
		"VIRTUAL_HOST=":   fmt.Sprintf("VIRTUAL_HOST=%s", virtualHost),
		"VIRTUAL_PORT=":   "VIRTUAL_PORT=3000",
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
		Key:     "start_nuxt_container",
		Name:    "Start Nuxt Container",
		Status:  "running",
		Message: "Starting Nuxt container...",
	}

	if err := s.container.CreateContainer(targetPath); err != nil {
		eventChan <- events.Event{
			Key:     "start_nuxt_container",
			Name:    "Start Nuxt Container",
			Status:  "error",
			Message: "Failed to start Nuxt container: " + err.Error(),
		}

		return err
	}

	if _, err := s.container.ExecCommand(
		containerName,
		"npm",
		"create",
		"nuxt@latest",
		containerName,
		"--",
		"--packageManager",
		"npm",
		"--no-gitInit",
		"--no-modules",
	); err != nil {
		eventChan <- events.Event{
			Key:     "start_nuxt_container",
			Name:    "Start Nuxt Container",
			Status:  "error",
			Message: "Failed to set up Nuxt application: " + err.Error(),
		}

		return err
	}

	replacements = map[string]any{
		"REPOSITORY=src": fmt.Sprintf("REPOSITORY=src/%s", containerName),
	}

	if err := utils.ReplaceAllValue(&updateContent, replacements); err != nil {
		eventChan <- events.Event{
			Key:     "start_nuxt_container",
			Name:    "Start Nuxt Container",
			Status:  "error",
			Message: "Failed to update .env file: " + err.Error(),
		}

		return err
	}

	if err := os.WriteFile(envFilePath, []byte(updateContent), 0644); err != nil {
		eventChan <- events.Event{
			Key:     "start_nuxt_container",
			Name:    "Start Nuxt Container",
			Status:  "error",
			Message: "Failed to write .env file: " + err.Error(),
		}

		return err
	}

	if err := s.container.CreateContainer(targetPath); err != nil {
		eventChan <- events.Event{
			Key:     "start_nuxt_container",
			Name:    "Start Nuxt Container",
			Status:  "error",
			Message: "Failed to restart Nuxt container: " + err.Error(),
		}

		return err
	}

	eventChan <- events.Event{
		Key:     "start_nuxt_container",
		Name:    "Start Nuxt Container",
		Status:  "success",
		Message: "Nuxt container started successfully",
	}

	devContainerExamplePath := filepath.Join(targetPath, ".devcontainer", "devcontainer.json.example")

	if _, err := os.Stat(devContainerExamplePath); os.IsNotExist(err) {
		eventChan <- events.Event{
			Key:     "create_devcontainer_file",
			Name:    "Create DevContainer File",
			Status:  "error",
			Message: "DevContainer example file does not exist",
		}

		return err
	}

	devContainerPath := filepath.Join(targetPath, ".devcontainer", "devcontainer.json")

	if err := utils.CopyFile(devContainerExamplePath, devContainerPath); err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_file",
			Name:    "Create DevContainer File",
			Status:  "error",
			Message: "Failed to create DevContainer file: " + err.Error(),
		}

		return err
	}

	devContainerContent, err := os.ReadFile(devContainerPath)

	if err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_file",
			Name:    "Create DevContainer File",
			Status:  "error",
			Message: "Failed to read DevContainer file: " + err.Error(),
		}

		return err
	}

	updateDevContainerContents := string(devContainerContent)

	replacements = map[string]any{
		`"name": "nodejs project",`: fmt.Sprintf(`"name": "%s",`, containerName),
	}

	if err := utils.ReplaceAllValue(&updateDevContainerContents, replacements); err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_file",
			Name:    "Create DevContainer File",
			Status:  "error",
			Message: "Failed to update DevContainer file: " + err.Error(),
		}

		return err
	}

	if err := os.WriteFile(devContainerPath, []byte(updateDevContainerContents), 0644); err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_file",
			Name:    "Create DevContainer File",
			Status:  "error",
			Message: "Failed to write DevContainer file: " + err.Error(),
		}

		return err
	}

	eventChan <- events.Event{
		Key:     "create_devcontainer_file",
		Name:    "Create DevContainer File",
		Status:  "success",
		Message: "DevContainer file created successfully",
	}

	eventChan <- events.Event{
		Key:     "nuxt_application_creation_complete",
		Name:    "Nuxt Application Creation Complete",
		Status:  "info",
		Message: "Nuxt application created successfully",
	}

	return nil
}

func (s *NuxtService) Clone(
	containerName string,
	virtualHost string,
	framework string,
	repoUrl string,
	eventChan chan<- events.Event,
	modules []string,
) error {

	eventChan <- events.Event{
		Key:     "clone_node_repository",
		Name:    "Clone Node Repository",
		Status:  "running",
		Message: "Cloning Node.js repository...",
	}

	homeDir, err := os.UserHomeDir()

	if err != nil {
		eventChan <- events.Event{
			Key:     "clone_node_repository",
			Name:    "Clone Node Repository",
			Status:  "error",
			Message: "Failed to get user home directory: " + err.Error(),
		}

		return err
	}

	targetPath := filepath.Join(homeDir, "dev", containerName)

	if _, err := os.Stat(targetPath); err == nil {
		eventChan <- events.Event{
			Key:     "clone_node_repository",
			Name:    "Clone Node Repository",
			Status:  "error",
			Message: "Target path already exists: " + targetPath,
		}

		return err
	}

	projectConfig := application.Project{
		ContainerName:  containerName,
		ContainerProxy: virtualHost,
		Path:           targetPath,
		Lang:           "node",
		Fw:             framework,
		Options: map[string]string{
			"type": "clone",
			"repo": repoUrl,
		},
		Modules: modules,
	}

	if err := s.config_service.AddProject(projectConfig); err != nil {
		eventChan <- events.Event{
			Key:     "clone_node_repository",
			Name:    "Clone Node Repository",
			Status:  "error",
			Message: "Failed to add project configuration: " + err.Error(),
		}

		return err
	}

	targetRepo := "https://github.com/takashiraki/docker_nodejs.git"

	if err := s.repository.CloneRepo(targetRepo, targetPath); err != nil {
		eventChan <- events.Event{
			Key:     "clone_node_repository",
			Name:    "Clone Node Repository",
			Status:  "error",
			Message: "Failed to clone repository: " + err.Error(),
		}

		return err
	}

	eventChan <- events.Event{
		Key:     "clone_node_repository",
		Name:    "Clone Node Repository",
		Status:  "success",
		Message: "Node repository cloned successfully",
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
			Message: "Failed to create .env file: " + err.Error(),
		}
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
		"DOCKER_PATH=":    "DOCKER_PATH=Infra",
		"VIRTUAL_HOST=":   fmt.Sprintf("VIRTUAL_HOST=%s", virtualHost),
		"VIRTUAL_PORT=":   "VIRTUAL_PORT=3000",
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
		Key:     "clone_project_repository",
		Name:    "Clone Project Repository",
		Status:  "running",
		Message: "Cloning project repository...",
	}

	srcPath := filepath.Join(targetPath, "src", containerName)

	if err := s.repository.CloneRepo(repoUrl, srcPath); err != nil {
		eventChan <- events.Event{
			Key:     "clone_project_repository",
			Name:    "Clone Project Repository",
			Status:  "error",
			Message: "Failed to clone project repository: " + err.Error(),
		}

		return err
	}

	eventChan <- events.Event{
		Key:     "clone_project_repository",
		Name:    "Clone Project Repository",
		Status:  "success",
		Message: "Project repository cloned successfully",
	}

	eventChan <- events.Event{
		Key:     "start_nuxt_container",
		Name:    "Start Nuxt Container",
		Status:  "running",
		Message: "Starting Nuxt container...",
	}

	if err := s.container.CreateContainer(targetPath); err != nil {
		eventChan <- events.Event{
			Key:     "start_nuxt_container",
			Name:    "Start Nuxt Container",
			Status:  "error",
			Message: "Failed to start Nuxt container: " + err.Error(),
		}

		return err
	}

	if _, err := s.container.ExecCommand(
		containerName,
		"npm",
		"install",
	); err != nil {
		eventChan <- events.Event{
			Key:     "start_nuxt_container",
			Name:    "Start Nuxt Container",
			Status:  "error",
			Message: "Failed to set up Nuxt application: " + err.Error(),
		}

		return err
	}

	eventChan <- events.Event{
		Key:     "start_nuxt_container",
		Name:    "Start Nuxt Container",
		Status:  "success",
		Message: "Nuxt container started successfully",
	}

	devContainerExamplePath := filepath.Join(targetPath, ".devcontainer", "devcontainer.json.example")

	if _, err := os.Stat(devContainerExamplePath); os.IsNotExist(err) {
		eventChan <- events.Event{
			Key:     "create_devcontainer_file",
			Name:    "Create DevContainer File",
			Status:  "error",
			Message: "DevContainer example file does not exist",
		}

		return err
	}

	devContainerPath := filepath.Join(targetPath, ".devcontainer", "devcontainer.json")

	if err := utils.CopyFile(devContainerExamplePath, devContainerPath); err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_file",
			Name:    "Create DevContainer File",
			Status:  "error",
			Message: "Failed to create DevContainer file: " + err.Error(),
		}

		return err
	}

	devContainerContent, err := os.ReadFile(devContainerPath)

	if err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_file",
			Name:    "Create DevContainer File",
			Status:  "error",
			Message: "Failed to read DevContainer file: " + err.Error(),
		}

		return err
	}

	updateDevContainerContents := string(devContainerContent)

	replacements = map[string]any{
		`"name": "nodejs project",`: fmt.Sprintf(`"name": "%s",`, containerName),
	}

	if err := utils.ReplaceAllValue(&updateDevContainerContents, replacements); err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_file",
			Name:    "Create DevContainer File",
			Status:  "error",
			Message: "Failed to update DevContainer file: " + err.Error(),
		}

		return err
	}

	if err := os.WriteFile(devContainerPath, []byte(updateDevContainerContents), 0644); err != nil {
		eventChan <- events.Event{
			Key:     "create_devcontainer_file",
			Name:    "Create DevContainer File",
			Status:  "error",
			Message: "Failed to write DevContainer file: " + err.Error(),
		}

		return err
	}

	eventChan <- events.Event{
		Key:     "create_devcontainer_file",
		Name:    "Create DevContainer File",
		Status:  "success",
		Message: "DevContainer file created successfully",
	}

	eventChan <- events.Event{
		Key:     "nuxt_application_creation_complete",
		Name:    "Nuxt Application Creation Complete",
		Status:  "info",
		Message: "Nuxt application created successfully",
	}

	return nil
}
