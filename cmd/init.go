/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"myenv/internal/config"
	"myenv/internal/lang/interfaces"
	"myenv/internal/utils"

	"github.com/spf13/cobra"
)

var (
	lang string
	fw   string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new containerized development environment",
	Long: `Initialize a new containerized development environment with your chosen language and framework.

This command guides you through setting up a development environment by:
  - Selecting a programming language (e.g., PHP)
  - Choosing a framework or starting with a basic setup
  - Creating the necessary Docker configuration and project files

Example:
  myenv init                      # Interactive mode with prompts
  myenv init -l PHP               # Specify language directly
  myenv init -l PHP -f Laravel    # Specify both language and framework`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.CheckConfig(); err != nil {
			fmt.Println("\n\033[31m✗ Error:\033[0m Configuration Missing")
			fmt.Println("\nNo configuration found. Please run the following command first to initialize myenv:")
			fmt.Println("\n  myenv setup")
			fmt.Println("\nThis will create the necessary configuration files in ~/.config/myenv/")
			return
		}

		utils.ClearTerminal()

		config.CheckForUpdates(version)

		interfaces.EntryPoint(lang, fw)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	initCmd.Flags().StringVarP(&lang, "lang", "l", "", "Specify the programming language (e.g., PHP)")
	initCmd.Flags().StringVarP(&fw, "framework", "f", "", "Specify the programming language (e.g., Laravel)")
}
