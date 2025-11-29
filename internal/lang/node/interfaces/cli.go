package interfaces

import (
	"log"
	"myenv/internal/lang/node/nuxt/interfaces"

	"github.com/AlecAivazis/survey/v2"
)

func EntryPoint(fw string) {
	if fw != "" {
		adoptedFw := map[string]any{
			"Nuxt": interfaces.EntryPoint,
		}

		if _, ok := adoptedFw[fw]; ok {
			adoptedFw[fw].(func())()
		} else {
			log.Fatal("Unsupported framework selected.")
		}
	} else {
		fwPrompt := &survey.Select{
			Message: "Select the framework you want to use: ",
			Options: []string{"Nuxt"},
		}

		if err := survey.AskOne(fwPrompt, &fw); err != nil {
			log.Fatal(err)
		}

		switch fw {
		case "Nuxt":
			interfaces.EntryPoint()
		default:
			log.Fatal("Unsupported framework selected.")
		}
	}
}
