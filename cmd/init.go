/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	"myenv/internal/framework"
	"myenv/internal/lang/php"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var (
	lang      string
	fw string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		clearTerminal()

		if lang == "" {
			fmt.Println("Initializing your environment...")
			var selectedLang string

			langPrompot := &survey.Select{
				Message: "Select the language you want to use:",
				Options: []string{"PHP"},
			}

			if err := survey.AskOne(langPrompot, &selectedLang); err != nil {
				log.Fatal(err)
			}

			lang = selectedLang
		}

		switch lang {
		case "PHP":
			if fw != "" {
				framework.PHPFramework(fw)
			} else {
				fwPrompt := &survey.Select{
					Message: "Select the framework you want to use: ",
					Options: []string{"Laravel", "None"},
				}

				if err := survey.AskOne(fwPrompt, &fw); err != nil {
					log.Fatal(err)
				}

				if fw == "None" {
					php.PHP()
				} else {
					framework.PHPFramework(fw)
				}
			}
		default:
			log.Fatal("Unsupported language selected.")
		}
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
