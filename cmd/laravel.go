/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

// laravelCmd represents the laravel command
var laravelCmd = &cobra.Command{
	Use:   "laravel",
	Short: "Create a new Laravel container",
	Long:  `Create a new Laravel container`,
	Run: func(cmd *cobra.Command, args []string) {
		clearTerminal()
		fmt.Println("laravel called")

		containerName := ""
		containerNamePrompt := &survey.Input{
			Message: "Enter the container name of laravel : ",
		}

		err := survey.AskOne(
			containerNamePrompt, &containerName,
			survey.WithValidator(survey.Required),
			survey.WithValidator(survey.MinLength(3)),
			survey.WithValidator(survey.MaxLength(20)),
			survey.WithValidator(validateLaravelProjectName),
		)

		if err != nil {
			log.Fatal(err)
		}

		port := 0
		portPrompt := &survey.Input{
			Message: "Enter the port of laravel : ",
		}

		err = survey.AskOne(
			portPrompt, &port,
			survey.WithValidator(survey.Required),
			survey.WithValidator(validatePort),
		)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(containerName, port)

		clearTerminal()
		fmt.Printf("select: Laravel\n\n")
		fmt.Println("Check your configuration:")
		fmt.Printf("Container Name: %s\n", containerName)
		fmt.Printf("Port: %d\n", port)

		confirm := false
		confirmPrompt := &survey.Confirm{
			Message: "Is it okay to create a Laravel container with this configuration?",
		}

		err = survey.AskOne(confirmPrompt, &confirm, survey.WithValidator(survey.Required))
		if err != nil {
			log.Fatal(err)
			return
		}

		if !confirm {
			fmt.Println("Container creation cancelled")
			return
		}

		targetRepo := "https://github.com/takashiraki/docker_laravel_port.git"
		homeDir, err := os.UserHomeDir()

		if err != nil {
			log.Fatalf("Error getting home directory: %v", err)
		}

		devPath := filepath.Join(homeDir, "dev")
		targetPath := filepath.Join(devPath, containerName)

		if err := os.MkdirAll(devPath, 0755); err != nil {
			log.Fatalf("Error creating dev directory: %v", err)
		}

		CloneRepository(targetRepo, targetPath)

		srcPath := "src"
		dockerPath := "Infra/php"

		if err := CreateEnvFile(
			containerName,
			targetPath,
			srcPath,
			dockerPath,
			80,
			8080,
		); err != nil {
			log.Fatalf("Error creating .env file: %v", err)
		}

		devContainerFilePath := filepath.Join(targetPath, ".devcontainer")
		if err := createDevContainerFile(containerName, devContainerFilePath); err != nil {
			log.Fatalf("Error creating devcontainer.json file: %v", err)
		}

		if err := bootContainer(targetPath); err != nil {
			log.Fatalf("Error booting container: %v", err)
		}

		commandArgs := []string{"compose", "exec", "php", "laravel", "new", containerName, "--no-interaction", "--phpunit"}
		indicator := "Creating new laravel project"
		completeMessage := "Creating new laravel project completed"

		if err := execCommand("docker", commandArgs, targetPath, indicator, completeMessage); err != nil {
			log.Fatalf("Error creating new laravel project: %v", err)
		}

		if err := updateEnvSrcPath(targetPath, srcPath + "/" + containerName); err != nil {
			log.Fatalf("Error updating .env file: %v", err)
		}

		if err := rebbuildContainer(targetPath); err != nil {
			log.Fatalf("Error rebuilding container: %v", err)
		}

		clearTerminal()

		fmt.Printf("\r\033[KLaravel container created successfully ✓\n")
		fmt.Printf("Access the project at http://localhost:%d\n", port)
	},
}

func init() {
	rootCmd.AddCommand(laravelCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// laravelCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// laravelCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func validateLaravelProjectName(val any) error {
	str := val.(string)

	if str == "laravel" {
		return errors.New("laravel is a reserved project name")
	}

	err := validateProjectName("laravel_" + str)

	if err != nil {
		return errors.New("project name already exists")
	}

	return nil
}
