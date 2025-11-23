package cli

import (
	"fmt"
	"myenv/internal/config/application"
	"myenv/internal/config/utils"
	"myenv/internal/events"
	"myenv/internal/infrastructure"
	"myenv/internal/lang/php/laravel/applications"
	Langutils "myenv/internal/lang/utils"
	CommonUtils "myenv/internal/utils"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
)

func EntryPoint() {
	CommonUtils.ClearTerminal()

	containerName := ""
	containerNamePrompt := &survey.Input{
		Message: "Enter the container name : ",
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
		fmt.Fprintf(os.Stderr, "\n\033[31mError:\033[0m %s\n", err.Error())
		return
	}

	containerProxy := ""
	containerProxyPrompt := &survey.Input{
		Message: "Enter the virtual host (e.g., myapp.local) : ",
	}

	ProxyErr := survey.AskOne(
		containerProxyPrompt, &containerProxy,
		survey.WithValidator(survey.Required),
		survey.WithValidator(utils.ValidateProxy),
	)

	if ProxyErr != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31mError:\033[0m %s\n", ProxyErr.Error())
		return
	}

	CommonUtils.ClearTerminal()

	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31mError:\033[0m %s\n", err.Error())
		return
	}

	targetDir := filepath.Join(homeDir, "dev", containerName)

	if _, err := os.Stat(targetDir); err == nil {
		fmt.Fprintf(os.Stderr, "\n\033[31mError:\033[0m Directory '%s' already exists.\n", targetDir)
		return
	}

	fmt.Printf("\n")
	fmt.Printf("\033[33m📋 Configuration:\033[0m\n")
	fmt.Printf("   • Container name : %s\n", containerName)
	fmt.Printf("   • Clone path     : %s\n", targetDir)
	fmt.Printf("   • Proxy          : %s\n", containerProxy)
	fmt.Printf("   • Framework      : Laravel\n")
	fmt.Printf("   • Language       : PHP\n\n")

	var confirmResult bool
	confirmPrompt := &survey.Confirm{
		Message: "Is it okay to start building the environment with this configuration?",
	}

	if err := survey.AskOne(confirmPrompt, &confirmResult); err != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31mError:\033[0m %s\n", err.Error())
		return
	}

	if !confirmResult {
		fmt.Println("\n\033[33m⚠️  Canceled by user.\033[0m")
		return
	}

	events := make(chan events.Event)
	container := infrastructure.NewDockerContainer()
	repository := infrastructure.NewGitRepository()
	configService, err := application.NewConfigService(container, repository)

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31mError:\033[0m %s\n", err.Error())
		return
	}

	service := applications.NewLaravelService(
		container,
		repository,
		*configService,
	)

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
				go CommonUtils.ShowLoadingIndicator(event.Name, loadingDone)
			case "success":
				stopLoading()
				fmt.Printf("\r\033[K\033[32m✓\033[0m %s\n", event.Message)
			case "error":
				stopLoading()
				fmt.Printf("\r\033[K\033[31m✗ %s\033[0m\n", event.Message)
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
		Langutils.ShowErrorHandling(errMsg)
		Langutils.CleanUpFailedSetup(containerName, targetDir)
		return
	}

	close(events)
	<-done

	fmt.Print("\r\033[K")
	Langutils.SetUpCompleted(
		containerName,
		targetDir,
		containerProxy,
	)
}
