package php

import (
	"fmt"
	"log"
	"myenv/internal/utils"

	"github.com/AlecAivazis/survey/v2"
)

func PHP() {
	utils.ClearTerminal()
	fmt.Println("PHP called")

	clonePrompt := &survey.Select{
		Message: "Do you want to clone repository of PHP project?",
		Options: []string{"Yes", "No"},
	}

	clone := ""

	if err := survey.AskOne(clonePrompt, &clone); err != nil {
		log.Fatal(err)
	}

	switch clone {
	case "Yes":
		fmt.Println("Clone project")
	case "No":
		fmt.Println("Create project")
		createProject()
	}
}

func createProject() {
	containerName := ""
	containerNamePrompt := &survey.Input{
		Message: "Enter the container name of PHP : ",
	}

	err := survey.AskOne(
		containerNamePrompt, &containerName,
		survey.WithValidator(survey.Required),
		survey.WithValidator(survey.MinLength(3)),
		survey.WithValidator(survey.MaxLength(20)),
		survey.WithValidator(utils.ValidateProjectName),
	)

	if err != nil {
		log.Fatal(err)
	}

	containerPort := 0
	containerPortPrompt := &survey.Input{
		Message: "Enter the port of PHP : ",
	}

	err = survey.AskOne(
		containerPortPrompt, &containerPort,
		survey.WithValidator(survey.Required),
		survey.WithValidator(utils.ValidatePort),
	)

	if err != nil {
		log.Fatal(err)
	}

	utils.ClearTerminal()

	fmt.Print(`
 ____  _   _ ____  
|  _ \| | | |  _ \ 
| |_) | |_| | |_) |
|  __/|  _  |  __/ 
|_|   |_| |_|_|    

`)

	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║             Configuration              ║")
	fmt.Println("╠════════════════════════════════════════╣")
	fmt.Printf("║ Container name : %-21s ║\n", containerName)
	fmt.Printf("║ Port           : %-21d ║\n", containerPort)
	fmt.Println("║ Framework      : None                  ║")
	fmt.Println("║ Language       : PHP                   ║")
	fmt.Println("╚════════════════════════════════════════╝")

	var confirmResult bool

	confirmPrompt := &survey.Confirm{
		Message: "Is it okay to start building the environment with this configuration?",
	}

	if err := survey.AskOne(confirmPrompt, &confirmResult); err != nil {
		log.Fatal(err)
	}

	if !confirmResult {
		fmt.Println("Please try again")
		createProject()
	}
}
