package interfaces

import (
	"log"
	NodeInterfaces "myenv/internal/lang/node/interfaces"
	"myenv/internal/lang/php/interfaces"

	"github.com/AlecAivazis/survey/v2"
)

func EntryPoint(lang string, fw string) {
	if lang == "" {
		var selectedLang string

		langPrompt := &survey.Select{
			Message: "Select the language you want to use:",
			Options: []string{"PHP", "JavaScript"},
		}

		if err := survey.AskOne(langPrompt, &selectedLang); err != nil {
			log.Fatal(err)
		}

		lang = selectedLang
	}

	switch lang {
	case "PHP":
		interfaces.EntryPoint(fw)
	case "JavaScript":
		NodeInterfaces.EntryPoint(fw)
	default:
		log.Fatal("Unsupported language selected.")
	}
}
