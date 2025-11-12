package interfaces

import (
	"fmt"
	"myenv/internal/config"
	"myenv/internal/config/application"
	"myenv/internal/events"
	"myenv/internal/infrastructure"
	"myenv/internal/lang/php/wordpress/appliactions"
	"myenv/internal/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

func EntryPoint() {
	utils.ClearTerminal()

	containerName := ""
	containerNamePrompt := &survey.Input{
		Message: "Enter the container name of WordPress : ",
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

	containerProxy := ""
	containerProxyPrompt := &survey.Input{
		Message: "Enter the local domain (e.g., myapp.localhost): ",
	}

	ProxyErr := survey.AskOne(
		containerProxyPrompt, &containerProxy,
		survey.WithValidator(survey.Required),
		survey.WithValidator(utils.ValidateProxy),
	)

	if ProxyErr != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", ProxyErr)
		return
	}

	utils.ClearTerminal()

	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	targetDir := filepath.Join(homeDir, "dev", containerName)

	if _, err := os.Stat(targetDir); err == nil {
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m Directory %s already exists\n", targetDir)
		return
	}

	fmt.Printf("\n")
	fmt.Printf("\033[33m📋 Configuration:\033[0m\n")
	fmt.Printf("   • Container name : %s\n", containerName)
	fmt.Printf("   • Clone path     : %s\n", targetDir)
	fmt.Printf("   • Proxy          : %s\n", containerProxy)
	fmt.Printf("   • Framework      : WordPress\n")
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
		fmt.Printf("\n\033[33mSetup cancelled.\033[0m Returning to configuration...")
		return
	}

	events := make(chan events.Event)
	container := infrastructure.NewDockerContainer()
	repository := infrastructure.NewGitRepository()
	configService, err := application.NewConfigService(container)

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	service := appliactions.NewWordpressService(
		container,
		repository,
		*configService,
	)

	done := make(chan bool)
	var loadingDone chan bool

	stopLoading := func() {
		if loadingDone != nil {
			loadingDone <- true
			fmt.Print("\r\033[K") // 行をクリア
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

	if err := service.Create(events, containerName, containerProxy); err != nil {
		close(events)
		<-done
		errMsg := err.Error()
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", errMsg)
		showErrorHandling(errMsg)
		cleanUpFailedSetup(containerName, targetDir)
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	close(events)
	<-done

	// ローディングアニメーションが完全に停止するまで少し待つ
	fmt.Print("\r\033[K")

	fmt.Printf("\n")
	fmt.Printf("\033[32m✓ Setup Complete!\033[0m 🎉\n\n")

	fmt.Printf("\033[33m📋 Configuration:\033[0m\n")
	fmt.Printf("   • Container Name : %s\n", containerName)
	fmt.Printf("   • Repository Path: %s\n", targetDir)
	fmt.Printf("   • Proxy          : %s\n\n", containerProxy)

	fmt.Printf("\033[36m🚀 Next steps:\033[0m\n")
	fmt.Printf("   1. Open VS Code:\n")
	fmt.Printf("      $ \033[36mcode %s\033[0m\n\n", targetDir)
	fmt.Printf("   2. Access your application:\n")
	fmt.Printf("      🌐 \033[36mhttp://%s\033[0m\n\n", containerProxy)
	fmt.Printf("   3. Start coding in the devcontainer!\n\n")

	codeCommand := exec.Command("code", "--version")
	devcontainerCommand := exec.Command("devcontainer", "--version")
	cursorCommand := exec.Command("cursor", "--version")

	_, codeErr := codeCommand.CombinedOutput()
	_, devcontainerErr := devcontainerCommand.CombinedOutput()
	_, cursorErr := cursorCommand.CombinedOutput()

	var options []string
	var commands []string

	if codeErr == nil {
		options = append(options, "VS Code (code)")
		commands = append(commands, "code")
	}

	if cursorErr == nil {
		options = append(options, "Cursor (cursor)")
		commands = append(commands, "cursor")
	}

	if devcontainerErr == nil {
		options = append(options, "devcontainer CLI")
		commands = append(commands, "devcontainer")
	}

	if len(options) > 0 {
		options = append(options, "Skip (open manually later)")

		var selectedOption string
		openPrompt := &survey.Select{
			Message: "How would you like to open this project?",
			Options: options,
		}

		if err := survey.AskOne(openPrompt, &selectedOption); err != nil {
			fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", err)
			return
		}

		if selectedOption != "Skip (open manually later)" {
			for i, option := range options[:len(options)-1] {
				if selectedOption == option {
					var openCommand *exec.Cmd

					if commands[i] == "devcontainer" {
						openCommand = exec.Command("devcontainer", "open", targetDir)
					} else {
						openCommand = exec.Command(commands[i], targetDir)
					}

					if _, err := openCommand.CombinedOutput(); err != nil {
						fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m Failed to open project: %v\n", err)
					} else {
						fmt.Printf("\n\033[32m✓\033[0m Project opened in %s\n", selectedOption)
					}
					break
				}
			}
		}
	}
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
