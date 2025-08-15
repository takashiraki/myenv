/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.0.2"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "myenv",
	Version: version,
	Short: "A CLI tool for managing containerized development environments",
	Long: `
╔═══════════════════════════════════════════════════════════════════════════════╗
║                                  MyEnv v` + version + `                                 ║
║                                                                               ║
║  🚀 A CLI tool for managing containerized development environments            ║
║                                                                               ║
║  • Automated Docker container setup                                           ║
║  • Smart port management & conflict prevention                                ║
║  • VS Code integration with devcontainer support                              ║
║  • Pre-configured development templates                                       ║
║                                                                               ║
║  Get started: myenv laravel                                                   ║
╚═══════════════════════════════════════════════════════════════════════════════╝
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(cmd.Long)
		fmt.Println("\nUsage:")
		fmt.Println("  myenv [command]")
		fmt.Println("\nAvailable Commands:")
		fmt.Println("  laravel     Create a new Laravel container")
		fmt.Println("  version     Show version information")
		fmt.Println("  completion  Generate the autocompletion script for the specified shell")
		fmt.Println("  help        Help about any command")
		fmt.Println("\nFlags:")
		fmt.Println("  -h, --help     help for myenv")
		fmt.Println("  -v, --version  version for myenv")
		fmt.Println("\nUse \"myenv [command] --help\" for more information about a command.")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.myenv.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
}


