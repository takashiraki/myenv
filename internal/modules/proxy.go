package modules

import (
	"fmt"
	"myenv/internal/config"
	"myenv/internal/config/application"
	"myenv/internal/infrastructure"
	"myenv/internal/utils"
	"os"
	"strings"
)

type ProxyService struct {
	container infrastructure.ContainerInterface
	repository infrastructure.RepositoryInterface
	config_service application.ConfigService
}

func NewProxyService(
	container infrastructure.ContainerInterface,
	repository infrastructure.RepositoryInterface,
	config_service application.ConfigService,
) *ProxyService {
	return &ProxyService{
		container: container,
		repository: repository,
		config_service: config_service,
	}
}

type ModuleConfig struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Module struct {
	Module ModuleConfig `json:"modules"`
}

func (p *ProxyService) CreateProxy(module Module) error {
	moduleConfig := application.Module{
		Name: module.Module.Name,
		Path: module.Module.Path,
	}
	
	if err := p.config_service.AddModule(moduleConfig); err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return err
	}

	done := make(chan bool)

	targetRepo := "https://github.com/takashiraki/docker_proxy_network.git"
	targetPath := module.Module.Path

	go utils.ShowLoadingIndicator("Cloning repository", done)

	if err := p.repository.CloneRepo(targetRepo,targetPath); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to clone repository\n")

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

		cleanUpFailedSetup(module.Module.Name, targetPath)
		return err
	}

	done <- true
	fmt.Printf("\r\033[KCloning repository completed \033[32m✓\033[0m\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Starting Docker containers", done)

	if err := p.container.CreateContainer(targetPath); err != nil {
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
			fmt.Fprintf(os.Stderr, "   Port %d is already in use by another application\n\n", 80)

			fmt.Fprintf(os.Stderr, "\033[36m🔧 How to fix:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   Option 1: Find and stop the conflicting process:\n")
			fmt.Fprintf(os.Stderr, "             $ \033[36msudo lsof -i :%d\033[0m\n\n", 80)
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
			fmt.Fprintf(os.Stderr, "   $ \033[36mcd %s && docker compose logs\033[0m\n\n", targetPath)
			fmt.Fprintf(os.Stderr, "   Check .env file:\n")
			fmt.Fprintf(os.Stderr, "   $ \033[36mcat %s/.env\033[0m\n\n", targetPath)

			fmt.Fprintf(os.Stderr, "\033[32m→ Next:\033[0m Check the error details above and try again\n\n")
		}

		cleanUpFailedSetup(module.Module.Name, targetPath)
		return err
	}

	done <- true
	fmt.Printf("\r\033[KStarting Docker containers completed \033[32m✓\033[0m\n")

	fmt.Printf("\n")
	fmt.Printf("\033[32m✓ Setup Complete!\033[0m 🎉\n\n")

	fmt.Printf("\033[33m📋 Configuration:\033[0m\n")
	fmt.Printf("   • Container Name : %s\n", module.Module.Name)
	fmt.Printf("   • Repository Path: %s\n", targetPath)

	return nil
}

func cleanUpFailedSetup(moduleName string, path string) {
	done := make(chan bool)

	go utils.ShowLoadingIndicator("Cleaning up config", done)
	if err := config.DeleteModulConfig(moduleName); err != nil {
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

	fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
}