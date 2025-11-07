package interfaces

import (
	"fmt"
	"myenv/internal/config/application"
)

func SetUp() {
	configService, err := application.NewConfigService()

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

	if err := configService.CreateConfig(lang, containerRuntime); err != nil {
		fmt.Printf("\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	fmt.Printf("\n\033[32m✓ Success:\033[0m Config file created successfully\n")
	fmt.Printf("  Lang: %s\n", lang)
	fmt.Printf("  Container Runtime: %s\n", containerRuntime)
}