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

	fmt.Println(`
 ____  _   _ ____  
|  _ \| | | |  _ \ 
| |_) | |_| | |_) |
|  __/|  _  |  __/ 
|_|   |_| |_|_|    
`)

	fmt.Println("--------------------------------")
	fmt.Println("Configuration")
	fmt.Printf("Container name : %s\n", containerName)
	fmt.Println("--------------------------------")

	var confirmResult bool

	confirmPrompt := &survey.Confirm{
		Message: "Is the configuration correct?",
	}

	if err := survey.AskOne(confirmPrompt, &confirmResult); err != nil {
		log.Fatal(err)
	}

	if !confirmResult {
		fmt.Println("Please try again")
		createProject()
	}
}