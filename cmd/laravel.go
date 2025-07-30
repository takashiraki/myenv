/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

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
			Message: "Enter the container name of laravel.",
		}

		err := survey.AskOne(
			containerNamePrompt, &containerName,
			survey.WithValidator(survey.Required),
			survey.WithValidator(survey.MinLength(3)),
			survey.WithValidator(survey.MaxLength(20)),
		)

		if err != nil {
			log.Fatal(err)
		}

		port := 0
		portPrompt := &survey.Input{
			Message: "Enter the port of laravel.",
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

		survey.AskOne(confirmPrompt, &confirm, survey.WithValidator(survey.Required))
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
