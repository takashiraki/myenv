/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"myenv/internal/config/interfaces"
	"myenv/internal/utils"

	"github.com/spf13/cobra"
)

var (
	quick bool
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure myenv settings and preferences",
	Long: `Configure myenv settings and preferences for your development environment.

This command sets up myenv configuration by:
  - Creating configuration files in ~/.config/myenv/
  - Setting up default preferences
  - Configuring environment settings

This command should be run once before using other myenv commands.

Example:
  myenv setup                  # Interactive setup with prompts
  myenv setup --quick          # Quick setup with default settings`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.ClearTerminal()
		interfaces.SetUp(quick)
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	setupCmd.Flags().BoolVar(&quick, "quick", false, "Quick setup")
}
