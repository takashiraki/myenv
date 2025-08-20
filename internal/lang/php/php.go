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
}
