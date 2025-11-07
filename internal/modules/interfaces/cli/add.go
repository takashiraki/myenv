package cli

import (
	"fmt"
	"myenv/internal/config/application"
	"myenv/internal/infrastructure"
	"myenv/internal/modules"
	"myenv/internal/modules/applications"
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
			Options: []string{"Proxy","MySQL"},
		}

		if err := survey.AskOne(modulePrompt, &selectModule); err != nil {
			fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
			return
		}

		module = selectModule
	} else {
		modules := []string{"Proxy","MySQL"}

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
	}
}

func addProxy() {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	targetDir := filepath.Join(homeDir, "dev", "docker_proxy_network")

	if utils.DirIsExists(targetDir) {
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
	configService, err := application.NewConfigService()
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

	if utils.DirIsExists(targetDir) {
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
	configService, err := application.NewConfigService()
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
}

func showErrorHandling(errMsg string){
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
		default:
			fmt.Fprintf(os.Stderr, "\033[33m💡 Hint:\033[0m Please check the error message above and try again.\n")
		}
}