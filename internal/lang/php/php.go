package php

import (
	"fmt"
	"log"
	"myenv/internal/config"
	"myenv/internal/infrastructure"
	"myenv/internal/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

type PHPService struct {
	container            infrastructure.ContainerInterface
	repository_interface infrastructure.RepositoryInterface
}

func newPHPService(
	container infrastructure.ContainerInterface,
	repository infrastructure.RepositoryInterface) *PHPService {
	return &PHPService{
		container:            container,
		repository_interface: repository,
	}
}

type PHPProjectDetail struct {
	Name    string            `json:"container_name"`
	Port    int               `json:"container_port"`
	Path    string            `json:"path"`
	Lang    string            `json:"lang"`
	Fw      string            `json:"framework"`
	Options map[string]string `json:"options"`
}

type PHPProject struct {
	Project PHPProjectDetail `json:"project"`
}

func PHP() {
	utils.ClearTerminal()
	fmt.Println("PHP called")

	clonePrompt := &survey.Select{
		Message: "Do you want to clone repository of PHP project?",
		Options: []string{"Yes", "No"},
	}

	clone := ""

	if err := survey.AskOne(clonePrompt, &clone); err != nil {
		log.Fatal(err)
	}

	containerRepo := infrastructure.NewDockerContainer()
	gitRepo := infrastructure.NewGitRepository()
	newPHPService := newPHPService(containerRepo, gitRepo)

	switch clone {
	case "Yes":
		fmt.Println("Clone project")
		cloneProject(newPHPService)
	case "No":
		fmt.Println("Create project")
		createProject(newPHPService)
	}
}

func createProject(p *PHPService) {
	containerName := ""
	containerNamePrompt := &survey.Input{
		Message: "Enter the container name of PHP : ",
	}

	err := survey.AskOne(
		containerNamePrompt, &containerName,
		survey.WithValidator(survey.Required),
		survey.WithValidator(survey.MinLength(3)),
		survey.WithValidator(survey.MaxLength(20)),
		survey.WithValidator(utils.ValidateProjectName),
		survey.WithValidator(utils.ValidateDirectory),
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	containerPort := 0
	containerPortPrompt := &survey.Input{
		Message: "Enter the port of PHP : ",
	}

	portErr := survey.AskOne(
		containerPortPrompt, &containerPort,
		survey.WithValidator(survey.Required),
		survey.WithValidator(utils.ValidatePort),
	)

	if portErr != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", portErr)
		return
	}

	utils.ClearTerminal()

	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	path := filepath.Join(homeDir, "dev", containerName)

	fmt.Printf("\n")
	fmt.Printf("\033[33m📋 Configuration:\033[0m\n")
	fmt.Printf("   • Container name : %s\n", containerName)
	fmt.Printf("   • Clone path     : %s\n", path)
	fmt.Printf("   • Port           : %d\n", containerPort)
	fmt.Printf("   • Framework      : None\n")
	fmt.Printf("   • Language       : PHP\n\n")

	var confirmResult bool

	confirmPrompt := &survey.Confirm{
		Message: "Is it okay to start building the environment with this configuration?",
	}

	if err := survey.AskOne(confirmPrompt, &confirmResult); err != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	if !confirmResult {
		fmt.Println("\n\033[33mSetup cancelled.\033[0m Returning to configuration...")
		return
	}

	fw := "none"
	lang := "php"
	options := map[string]string{
		"type": "new",
	}

	createConfigFile(containerName, containerPort, path, lang, fw, options)

	targetRepo := "https://github.com/takashiraki/docker_php.git"

	done := make(chan bool)

	go utils.ShowLoadingIndicator("Cloning repository", done)

	// if err := utils.CloneRepo(targetRepo, path); err != nil {
	if err := p.repository_interface.CloneRepo(targetRepo, path); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to clone PHP project\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "Repository not found") ||
			strings.Contains(errMsg, "fatal: repository"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Possible causes:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Repository URL is incorrect\n")
			fmt.Fprintf(os.Stderr, "   • Repository is private and requires authentication\n")
			fmt.Fprintf(os.Stderr, "   • Repository does not exist\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Check the repository URL and try again\n\n")

		case strings.Contains(errMsg, "Could not resolve host") ||
			strings.Contains(errMsg, "network"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Network issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Check your internet connection\n")
			fmt.Fprintf(os.Stderr, "   • Verify DNS settings\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Check network andtry again\n\n")

		case strings.Contains(errMsg, "git: command not found") ||
			strings.Contains(errMsg, "executable file not found"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Git is not installed:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Ubuntu/Debian: \033[36msudo apt install git\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • macOS: \033[36mbrew install git\033[0m\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Install git and try again\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\nDetails: %v\n", err)
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Fix the issue above and try again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)

		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	done <- true

	fmt.Printf("\r\033[KCloning repository completed \033[32m✓\033[0m\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Creating .env file", done)

	if err := utils.CreateEnvFile(path); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to create .env file\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "permission denied"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Permission issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Check directory permissions: \033[36mls -la %s\033[0m\n", path)
			fmt.Fprintf(os.Stderr, "   • Fix ownership: \033[36msudo chown -R $USER %s\033[0m\n", path)
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Fix permissions and try again\n\n")

		case strings.Contains(errMsg, ".env.example") && strings.Contains(errMsg, "no such file"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Template file missing:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The cloned repository is missing .env.example\n")
			fmt.Fprintf(os.Stderr, "   This file is required as a template for .env creation\n\n")
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Check if the repository includes .env.example\n")
			fmt.Fprintf(os.Stderr, "   • Or manually create .env file in: %s\n\n", path)

		case strings.Contains(errMsg, "already exists"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 File already exists:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   .env file already exists from a previous run\n")
			fmt.Fprintf(os.Stderr, "   • View existing file: \033[36mcat %s/.env\033[0m\n", path)
			fmt.Fprintf(os.Stderr, "   • Or remove it: \033[36mrm %s/.env\033[0m\n\n", path)
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m Remove or backup the existing .env and try again\n\n")

		case strings.Contains(errMsg, "no space left"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Disk space issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Check disk space: \033[36mdf -h\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Clean Docker: \033[36mdocker system prune -a\033[0m\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Free up space and try again\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\nDetails: %v\n", err)
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Check the error above and try again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	done <- true

	fmt.Printf("\r\033[KCreating .env file completed \033[32m✓\033[0m\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Setup env file", done)

	envFilePath := filepath.Join(path, ".env")

	repositoryPath := "src"
	hostPort := 80

	content, err := os.ReadFile(envFilePath)

	if err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to read .env file\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "no such file"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 File not found:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The .env file doesn't exist\n")
			fmt.Fprintf(os.Stderr, "   Expected location: \033[36m%s\033[0m\n\n", envFilePath)
			fmt.Fprintf(os.Stderr, "\033[33m🤔 Why:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Previous step may have failed silently\n")
			fmt.Fprintf(os.Stderr, "   • File may have been manually deleted\n\n")
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m This shouldn't happen normally.\n")
			fmt.Fprintf(os.Stderr, "   Try running the setup again from the beginning\n\n")

		case strings.Contains(errMsg, "permission denied"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Permission issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You don't have permission to read this file\n\n")
			fmt.Fprintf(os.Stderr, "\033[33m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check file permissions:\n")
			fmt.Fprintf(os.Stderr, "   \033[36mls -la %s\033[0m\n\n", envFilePath)
			fmt.Fprintf(os.Stderr, "   Fix permissions:\n")
			fmt.Fprintf(os.Stderr, "   \033[36msudo chmod 644 %s\033[0m\n\n", envFilePath)
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m Fix permissions and try again\n\n")

		case strings.Contains(errMsg, "is a directory"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Unexpected directory:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   There's a directory named '.env' instead of a file\n\n")
			fmt.Fprintf(os.Stderr, "\033[33m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Remove the directory:\n")
			fmt.Fprintf(os.Stderr, "   \033[36mrm -rf %s\033[0m\n\n", envFilePath)
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m Remove the directory and try again\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\nDetails: %v\n\n", err)
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m Fix the issue above and try again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	updateContent := string(content)

	replacements := map[string]any{
		"REPOSITORY_PATH=": fmt.Sprintf("REPOSITORY_PATH=%s", repositoryPath),
		"CONTAINER_NAME=":  fmt.Sprintf("CONTAINER_NAME=%s", containerName),
		"HOST_PORT=":       fmt.Sprintf("HOST_PORT=%d", containerPort),
		"CONTAINER_PORT=":  fmt.Sprintf("CONTAINER_PORT=%d", hostPort),
	}

	if err := utils.ReplaceAllValue(&updateContent, replacements); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to update .env file\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "missing"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Template mismatch:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The .env file doesn't have the expected format\n\n")
			fmt.Fprintf(os.Stderr, "\033[33m🤔 Why:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The template might be missing placeholders like:\n")
			fmt.Fprintf(os.Stderr, "   • REPOSITORY_PATH=\n")
			fmt.Fprintf(os.Stderr, "   • CONTAINER_NAME=\n")
			fmt.Fprintf(os.Stderr, "   • HOST_PORT=\n")
			fmt.Fprintf(os.Stderr, "   • CONTAINER_PORT=\n\n")
			fmt.Fprintf(os.Stderr, "\033[33m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check the .env file:\n")
			fmt.Fprintf(os.Stderr, "   \033[36mcat %s\033[0m\n\n", envFilePath)
			fmt.Fprintf(os.Stderr, "   Option 1: Update your repository's .env.example\n")
			fmt.Fprintf(os.Stderr, "   Option 2: Manually edit the .env file with correct values\n\n")
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m Fix the template format and try again\n\n")

		case strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "empty"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Invalid configuration:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   One of the configuration values is invalid\n\n")
			fmt.Fprintf(os.Stderr, "   Container name: %s\n", containerName)
			fmt.Fprintf(os.Stderr, "   Container port: %d\n\n", containerPort)
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m This shouldn't happen normally.\n")
			fmt.Fprintf(os.Stderr, "   Please report this as a bug\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Unexpected error:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Details: %v\n\n", err)
			fmt.Fprintf(os.Stderr, "\033[33m🔧 Quick fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You can manually edit the .env file:\n")
			fmt.Fprintf(os.Stderr, "   \033[36m%s\033[0m\n\n", envFilePath)
			fmt.Fprintf(os.Stderr, "   Set these values:\n")
			fmt.Fprintf(os.Stderr, "   • REPOSITORY_PATH=%s\n", repositoryPath)
			fmt.Fprintf(os.Stderr, "   • CONTAINER_NAME=%s\n", containerName)
			fmt.Fprintf(os.Stderr, "   • HOST_PORT=%d\n", containerPort)
			fmt.Fprintf(os.Stderr, "   • CONTAINER_PORT=80\n\n")
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m After manual edit, continue with docker compose up\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	if err := os.WriteFile(envFilePath, []byte(updateContent), 0644); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to write .env file\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "permission denied"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Permission issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You don't have permission to write this file\n\n")
			fmt.Fprintf(os.Stderr, "\033[33m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check file permissions:\n")
			fmt.Fprintf(os.Stderr, "   \033[36mls -la %s\033[0m\n\n", envFilePath)
			fmt.Fprintf(os.Stderr, "   Fix permissions:\n")
			fmt.Fprintf(os.Stderr, "   \033[36msudo chmod 644 %s\033[0m\n\n", envFilePath)
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m Fix permissions and try again\n\n")

		case strings.Contains(errMsg, "no space left"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Disk space issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Your disk is full - no space left to write files\n\n")
			fmt.Fprintf(os.Stderr, "\033[33m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check disk space:\n")
			fmt.Fprintf(os.Stderr, "   \033[36mdf -h\033[0m\n\n")
			fmt.Fprintf(os.Stderr, "   Clean up Docker:\n")
			fmt.Fprintf(os.Stderr, "   \033[36mdocker system prune -a\033[0m\n\n")
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m Free up space and try again\n\n")

		case strings.Contains(errMsg, "read-only file system"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Read-only filesystem:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The filesystem is mounted as read-only\n\n")
			fmt.Fprintf(os.Stderr, "\033[33m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   This usually happens when the disk has errors\n")
			fmt.Fprintf(os.Stderr, "   Check system logs:\n")
			fmt.Fprintf(os.Stderr, "   \033[36mdmesg | grep -i error\033[0m\n\n")
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m System administrator help may be needed\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\nDetails: %v\n\n", err)
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m Fix the issue above and try again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	done <- true

	fmt.Printf("\r\033[KSetup .env file completed \033[32m✓\033[0m\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Starting Docker containers", done)

	if err := p.container.CreateContainer(path); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to start Docker containers\n")

		errMsg := err.Error()

		switch {
		case strings.Contains(errMsg, "Cannot connect to the Docker daemon"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 What happened:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Docker daemon is not running\n\n")

			fmt.Fprintf(os.Stderr, "\033[33m🔍 Why:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Docker Desktop is not started\n")
			fmt.Fprintf(os.Stderr, "   • Docker service is stopped\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Option 1: Start Docker Desktop\n\n")
			fmt.Fprintf(os.Stderr, "   Option 2: Start Docker service:\n")
			fmt.Fprintf(os.Stderr, "             $ \033[36msudo systemctl start docker\033[0m\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Start Docker and try again\n\n")

		case strings.Contains(errMsg, "port is already allocated") || strings.Contains(errMsg, "address already in use"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 What happened:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Port %d is already in use by another application\n\n", containerPort)

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Option 1: Find and stop the conflicting process:\n")
			fmt.Fprintf(os.Stderr, "             $ \033[36msudo lsof -i :%d\033[0m\n\n", containerPort)
			fmt.Fprintf(os.Stderr, "   Option 2: Use a different port when running this setup\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Free up the port and try again\n\n")

		case strings.Contains(errMsg, "permission denied"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 What happened:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You don't have permission to run Docker commands\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Add your user to the docker group:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36msudo usermod -aG docker $USER\033[0m\n\n")
			fmt.Fprintf(os.Stderr, "   Then log out and log back in, or run:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mnewgrp docker\033[0m\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix permissions and try again\n\n")

		case strings.Contains(errMsg, "no space left"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 What happened:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Your disk is full - no space left for Docker\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check disk space:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdf -h\033[0m\n\n")
			fmt.Fprintf(os.Stderr, "   Clean up Docker resources:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdocker system prune -a\033[0m\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Free up space and try again\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 What happened:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Docker containers failed to start\n\n")

			fmt.Fprintf(os.Stderr, "\033[33m📋 Error details:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   %v\n\n", err)

			fmt.Fprintf(os.Stderr, "\033[36m🔧 Troubleshooting:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check Docker status:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdocker ps\033[0m\n\n")
			fmt.Fprintf(os.Stderr, "   View container logs:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mcd %s && docker compose logs\033[0m\n\n", path)
			fmt.Fprintf(os.Stderr, "   Check .env file:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mcat %s/.env\033[0m\n\n", path)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Check the error details above and try again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	done <- true

	fmt.Printf("\r\033[KStarting Docker containers completed \033[32m✓\033[0m\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Creating container workspace", done)

	devContainerPath := filepath.Join(path, ".devcontainer", "devcontainer.json")

	devContainerExamplePath := filepath.Join(path, ".devcontainer", "devcontainer.json.example")

	if _, err := os.Stat(devContainerExamplePath); err != nil {
		done <- true

		if os.IsNotExist(err) {
			fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Template file not found\n")

			fmt.Fprintf(os.Stderr, "\n\033[33m💡 What happened:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The required file is missing:\n")
			fmt.Fprintf(os.Stderr, "   \033[36m%s\033[0m\n\n", devContainerExamplePath)

			fmt.Fprintf(os.Stderr, "\033[33m🔍 Why:\033[0m\n")

			// Check if .devcontainer directory exists
			devContainerDir := filepath.Join(path, ".devcontainer")
			if _, dirErr := os.Stat(devContainerDir); os.IsNotExist(dirErr) {
				fmt.Fprintf(os.Stderr, "   • The .devcontainer directory doesn't exist in this repository\n")
				fmt.Fprintf(os.Stderr, "   • This repository may not support devcontainer development\n\n")

				fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
				fmt.Fprintf(os.Stderr, "   Option 1: Check if you're using the correct repository\n")
				fmt.Fprintf(os.Stderr, "             Expected: https://github.com/takashiraki/docker_php.git\n\n")
				fmt.Fprintf(os.Stderr, "   Option 2: Manually create the .devcontainer directory:\n")
				fmt.Fprintf(os.Stderr, "             $ \033[36mmkdir -p %s\033[0m\n\n", devContainerDir)
			} else {
				fmt.Fprintf(os.Stderr, "   • The .devcontainer.json.example file is missing\n")
				fmt.Fprintf(os.Stderr, "   • The repository structure may be incomplete\n\n")

				fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
				fmt.Fprintf(os.Stderr, "   Check what files exist:\n")
				fmt.Fprintf(os.Stderr, "   $ \033[36mls -la %s\033[0m\n\n", devContainerDir)
				fmt.Fprintf(os.Stderr, "   Contact the repository maintainer about the missing file\n\n")
			}

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix the issue above and run this command again\n\n")

		} else if strings.Contains(err.Error(), "permission denied") {
			fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Permission denied\n")

			fmt.Fprintf(os.Stderr, "\n\033[33m💡 What happened:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You don't have permission to access this directory\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check directory permissions:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mls -la %s\033[0m\n\n", filepath.Dir(devContainerExamplePath))
			fmt.Fprintf(os.Stderr, "   Fix ownership:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36msudo chown -R $USER %s\033[0m\n\n", path)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix permissions and run this command again\n\n")

		} else {
			fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Cannot access file\n")

			fmt.Fprintf(os.Stderr, "\n\033[33m💡 What happened:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   An unexpected error occurred while checking for:\n")
			fmt.Fprintf(os.Stderr, "   \033[36m%s\033[0m\n\n", devContainerExamplePath)

			fmt.Fprintf(os.Stderr, "\033[33m📋 Error details:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   %v\n\n", err)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Try to fix the error above and run this command again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	if err := utils.CopyFile(devContainerExamplePath, devContainerPath); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to copy devcontainer.json file\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "permission denied"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Permission issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You don't have permission to copy files\n\n")
			fmt.Fprintf(os.Stderr, "\033[33m🔍 Why:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Can't read: %s\n", devContainerExamplePath)
			fmt.Fprintf(os.Stderr, "   • Or can't write to: %s\n\n", filepath.Dir(devContainerPath))

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check permissions:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mls -la %s\033[0m\n\n", filepath.Dir(devContainerPath))
			fmt.Fprintf(os.Stderr, "   Fix ownership:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36msudo chown -R $USER %s\033[0m\n\n", path)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix permissions and try again\n\n")

		case strings.Contains(errMsg, "no space left"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Disk space issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Your disk is full - no space left to create files\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check disk space:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdf -h\033[0m\n\n")
			fmt.Fprintf(os.Stderr, "   Clean up Docker:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdocker system prune -a\033[0m\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Free up space and try again\n\n")

		case strings.Contains(errMsg, "already exists"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 File already exists:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The devcontainer.json file already exists\n")
			fmt.Fprintf(os.Stderr, "   Location: \033[36m%s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[33m🔍 Why:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Previous run may have partially completed\n")
			fmt.Fprintf(os.Stderr, "   • File was manually created\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check the existing file:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mcat %s\033[0m\n\n", devContainerPath)
			fmt.Fprintf(os.Stderr, "   Remove it if needed:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mrm %s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Remove the file and try again\n\n")

		case strings.Contains(errMsg, "read-only file system"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Read-only filesystem:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The filesystem is mounted as read-only\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   This usually happens when the disk has errors\n")
			fmt.Fprintf(os.Stderr, "   Check system logs:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdmesg | grep -i error\033[0m\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m System administrator help may be needed\n\n")

		case strings.Contains(errMsg, "no such file"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Source file disappeared:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The template file existed but is now missing:\n")
			fmt.Fprintf(os.Stderr, "   \033[36m%s\033[0m\n\n", devContainerExamplePath)

			fmt.Fprintf(os.Stderr, "\033[33m🔍 Why:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • File was deleted between checks\n")
			fmt.Fprintf(os.Stderr, "   • Filesystem issue\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m This shouldn't happen normally. Try running again\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Unexpected error:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Failed to copy file:\n")
			fmt.Fprintf(os.Stderr, "   From: \033[36m%s\033[0m\n", devContainerExamplePath)
			fmt.Fprintf(os.Stderr, "   To:   \033[36m%s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[33m📋 Error details:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   %v\n\n", err)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Try to fix the error above and run again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	devContainerContents, err := os.ReadFile(devContainerPath)

	if err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to read devcontainer.json file\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "no such file"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 File disappeared:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The file was just created but is now missing:\n")
			fmt.Fprintf(os.Stderr, "   \033[36m%s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[33m🔍 Why:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Copy operation may have failed silently\n")
			fmt.Fprintf(os.Stderr, "   • File was deleted immediately after creation\n")
			fmt.Fprintf(os.Stderr, "   • Filesystem issue\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m This shouldn't happen normally. Try running again\n\n")

		case strings.Contains(errMsg, "permission denied"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Permission issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You don't have permission to read this file\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check file permissions:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mls -la %s\033[0m\n\n", devContainerPath)
			fmt.Fprintf(os.Stderr, "   Fix permissions:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36msudo chmod 644 %s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix permissions and try again\n\n")

		case strings.Contains(errMsg, "is a directory"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Unexpected directory:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   There's a directory named 'devcontainer.json' instead of a file\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Remove the directory:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mrm -rf %s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Remove the directory and try again\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\n\033[33m📋 Error details:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   %v\n\n", err)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix the issue above and try again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	updateDevContainerContents := string(devContainerContents)

	replacements = map[string]any{
		`"name": "php debug",`: fmt.Sprintf(`"name": "%s",`, containerName),
	}

	if err := utils.ReplaceAllValue(&updateDevContainerContents, replacements); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to update devcontainer.json file\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "missing"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Template mismatch:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The devcontainer.json doesn't have the expected format\n\n")

			fmt.Fprintf(os.Stderr, "\033[33m🔍 Why:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Looking for: \033[36m\"name\": \"php debug\",\033[0m\n")
			fmt.Fprintf(os.Stderr, "   But it doesn't exist in the file\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check the file contents:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mcat %s\033[0m\n\n", devContainerPath)
			fmt.Fprintf(os.Stderr, "   Option 1: Update the repository's devcontainer.json.example\n")
			fmt.Fprintf(os.Stderr, "   Option 2: Manually edit the devcontainer.json:\n")
			fmt.Fprintf(os.Stderr, "             Change the \"name\" field to: \033[36m\"%s\"\033[0m\n\n", containerName)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix the template format and try again\n\n")

		case strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "empty"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Invalid configuration:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The container name is invalid or empty\n")
			fmt.Fprintf(os.Stderr, "   Container name: \033[36m%s\033[0m\n\n", containerName)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m This shouldn't happen. Please report this as a bug\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Unexpected error:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Failed to update the devcontainer.json file\n\n")

			fmt.Fprintf(os.Stderr, "\033[33m📋 Error details:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   %v\n\n", err)

			fmt.Fprintf(os.Stderr, "\033[36m🔧 Quick fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You can manually edit the file:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36m%s\033[0m\n\n", devContainerPath)
			fmt.Fprintf(os.Stderr, "   Change \"name\" to: \033[36m\"%s\"\033[0m\n\n", containerName)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m After manual edit, continue setup\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	if err := os.WriteFile(devContainerPath, []byte(updateDevContainerContents), 0644); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to write devcontainer.json file\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "permission denied"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Permission issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You don't have permission to write this file\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check file permissions:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mls -la %s\033[0m\n\n", devContainerPath)
			fmt.Fprintf(os.Stderr, "   Fix permissions:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36msudo chmod 644 %s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix permissions and try again\n\n")

		case strings.Contains(errMsg, "no space left"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Disk space issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Your disk is full - no space left to write files\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check disk space:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdf -h\033[0m\n\n")
			fmt.Fprintf(os.Stderr, "   Clean up Docker:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdocker system prune -a\033[0m\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Free up space and try again\n\n")

		case strings.Contains(errMsg, "read-only file system"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Read-only filesystem:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The filesystem is mounted as read-only\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   This usually happens when the disk has errors\n")
			fmt.Fprintf(os.Stderr, "   Check system logs:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdmesg | grep -i error\033[0m\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m System administrator help may be needed\n\n")

		case strings.Contains(errMsg, "is a directory"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Path is a directory:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The path exists but it's a directory, not a file:\n")
			fmt.Fprintf(os.Stderr, "   \033[36m%s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Remove the directory:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mrm -rf %s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Remove the directory and try again\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\n\033[33m📋 Error details:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   %v\n\n", err)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix the issue above and try again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	done <- true

	fmt.Printf("\r\033[KCreating container workspace completed \033[32m✓\033[0m\n")

	utils.ClearTerminal()

	fmt.Printf("\n")
	fmt.Printf("\033[32m✓ Setup Complete!\033[0m 🎉\n\n")

	fmt.Printf("\033[33m📋 Configuration:\033[0m\n")
	fmt.Printf("   • Container Name : %s\n", containerName)
	fmt.Printf("   • Repository Path: %s\n", path)
	fmt.Printf("   • Port          : %d\n\n", containerPort)

	fmt.Printf("\033[36m🚀 Next steps:\033[0m\n")
	fmt.Printf("   1. Open VS Code:\n")
	fmt.Printf("      $ \033[36mcode %s\033[0m\n\n", path)
	fmt.Printf("   2. Access your application:\n")
	fmt.Printf("      🌐 \033[36mhttp://localhost:%d\033[0m\n\n", containerPort)
	fmt.Printf("   3. Start coding in the devcontainer!\n\n")

	codeVersionCommand := exec.Command("code", "--version")

	if _, err = codeVersionCommand.CombinedOutput(); err == nil {

		var openInVSCode bool

		openInVSCodePrompt := &survey.Confirm{
			Message: "Do you want to open the project in VS Code?",
		}

		if err := survey.AskOne(openInVSCodePrompt, &openInVSCode); err != nil {
			fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", err)
			return
		}

		if openInVSCode {
			openCommand := exec.Command("code", path)

			if _, err := openCommand.CombinedOutput(); err != nil {
				fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m Failed to open project in VS Code: %v\n", err)
				return
			}
		}
	}
}

func createConfigFile(containerName string, containerPort int, path string, lang string, fw string, options map[string]string) {
	err := config.AddProjectConfig(containerName, containerPort, path, lang, fw, options)

	if err != nil {
		log.Fatal(err)
	}
}

func cloneProject(p *PHPService) {
	gitRepo := ""
	gitRepoPrompt := &survey.Input{
		Message: "Enter the Git repository URL of PHP project : ",
	}

	err := survey.AskOne(
		gitRepoPrompt, &gitRepo,
		survey.WithValidator(survey.Required),
		survey.WithValidator(utils.ValidateGitRepoUrl),
		survey.WithValidator(utils.ValidateGitRepoProjectExists),
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	containerPort := 0
	containerPortPrompt := &survey.Input{
		Message: "Enter the port of PHP : ",
	}

	portErr := survey.AskOne(
		containerPortPrompt, &containerPort,
		survey.WithValidator(survey.Required),
		survey.WithValidator(utils.ValidatePort),
	)

	if portErr != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", portErr)
		return
	}

	utils.ClearTerminal()

	repoName := utils.ExtractionRepoName(gitRepo)

	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	path := filepath.Join(homeDir, "dev", repoName)
	containerName := repoName

	fmt.Printf("\n")
	fmt.Printf("\033[33m📋 Configuration:\033[0m\n")
	fmt.Printf("   • Container name : %s\n", containerName)
	fmt.Printf("   • Clone path     : %s\n", path)
	fmt.Printf("   • Port           : %d\n", containerPort)
	fmt.Printf("   • Framework      : None\n")
	fmt.Printf("   • Language       : PHP\n\n")

	var confirmResult bool

	confirmPrompt := &survey.Confirm{
		Message: "Is it okay to start building the environment with this configuration?",
	}

	if err := survey.AskOne(confirmPrompt, &confirmResult); err != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	if !confirmResult {
		fmt.Println("\n\033[33mSetup cancelled.\033[0m Returning to configuration...")
		return
	}

	fw := "none"
	lang := "php"
	options := map[string]string{
		"type": "clone",
		"repo": gitRepo,
	}

	createConfigFile(
		containerName,
		containerPort,
		path,
		lang,
		fw,
		options,
	)

	targetRepo := "https://github.com/takashiraki/docker_php.git"

	done := make(chan bool)

	go utils.ShowLoadingIndicator("Cloning repository", done)

	// if err := utils.CloneRepo(targetRepo, path); err != nil {
	if err := p.repository_interface.CloneRepo(targetRepo, path); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to clone PHP project\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "Repository not found") ||
			strings.Contains(errMsg, "fatal: repository"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Possible causes:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Repository URL is incorrect\n")
			fmt.Fprintf(os.Stderr, "   • Repository is private and requires authentication\n")
			fmt.Fprintf(os.Stderr, "   • Repository does not exist\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Check the repository URL and try again\n\n")

		case strings.Contains(errMsg, "Could not resolve host") ||
			strings.Contains(errMsg, "network"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Network issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Check your internet connection\n")
			fmt.Fprintf(os.Stderr, "   • Verify DNS settings\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Check network andtry again\n\n")

		case strings.Contains(errMsg, "git: command not found") ||
			strings.Contains(errMsg, "executable file not found"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Git is not installed:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Ubuntu/Debian: \033[36msudo apt install git\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • macOS: \033[36mbrew install git\033[0m\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Install git and try again\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\nDetails: %v\n", err)
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Fix the issue above and try again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)

		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	done <- true

	fmt.Printf("\r\033[KCloning repository completed \033[32m✓\033[0m\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Creating .env file", done)

	if err := utils.CreateEnvFile(path); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to create .env file\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "permission denied"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Permission issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Check directory permissions: \033[36mls -la %s\033[0m\n", path)
			fmt.Fprintf(os.Stderr, "   • Fix ownership: \033[36msudo chown -R $USER %s\033[0m\n", path)
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Fix permissions and try again\n\n")

		case strings.Contains(errMsg, ".env.example") && strings.Contains(errMsg, "no such file"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Template file missing:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The cloned repository is missing .env.example\n")
			fmt.Fprintf(os.Stderr, "   This file is required as a template for .env creation\n\n")
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Check if the repository includes .env.example\n")
			fmt.Fprintf(os.Stderr, "   • Or manually create .env file in: %s\n\n", path)

		case strings.Contains(errMsg, "already exists"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 File already exists:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   .env file already exists from a previous run\n")
			fmt.Fprintf(os.Stderr, "   • View existing file: \033[36mcat %s/.env\033[0m\n", path)
			fmt.Fprintf(os.Stderr, "   • Or remove it: \033[36mrm %s/.env\033[0m\n\n", path)
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m Remove or backup the existing .env and try again\n\n")

		case strings.Contains(errMsg, "no space left"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Disk space issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Check disk space: \033[36mdf -h\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Clean Docker: \033[36mdocker system prune -a\033[0m\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Free up space and try again\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\nDetails: %v\n", err)
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Check the error above and try again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	done <- true

	fmt.Printf("\r\033[KCreating .env file completed \033[32m✓\033[0m\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Clone your PHP project.", done)

	srcTargetPath := filepath.Join(path, "src", containerName)

	// if err := utils.CloneRepo(gitRepo, srcTargetPath); err != nil {
	if err := p.repository_interface.CloneRepo(gitRepo, srcTargetPath); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to clone PHP project\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "Repository not found") ||
			strings.Contains(errMsg, "fatal: repository"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Possible causes:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Repository URL is incorrect\n")
			fmt.Fprintf(os.Stderr, "   • Repository is private and requires authentication\n")
			fmt.Fprintf(os.Stderr, "   • Repository does not exist\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Check the repository URL and try again\n\n")

		case strings.Contains(errMsg, "Authentication failed") ||
			strings.Contains(errMsg, "Permission denied"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Authentication required:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Set up SSH keys: ssh-keygen && cat ~/.ssh/id_rsa.pub\n")
			fmt.Fprintf(os.Stderr, "   • Or use HTTPS with personal access token\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Configure git credentials and try again\n\n")

		case strings.Contains(errMsg, "Could not resolve host") ||
			strings.Contains(errMsg, "network"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Network issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Check your internet connection\n")
			fmt.Fprintf(os.Stderr, "   • Verify DNS settings\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Check network andtry again\n\n")

		case strings.Contains(errMsg, "git: command not found") ||
			strings.Contains(errMsg, "executable file not found"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Git is not installed:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Ubuntu/Debian: \033[36msudo apt install git\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • macOS: \033[36mbrew install git\033[0m\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Install git and try again\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\nDetails: %v\n", err)
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Fix the issue above and try again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)

		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	done <- true

	fmt.Printf("\r\033[KCloning your PHP project completed \033[32m✓\033[0m\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Setup env file", done)

	envFilePath := filepath.Join(path, ".env")

	repositoryPath := fmt.Sprintf("src/%s", containerName)
	hostPort := 80

	content, err := os.ReadFile(envFilePath)

	if err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to read .env file\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "no such file"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 File not found:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The .env file doesn't exist\n")
			fmt.Fprintf(os.Stderr, "   Expected location: \033[36m%s\033[0m\n\n", envFilePath)
			fmt.Fprintf(os.Stderr, "\033[33m🤔 Why:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Previous step may have failed silently\n")
			fmt.Fprintf(os.Stderr, "   • File may have been manually deleted\n\n")
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m This shouldn't happen normally.\n")
			fmt.Fprintf(os.Stderr, "   Try running the setup again from the beginning\n\n")

		case strings.Contains(errMsg, "permission denied"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Permission issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You don't have permission to read this file\n\n")
			fmt.Fprintf(os.Stderr, "\033[33m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check file permissions:\n")
			fmt.Fprintf(os.Stderr, "   \033[36mls -la %s\033[0m\n\n", envFilePath)
			fmt.Fprintf(os.Stderr, "   Fix permissions:\n")
			fmt.Fprintf(os.Stderr, "   \033[36msudo chmod 644 %s\033[0m\n\n", envFilePath)
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m Fix permissions and try again\n\n")

		case strings.Contains(errMsg, "is a directory"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Unexpected directory:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   There's a directory named '.env' instead of a file\n\n")
			fmt.Fprintf(os.Stderr, "\033[33m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Remove the directory:\n")
			fmt.Fprintf(os.Stderr, "   \033[36mrm -rf %s\033[0m\n\n", envFilePath)
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m Remove the directory and try again\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\nDetails: %v\n\n", err)
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m Fix the issue above and try again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	updateContent := string(content)

	replacements := map[string]interface{}{
		"REPOSITORY_PATH=": fmt.Sprintf("REPOSITORY_PATH=%s", repositoryPath),
		"CONTAINER_NAME=":  fmt.Sprintf("CONTAINER_NAME=%s", containerName),
		"HOST_PORT=":       fmt.Sprintf("HOST_PORT=%d", containerPort),
		"CONTAINER_PORT=":  fmt.Sprintf("CONTAINER_PORT=%d", hostPort),
	}

	if err := utils.ReplaceAllValue(&updateContent, replacements); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to update .env file\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "missing"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Template mismatch:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The .env file doesn't have the expected format\n\n")
			fmt.Fprintf(os.Stderr, "\033[33m🤔 Why:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The template might be missing placeholders like:\n")
			fmt.Fprintf(os.Stderr, "   • REPOSITORY_PATH=\n")
			fmt.Fprintf(os.Stderr, "   • CONTAINER_NAME=\n")
			fmt.Fprintf(os.Stderr, "   • HOST_PORT=\n")
			fmt.Fprintf(os.Stderr, "   • CONTAINER_PORT=\n\n")
			fmt.Fprintf(os.Stderr, "\033[33m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check the .env file:\n")
			fmt.Fprintf(os.Stderr, "   \033[36mcat %s\033[0m\n\n", envFilePath)
			fmt.Fprintf(os.Stderr, "   Option 1: Update your repository's .env.example\n")
			fmt.Fprintf(os.Stderr, "   Option 2: Manually edit the .env file with correct values\n\n")
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m Fix the template format and try again\n\n")

		case strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "empty"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Invalid configuration:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   One of the configuration values is invalid\n\n")
			fmt.Fprintf(os.Stderr, "   Container name: %s\n", containerName)
			fmt.Fprintf(os.Stderr, "   Container port: %d\n\n", containerPort)
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m This shouldn't happen normally.\n")
			fmt.Fprintf(os.Stderr, "   Please report this as a bug\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Unexpected error:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Details: %v\n\n", err)
			fmt.Fprintf(os.Stderr, "\033[33m🔧 Quick fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You can manually edit the .env file:\n")
			fmt.Fprintf(os.Stderr, "   \033[36m%s\033[0m\n\n", envFilePath)
			fmt.Fprintf(os.Stderr, "   Set these values:\n")
			fmt.Fprintf(os.Stderr, "   • REPOSITORY_PATH=src/%s\n", containerName)
			fmt.Fprintf(os.Stderr, "   • CONTAINER_NAME=%s\n", containerName)
			fmt.Fprintf(os.Stderr, "   • HOST_PORT=%d\n", containerPort)
			fmt.Fprintf(os.Stderr, "   • CONTAINER_PORT=80\n\n")
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m After manual edit, continue with docker compose up\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	if err := os.WriteFile(envFilePath, []byte(updateContent), 0644); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to write .env file\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "permission denied"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Permission issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You don't have permission to write this file\n\n")
			fmt.Fprintf(os.Stderr, "\033[33m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check file permissions:\n")
			fmt.Fprintf(os.Stderr, "   \033[36mls -la %s\033[0m\n\n", envFilePath)
			fmt.Fprintf(os.Stderr, "   Fix permissions:\n")
			fmt.Fprintf(os.Stderr, "   \033[36msudo chmod 644 %s\033[0m\n\n", envFilePath)
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m Fix permissions and try again\n\n")

		case strings.Contains(errMsg, "no space left"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Disk space issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Your disk is full - no space left to write files\n\n")
			fmt.Fprintf(os.Stderr, "\033[33m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check disk space:\n")
			fmt.Fprintf(os.Stderr, "   \033[36mdf -h\033[0m\n\n")
			fmt.Fprintf(os.Stderr, "   Clean up Docker:\n")
			fmt.Fprintf(os.Stderr, "   \033[36mdocker system prune -a\033[0m\n\n")
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m Free up space and try again\n\n")

		case strings.Contains(errMsg, "read-only file system"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Read-only filesystem:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The filesystem is mounted as read-only\n\n")
			fmt.Fprintf(os.Stderr, "\033[33m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   This usually happens when the disk has errors\n")
			fmt.Fprintf(os.Stderr, "   Check system logs:\n")
			fmt.Fprintf(os.Stderr, "   \033[36mdmesg | grep -i error\033[0m\n\n")
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m System administrator help may be needed\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\nDetails: %v\n\n", err)
			fmt.Fprintf(os.Stderr, "\033[36m→ Next steps:\033[0m Fix the issue above and try again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	done <- true

	fmt.Printf("\r\033[KSetup .env file completed \033[32m✓\033[0m\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Creating container workspace", done)

	devContainerPath := filepath.Join(path, ".devcontainer", "devcontainer.json")

	devContainerExamplePath := filepath.Join(path, ".devcontainer", "devcontainer.json.example")

	if _, err := os.Stat(devContainerExamplePath); err != nil {
		done <- true

		if os.IsNotExist(err) {
			fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Template file not found\n")

			fmt.Fprintf(os.Stderr, "\n\033[33m💡 What happened:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The required file is missing:\n")
			fmt.Fprintf(os.Stderr, "   \033[36m%s\033[0m\n\n", devContainerExamplePath)

			fmt.Fprintf(os.Stderr, "\033[33m🔍 Why:\033[0m\n")

			// Check if .devcontainer directory exists
			devContainerDir := filepath.Join(path, ".devcontainer")
			if _, dirErr := os.Stat(devContainerDir); os.IsNotExist(dirErr) {
				fmt.Fprintf(os.Stderr, "   • The .devcontainer directory doesn't exist in this repository\n")
				fmt.Fprintf(os.Stderr, "   • This repository may not support devcontainer development\n\n")

				fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
				fmt.Fprintf(os.Stderr, "   Option 1: Check if you're using the correct repository\n")
				fmt.Fprintf(os.Stderr, "             Expected: https://github.com/takashiraki/docker_php.git\n\n")
				fmt.Fprintf(os.Stderr, "   Option 2: Manually create the .devcontainer directory:\n")
				fmt.Fprintf(os.Stderr, "             $ \033[36mmkdir -p %s\033[0m\n\n", devContainerDir)
			} else {
				fmt.Fprintf(os.Stderr, "   • The .devcontainer.json.example file is missing\n")
				fmt.Fprintf(os.Stderr, "   • The repository structure may be incomplete\n\n")

				fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
				fmt.Fprintf(os.Stderr, "   Check what files exist:\n")
				fmt.Fprintf(os.Stderr, "   $ \033[36mls -la %s\033[0m\n\n", devContainerDir)
				fmt.Fprintf(os.Stderr, "   Contact the repository maintainer about the missing file\n\n")
			}

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix the issue above and run this command again\n\n")

		} else if strings.Contains(err.Error(), "permission denied") {
			fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Permission denied\n")

			fmt.Fprintf(os.Stderr, "\n\033[33m💡 What happened:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You don't have permission to access this directory\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check directory permissions:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mls -la %s\033[0m\n\n", filepath.Dir(devContainerExamplePath))
			fmt.Fprintf(os.Stderr, "   Fix ownership:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36msudo chown -R $USER %s\033[0m\n\n", path)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix permissions and run this command again\n\n")

		} else {
			fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Cannot access file\n")

			fmt.Fprintf(os.Stderr, "\n\033[33m💡 What happened:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   An unexpected error occurred while checking for:\n")
			fmt.Fprintf(os.Stderr, "   \033[36m%s\033[0m\n\n", devContainerExamplePath)

			fmt.Fprintf(os.Stderr, "\033[33m📋 Error details:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   %v\n\n", err)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Try to fix the error above and run this command again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	if err := utils.CopyFile(devContainerExamplePath, devContainerPath); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to copy devcontainer.json file\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "permission denied"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Permission issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You don't have permission to copy files\n\n")
			fmt.Fprintf(os.Stderr, "\033[33m🔍 Why:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Can't read: %s\n", devContainerExamplePath)
			fmt.Fprintf(os.Stderr, "   • Or can't write to: %s\n\n", filepath.Dir(devContainerPath))

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check permissions:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mls -la %s\033[0m\n\n", filepath.Dir(devContainerPath))
			fmt.Fprintf(os.Stderr, "   Fix ownership:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36msudo chown -R $USER %s\033[0m\n\n", path)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix permissions and try again\n\n")

		case strings.Contains(errMsg, "no space left"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Disk space issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Your disk is full - no space left to create files\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check disk space:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdf -h\033[0m\n\n")
			fmt.Fprintf(os.Stderr, "   Clean up Docker:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdocker system prune -a\033[0m\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Free up space and try again\n\n")

		case strings.Contains(errMsg, "already exists"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 File already exists:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The devcontainer.json file already exists\n")
			fmt.Fprintf(os.Stderr, "   Location: \033[36m%s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[33m🔍 Why:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Previous run may have partially completed\n")
			fmt.Fprintf(os.Stderr, "   • File was manually created\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check the existing file:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mcat %s\033[0m\n\n", devContainerPath)
			fmt.Fprintf(os.Stderr, "   Remove it if needed:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mrm %s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Remove the file and try again\n\n")

		case strings.Contains(errMsg, "read-only file system"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Read-only filesystem:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The filesystem is mounted as read-only\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   This usually happens when the disk has errors\n")
			fmt.Fprintf(os.Stderr, "   Check system logs:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdmesg | grep -i error\033[0m\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m System administrator help may be needed\n\n")

		case strings.Contains(errMsg, "no such file"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Source file disappeared:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The template file existed but is now missing:\n")
			fmt.Fprintf(os.Stderr, "   \033[36m%s\033[0m\n\n", devContainerExamplePath)

			fmt.Fprintf(os.Stderr, "\033[33m🔍 Why:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • File was deleted between checks\n")
			fmt.Fprintf(os.Stderr, "   • Filesystem issue\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m This shouldn't happen normally. Try running again\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Unexpected error:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Failed to copy file:\n")
			fmt.Fprintf(os.Stderr, "   From: \033[36m%s\033[0m\n", devContainerExamplePath)
			fmt.Fprintf(os.Stderr, "   To:   \033[36m%s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[33m📋 Error details:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   %v\n\n", err)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Try to fix the error above and run again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	devContainerContents, err := os.ReadFile(devContainerPath)

	if err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to read devcontainer.json file\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "no such file"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 File disappeared:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The file was just created but is now missing:\n")
			fmt.Fprintf(os.Stderr, "   \033[36m%s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[33m🔍 Why:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Copy operation may have failed silently\n")
			fmt.Fprintf(os.Stderr, "   • File was deleted immediately after creation\n")
			fmt.Fprintf(os.Stderr, "   • Filesystem issue\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m This shouldn't happen normally. Try running again\n\n")

		case strings.Contains(errMsg, "permission denied"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Permission issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You don't have permission to read this file\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check file permissions:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mls -la %s\033[0m\n\n", devContainerPath)
			fmt.Fprintf(os.Stderr, "   Fix permissions:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36msudo chmod 644 %s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix permissions and try again\n\n")

		case strings.Contains(errMsg, "is a directory"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Unexpected directory:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   There's a directory named 'devcontainer.json' instead of a file\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Remove the directory:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mrm -rf %s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Remove the directory and try again\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\n\033[33m📋 Error details:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   %v\n\n", err)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix the issue above and try again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	updateDevContainerContents := string(devContainerContents)

	replacements = map[string]interface{}{
		`"name": "php debug",`: fmt.Sprintf(`"name": "%s",`, containerName),
	}

	if err := utils.ReplaceAllValue(&updateDevContainerContents, replacements); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to update devcontainer.json file\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "missing"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Template mismatch:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The devcontainer.json doesn't have the expected format\n\n")

			fmt.Fprintf(os.Stderr, "\033[33m🔍 Why:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Looking for: \033[36m\"name\": \"php debug\",\033[0m\n")
			fmt.Fprintf(os.Stderr, "   But it doesn't exist in the file\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check the file contents:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mcat %s\033[0m\n\n", devContainerPath)
			fmt.Fprintf(os.Stderr, "   Option 1: Update the repository's devcontainer.json.example\n")
			fmt.Fprintf(os.Stderr, "   Option 2: Manually edit the devcontainer.json:\n")
			fmt.Fprintf(os.Stderr, "             Change the \"name\" field to: \033[36m\"%s\"\033[0m\n\n", containerName)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix the template format and try again\n\n")

		case strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "empty"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Invalid configuration:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The container name is invalid or empty\n")
			fmt.Fprintf(os.Stderr, "   Container name: \033[36m%s\033[0m\n\n", containerName)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m This shouldn't happen. Please report this as a bug\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Unexpected error:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Failed to update the devcontainer.json file\n\n")

			fmt.Fprintf(os.Stderr, "\033[33m📋 Error details:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   %v\n\n", err)

			fmt.Fprintf(os.Stderr, "\033[36m🔧 Quick fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You can manually edit the file:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36m%s\033[0m\n\n", devContainerPath)
			fmt.Fprintf(os.Stderr, "   Change \"name\" to: \033[36m\"%s\"\033[0m\n\n", containerName)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m After manual edit, continue setup\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	if err := os.WriteFile(devContainerPath, []byte(updateDevContainerContents), 0644); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to write devcontainer.json file\n")

		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "permission denied"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Permission issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You don't have permission to write this file\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check file permissions:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mls -la %s\033[0m\n\n", devContainerPath)
			fmt.Fprintf(os.Stderr, "   Fix permissions:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36msudo chmod 644 %s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix permissions and try again\n\n")

		case strings.Contains(errMsg, "no space left"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Disk space issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Your disk is full - no space left to write files\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check disk space:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdf -h\033[0m\n\n")
			fmt.Fprintf(os.Stderr, "   Clean up Docker:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdocker system prune -a\033[0m\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Free up space and try again\n\n")

		case strings.Contains(errMsg, "read-only file system"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Read-only filesystem:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The filesystem is mounted as read-only\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   This usually happens when the disk has errors\n")
			fmt.Fprintf(os.Stderr, "   Check system logs:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdmesg | grep -i error\033[0m\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m System administrator help may be needed\n\n")

		case strings.Contains(errMsg, "is a directory"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Path is a directory:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   The path exists but it's a directory, not a file:\n")
			fmt.Fprintf(os.Stderr, "   \033[36m%s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Remove the directory:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mrm -rf %s\033[0m\n\n", devContainerPath)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Remove the directory and try again\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\n\033[33m📋 Error details:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   %v\n\n", err)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix the issue above and try again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	done <- true

	fmt.Printf("\r\033[KCreating container workspace completed \033[32m✓\033[0m\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Starting Docker containers", done)

	// if err := utils.UpWithBuild(path); err != nil {
	if err := p.container.CreateContainer(path); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to start Docker containers\n")

		errMsg := err.Error()

		switch {
		case strings.Contains(errMsg, "Cannot connect to the Docker daemon"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 What happened:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Docker daemon is not running\n\n")

			fmt.Fprintf(os.Stderr, "\033[33m🔍 Why:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Docker Desktop is not started\n")
			fmt.Fprintf(os.Stderr, "   • Docker service is stopped\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Option 1: Start Docker Desktop\n\n")
			fmt.Fprintf(os.Stderr, "   Option 2: Start Docker service:\n")
			fmt.Fprintf(os.Stderr, "             $ \033[36msudo systemctl start docker\033[0m\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Start Docker and try again\n\n")

		case strings.Contains(errMsg, "port is already allocated") || strings.Contains(errMsg, "address already in use"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 What happened:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Port %d is already in use by another application\n\n", containerPort)

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Option 1: Find and stop the conflicting process:\n")
			fmt.Fprintf(os.Stderr, "             $ \033[36msudo lsof -i :%d\033[0m\n\n", containerPort)
			fmt.Fprintf(os.Stderr, "   Option 2: Use a different port when running this setup\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Free up the port and try again\n\n")

		case strings.Contains(errMsg, "permission denied"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 What happened:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   You don't have permission to run Docker commands\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Add your user to the docker group:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36msudo usermod -aG docker $USER\033[0m\n\n")
			fmt.Fprintf(os.Stderr, "   Then log out and log back in, or run:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mnewgrp docker\033[0m\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Fix permissions and try again\n\n")

		case strings.Contains(errMsg, "no space left"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 What happened:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Your disk is full - no space left for Docker\n\n")

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check disk space:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdf -h\033[0m\n\n")
			fmt.Fprintf(os.Stderr, "   Clean up Docker resources:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdocker system prune -a\033[0m\n\n")

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Free up space and try again\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 What happened:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Docker containers failed to start\n\n")

			fmt.Fprintf(os.Stderr, "\033[33m📋 Error details:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   %v\n\n", err)

			fmt.Fprintf(os.Stderr, "\033[36m🔧 Troubleshooting:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Check Docker status:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mdocker ps\033[0m\n\n")
			fmt.Fprintf(os.Stderr, "   View container logs:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mcd %s && docker compose logs\033[0m\n\n", path)
			fmt.Fprintf(os.Stderr, "   Check .env file:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mcat %s/.env\033[0m\n\n", path)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Check the error details above and try again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	done <- true
	fmt.Printf("\r\033[KStarting Docker containers completed \033[32m✓\033[0m\n")

	fmt.Printf("\n")
	fmt.Printf("\033[32m✓ Setup Complete!\033[0m 🎉\n\n")

	fmt.Printf("\033[33m📋 Configuration:\033[0m\n")
	fmt.Printf("   • Container Name : %s\n", containerName)
	fmt.Printf("   • Repository Path: %s\n", path)
	fmt.Printf("   • Port          : %d\n\n", containerPort)

	fmt.Printf("\033[36m🚀 Next steps:\033[0m\n")
	fmt.Printf("   1. Open VS Code:\n")
	fmt.Printf("      $ \033[36mcode %s\033[0m\n\n", path)
	fmt.Printf("   2. Access your application:\n")
	fmt.Printf("      🌐 \033[36mhttp://localhost:%d\033[0m\n\n", containerPort)
	fmt.Printf("   3. Start coding in the devcontainer!\n\n")

	codeVersionCommand := exec.Command("code", "--version")

	if _, err = codeVersionCommand.CombinedOutput(); err == nil {

		var openInVSCode bool

		openInVSCodePrompt := &survey.Confirm{
			Message: "Do you want to open the project in VS Code?",
		}

		if err := survey.AskOne(openInVSCodePrompt, &openInVSCode); err != nil {
			log.Fatal(err)
		}

		if openInVSCode {
			openCommand := exec.Command("code", path)

			if _, err := openCommand.CombinedOutput(); err != nil {
				log.Fatalf("error opening project in VS Code: %v", err)
			}
		}
	}
}

func cleanUpFailedSetup(containerName string, path string) {
	done := make(chan bool)

	go utils.ShowLoadingIndicator("Cleaning up config", done)
	if err := config.DeleteProjectConfig(containerName); err != nil {
		done <- true
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m Failed to remove project configuration: %v\n", err)
	} else {
		done <- true
		fmt.Printf("\r\033[KRemoved project configuration \033[32m✓\033[0m\n")
	}

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Removing cloned repository", done)
	if err := os.RemoveAll(path); err != nil {
		done <- true
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m Failed to remove cloned repository: %v\n", err)
	} else {
		done <- true
		fmt.Printf("\r\033[KRemoved cloned repository \033[32m✓\033[0m\n")
	}
}
