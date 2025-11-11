package interfaces

import (
	"fmt"
	"myenv/internal/config/application"
	"myenv/internal/infrastructure"
	"myenv/internal/utils"
	"os"
)

func SetUp() {
	container := infrastructure.NewDockerContainer()
	configService, err := application.NewConfigService(container)

	if err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	if _, err = configService.GetConfig();err == nil {
		fmt.Printf("\n\033[33mℹ Info:\033[0m Config file already exists\n")
		return
	}

	lang := "en"
	containerRuntime := "docker"
	events := make(chan application.Event)
	done := make(chan bool)
	var loadingDone chan bool

	stopLoading := func ()  {
		if loadingDone != nil {
			select {
			case loadingDone <- true:
			default:
			}

			loadingDone = nil
		}
	}
	go func ()  {
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

	if err := configService.CreateConfig(lang, containerRuntime, events); err != nil {
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