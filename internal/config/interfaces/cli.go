package interfaces

import (
	"fmt"
	"myenv/internal/config/application"
	"myenv/internal/infrastructure"
	"myenv/internal/utils"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

func SetUp(quick bool) {
	container := infrastructure.NewDockerContainer()
	repository := infrastructure.NewGitRepository()
	configService, err := application.NewConfigService(container, repository)

	if err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	if _, err = configService.GetConfig(); err == nil {
		fmt.Printf("\n\033[33mℹ Info:\033[0m Config file already exists\n")
		return
	}

	fmt.Printf("\n\033[36m🔧 Setup Mode:\033[0m %s\n", map[bool]string{true: "Quick Setup", false: "Standard Setup"}[quick])
	if quick {
		fmt.Printf("\033[33mℹ Info:\033[0m Quick setup will create basic configuration only.\n")
		fmt.Printf("          Use 'myenv create' after setup to create project containers.\n\n")
	} else {
		fmt.Printf("\033[33mℹ Info:\033[0m Standard setup includes configuration and network creation.\n\n")
	}

	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: "Do you want to continue with this setup?",
		Default: true,
	}

	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	if !confirm {
		fmt.Printf("\n\033[33mℹ Info:\033[0m Setup cancelled.\n")
		return
	}

	lang := "en"
	containerRuntime := "docker"
	events := make(chan application.Event)
	done := make(chan bool)
	var loadingDone chan bool

	stopLoading := func() {
		if loadingDone != nil {
			loadingDone <- true
			fmt.Print("\r\033[K")
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
			case "skipped":
				stopLoading()
				fmt.Printf("\r\033[K\033[33mℹ\033[0m %s\n", event.Message)
				loadingDone = make(chan bool)
				go utils.ShowLoadingIndicator("Continuing...", loadingDone)
			case "error":
				stopLoading()
				fmt.Printf("\r\033[K\033[31m✗\033[0m %s\n", event.Message)
			}
		}

		stopLoading()
		done <- true
	}()

	if err := configService.CreateConfig(lang, containerRuntime, events, quick); err != nil {
		close(events)
		<-done

		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	close(events)
	<-done

	fmt.Printf("\n")
	fmt.Printf("\033[32m✓ Setup Complete!\033[0m 🎉\n\n")

	fmt.Printf("\033[33m📋 Configuration:\033[0m\n")
	fmt.Printf("   • Container runtime    : %s\n", containerRuntime)
	fmt.Printf("   • Language             : %s\n", lang)
	fmt.Printf("   • Create Proxy Network : %s\n", "yes")
	fmt.Printf("   • Create Infra Network : %s\n", "yes")
}

func UpProject() {
	container := infrastructure.NewDockerContainer()
	repository := infrastructure.NewGitRepository()
	configService, err := application.NewConfigService(container, repository)

	if err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	projects, err := configService.GetProjects()

	if err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	projectNames := []string{}
	for _, project := range projects {
		projectNames = append(projectNames, project.ContainerName)
	}

	projectPromopt := &survey.Select{
		Message: "Select the project you want to up: ",
		Options: projectNames,
	}

	projectName := ""

	if err = survey.AskOne(projectPromopt, &projectName); err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	done := make(chan bool)

	go utils.ShowLoadingIndicator("Upping project", done)

	project, err := configService.UpProject(projectName)

	if err != nil {
		done <- true
		fmt.Print("\r\033[K")
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", err)
		showErrorHandling(err.Error())
		return
	}

	done <- true
	fmt.Print("\r\033[K")

	fmt.Printf("\n\033[32m✓ Project upped!\033[0m 🎉\n\n")

	fmt.Printf("\033[33m📋 Configuration:\033[0m\n")
	fmt.Printf("   • Container Name : %s\n", project.ContainerName)
	fmt.Printf("   • Repository Path: %s\n", project.Path)
	fmt.Printf("   • Proxy          : %s\n\n", project.ContainerProxy)

	fmt.Printf("\033[36m🚀 Next steps:\033[0m\n")
	fmt.Printf("   1. Open VS Code:\n")
	fmt.Printf("      $ \033[36mcode %s\033[0m\n\n", project.Path)
	fmt.Printf("   2. Access your application:\n")
	fmt.Printf("      🌐 \033[36mhttp://%s\033[0m\n\n", project.ContainerProxy)
	fmt.Printf("   3. Start coding in the devcontainer!\n\n")

	codeCommand := exec.Command("code", "--version")
	devcontainerCommand := exec.Command("devcontainer", "--version")
	cursorCommand := exec.Command("cursor", "--version")

	_, codeErr := codeCommand.CombinedOutput()
	_, devcontainerErr := devcontainerCommand.CombinedOutput()
	_, cursorErr := cursorCommand.CombinedOutput()

	targetDir := project.Path

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
	case strings.Contains(errMsg, "project not found"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Project not found in configuration.\n")
		fmt.Fprintf(os.Stderr, "           Please create a project first using 'myenv init'.\n")
	case strings.Contains(errMsg, "project file does not exist"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Configuration file does not exist.\n")
		fmt.Fprintf(os.Stderr, "           Please run 'myenv setup' first to initialize the configuration.\n")
	case strings.Contains(errMsg, "module not found"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Required module not found in configuration.\n")
		fmt.Fprintf(os.Stderr, "           The project may reference a module that hasn't been added.\n")
		fmt.Fprintf(os.Stderr, "           Please check your project configuration or add the missing module.\n")
	case strings.Contains(errMsg, "Cannot connect to the Docker daemon") || strings.Contains(errMsg, "Is the docker daemon running"):
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
	case strings.Contains(errMsg, "permission denied"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Permission denied.\n")
		fmt.Fprintf(os.Stderr, "           Please check file/directory permissions or Docker socket permissions.\n")
	case strings.Contains(errMsg, "container already running"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Container is already running.\n")
		fmt.Fprintf(os.Stderr, "           The containers may already be started.\n")
	case strings.Contains(errMsg, "no such file or directory"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Required file or directory not found.\n")
		fmt.Fprintf(os.Stderr, "           Please check if the project path exists and contains docker-compose.yml.\n")
	case strings.Contains(errMsg, "network") && strings.Contains(errMsg, "not found"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Docker network not found.\n")
		fmt.Fprintf(os.Stderr, "           Please run 'myenv setup' to create the required networks.\n")
	case strings.Contains(errMsg, "Compose file") || strings.Contains(errMsg, "docker-compose"):
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Docker Compose configuration error.\n")
		fmt.Fprintf(os.Stderr, "           Please check your docker-compose.yml file for syntax errors.\n")
	default:
		fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Please check the error message above and try again.\n")
		fmt.Fprintf(os.Stderr, "           If the problem persists, check Docker logs or project configuration.\n")
	}
}
