/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"myenv/internal/config/interfaces"
	"myenv/internal/utils"

	"github.com/spf13/cobra"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start up your development environment containers",
	Long: `Start up and run your containerized development environment.

This command starts all Docker containers for your project based on
the configuration created with 'myenv init'.

Example:
  myenv up                     # Start all containers for your project`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.ClearTerminal()
		interfaces.UpProject()
	},
}

func init() {
	rootCmd.AddCommand(upCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// upCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// upCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
