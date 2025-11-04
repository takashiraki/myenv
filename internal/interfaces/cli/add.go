package cli

import (
	"fmt"
	"myenv/internal/infrastructure"
	"myenv/internal/modules"
	"myenv/internal/utils"
	"os"
	"path/filepath"
	"slices"

	"github.com/AlecAivazis/survey/v2"
)

func EntryPoint(module string) {
	if module == "" {
		var selectModule string

		modulePrompt := &survey.Select{
			Message: "Select the module you want to add:",
			Options: []string{"Proxy"},
		}

		if err := survey.AskOne(modulePrompt, &selectModule); err != nil {
			fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
			return
		}

		module = selectModule
	} else {
		modules := []string{"Proxy"}

		if !slices.Contains(modules, module) {
			fmt.Printf("\n\033[31m✗ Error:\033[0m Invalid module selected\n")
			return
		}
	}

	switch module {
	case "Proxy":
		addProxy()
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
	service := modules.NewProxyService(container, repository)

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