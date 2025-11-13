package cli

import (
	"fmt"
	"myenv/internal/config/application"
	"myenv/internal/infrastructure"
	"myenv/internal/modules"
	"myenv/internal/utils"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

func EntryPoint(module string) {
	if module == "" {
		var selectModule string

		modulePrompt := &survey.Select{
			Message: "Select the module you want to add:",
			Options: []string{"Proxy", "MySQL", "Mailpit"},
		}

		if err := survey.AskOne(modulePrompt, &selectModule); err != nil {
			fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
			return
		}

		module = selectModule
	} else {
		modules := []string{"Proxy", "MySQL", "Mailpit"}

		if !slices.Contains(modules, module) {
			fmt.Printf("\n\033[31m✗ Error:\033[0m Invalid module selected\n")
			return
		}
	}

	switch module {
	case "Proxy":
		addProxy()
	case "MySQL":
		AddMySQL()
	case "Mailpit":
		AddMailpit()
	}
}

func addProxy() {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	targetDir := filepath.Join(homeDir, "dev", "docker_proxy_network")

	if _,err := os.Stat(targetDir); !os.IsNotExist(err) {
		fmt.Printf("\n\033[31m✗ Error:\033[0m Directory %s already exists\n", targetDir)
		return
	}

	utils.ClearTerminal()

	fmt.Printf("\n")
	fmt.Printf("\033[33m📋 Configuration:\033[0m\n")
	fmt.Printf("   • Module Name     : %s\n", "proxy")
	fmt.Printf("   • Target Directory : %s\n", targetDir)

	var confirmResult bool
	confirmPrompt := &survey.Confirm{
		Message: "Is it okay to start building the environment with this configuration?",
	}

	if err := survey.AskOne(confirmPrompt, &confirmResult); err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	if !confirmResult {
		fmt.Printf("\n\033[33mSetup cancelled.\033[0m Returning to configuration...")
		return
	}

	container := infrastructure.NewDockerContainer()
	repository := infrastructure.NewGitRepository()
	configService, err := application.NewConfigService(container)
	if err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}
	service := modules.NewProxyService(container, repository, *configService)

	module := modules.Module{
		Module: modules.ModuleConfig{
			Name: "proxy",
			Path: targetDir,
		},
	}

	if err := service.CreateProxy(module); err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}
}

func AddMySQL() {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	targetDir := filepath.Join(homeDir, "dev", "docker_mysql")

	if _,err := os.Stat(targetDir); !os.IsNotExist(err) {
		fmt.Printf("\n\033[31m✗ Error:\033[0m Directory %s already exists\n", targetDir)
		return
	}

	utils.ClearTerminal()

	fmt.Printf("\n")
	fmt.Printf("\033[33m📋 Configuration:\033[0m\n")
	fmt.Printf("   • Module Name     : %s\n", "mysql")
	fmt.Printf("   • Target Directory : %s\n", targetDir)

	var confirmResult bool
	confirmPrompt := &survey.Confirm{
		Message: "Is it okay to start building the environment with this configuration?",
	}

	if err := survey.AskOne(confirmPrompt, &confirmResult); err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	if !confirmResult {
		fmt.Printf("\n\033[33mSetup cancelled.\033[0m Returning to configuration...")
		return
	}

	events := make(chan applications.Event)
	container := infrastructure.NewDockerContainer()
	repository := infrastructure.NewGitRepository()
	configService, err := application.NewConfigService(container)
	if err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}
	service := applications.NewMySQLService(container, repository, *configService)

	done := make(chan bool)
	var loadingDone chan bool

	stopLoading := func() {
		if loadingDone != nil {
			select {
			case loadingDone <- true:
			default:
			}
			loadingDone = nil
		}
	}

	go func() {
		for event := range events {
			switch event.Status {
			case "running":
				stopLoading()
				loadingDone = make(chan bool)
				go utils.ShowLoadingIndicator(event.Message, loadingDone)
			case "success":
				stopLoading()
				fmt.Printf("\r\033[K\033[32m✓\033[0m %s\n", event.Message)
			case "error":
				stopLoading()
				fmt.Printf("\r\033[K\033[31m✗\033[0m %s\n", event.Message)
			}
		}

		stopLoading()
		done <- true
	}()

	if err := service.Create(events); err != nil {
		close(events)
		<-done
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", err)

		errMsg := err.Error()
		showErrorHandling(errMsg)
		return
	}

	close(events)
	<-done

	fmt.Printf("\n")
	fmt.Printf("\033[32m✓ Setup Complete!\033[0m 🎉\n\n")

	fmt.Printf("\033[33m📋 Configuration:\033[0m\n")
	fmt.Printf("   • Container Name : %s\n", "mysql")
	fmt.Printf("   • Repository Path: %s\n", targetDir)
}

