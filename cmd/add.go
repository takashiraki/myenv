/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"myenv/internal/config"
	"myenv/internal/modules/interfaces/cli"
	"myenv/internal/utils"

	"github.com/spf13/cobra"
)

var (
	module string
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
		cli.EntryPoint(module)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	addCmd.Flags().StringVarP(&module, "module", "m", "", "Specify the module you want to add")
}