func AddMailpit() {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	targetDir := filepath.Join(homeDir, "dev", "docker_mailpit")

	if _, err := os.Stat(targetDir);  err == nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m Directory %s already exists\n", targetDir)
		return
	}

	utils.ClearTerminal()

	fmt.Printf("\n")
	fmt.Printf("\033[33m📋 Configuration:\033[0m\n")
	fmt.Printf("   • Module Name     : %s\n", "mailpit")
	fmt.Printf("   • Target Directory : %s\n", targetDir)

	var confirmResult bool
	confirmPrompt := &survey.Confirm{
		Message: "Is it okay to start building the environment with this configuration?",
	}

	if err := survey.AskOne(confirmPrompt, &confirmResult); err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	if !confirmResult {
		fmt.Printf("\n\033[33mSetup cancelled.\033[0m Returning to configuration...")
		return
	}

	container := infrastructure.NewDockerContainer()
	repository := infrastructure.NewGitRepository()
	configService, err := application.NewConfigService(container)
	if err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	service := applications.NewMailpitService(container, repository, *configService)

	events := make(chan applications.Event)

	done := make(chan bool)
	var loadingDone chan bool

	stopLoading := func() {
		if loadingDone != nil {
			select {
			case loadingDone <- true:
			default:
			}
			loadingDone = nil
		}
	}

	go func() {
		for event := range events {
			switch event.Status {
			case "running":
				stopLoading()
				loadingDone = make(chan bool)
				go utils.ShowLoadingIndicator(event.Message, loadingDone)
			case "success":
				stopLoading()
				fmt.Printf("\r\033[K\033[32m✓\033[0m %s\n", event.Message)
			case "error":
				stopLoading()
				fmt.Printf("\r\033[K\033[31m✗\033[0m %s\n", event.Message)
			}
		}

		stopLoading()
		done <- true
	}()

	if err := service.Create(events); err != nil {
		close(events)
		<-done
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", err)

		errMsg := err.Error()
		showErrorHandling(errMsg)
		return
	}

	close(events)
	<-done

	fmt.Printf("\n")
	fmt.Printf("\033[32m✓ Setup Complete!\033[0m 🎉\n\n")

	fmt.Printf("\033[33m📋 Configuration:\033[0m\n")
	fmt.Printf("   • Container Name : %s\n", "mailpit")
	fmt.Printf("   • Repository Path: %s\n", targetDir)
}

func showErrorHandling(errMsg string) {
	switch {
	case strings.Contains(errMsg, "Could not resolve host"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Network connection is not available.\n")
		fmt.Fprintf(os.Stderr, "           Please check your internet connection and try again.\n")
	case strings.Contains(errMsg, "exit status 128"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Git command failed.\n")
		fmt.Fprintf(os.Stderr, "           Please check your network connection or repository URL.\n")
	case strings.Contains(errMsg, "already exists"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Target directory already exists.\n")
		fmt.Fprintf(os.Stderr, "           Please remove the existing directory or choose a different location.\n")
	case strings.Contains(errMsg, "permission denied"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Permission denied.\n")
		fmt.Fprintf(os.Stderr, "           Please check file permissions and try again.\n")
	case strings.Contains(errMsg, "Cannot connect to the Docker daemon"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Docker daemon is not running.\n")
		fmt.Fprintf(os.Stderr, "           Please start Docker and try again.\n")
	case strings.Contains(errMsg, "docker: not found") || strings.Contains(errMsg, "executable file not found"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Docker is not installed or not in PATH.\n")
		fmt.Fprintf(os.Stderr, "           Please install Docker and try again.\n")
	case strings.Contains(errMsg, "port is already allocated") || strings.Contains(errMsg, "address already in use"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m The port is already in use.\n")
		fmt.Fprintf(os.Stderr, "           Please stop the conflicting container or service and try again.\n")
	case strings.Contains(errMsg, "no space left on device"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Not enough disk space.\n")
		fmt.Fprintf(os.Stderr, "           Please free up disk space and try again.\n")
	case strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline exceeded"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Operation timed out.\n")
		fmt.Fprintf(os.Stderr, "           Please check your network connection or try again later.\n")
	case strings.Contains(errMsg, "repository not found"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Repository not found.\n")
		fmt.Fprintf(os.Stderr, "           Please check the repository URL and try again.\n")
	case strings.Contains(errMsg, "authentication failed") || strings.Contains(errMsg, "Authentication failed"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Git authentication failed.\n")
		fmt.Fprintf(os.Stderr, "           Please check your credentials or access permissions.\n")
	case strings.Contains(errMsg, "container already running"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Container is already running.\n")
		fmt.Fprintf(os.Stderr, "           Please stop the existing container and try again.\n")
	case strings.Contains(errMsg, "no such file or directory"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Required file or directory not found.\n")
		fmt.Fprintf(os.Stderr, "           Please check the file path and try again.\n")
	default:
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Please check the error message above and try again.\n")
	}
}
