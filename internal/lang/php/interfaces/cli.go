package interfaces

import (
	"log"
	"myenv/internal/lang/php/none/interfaces/cli"
	"myenv/internal/lang/php/wordpress/interfaces"

	"github.com/AlecAivazis/survey/v2"
)

func EntryPoint(fw string) {

	if fw != "" {
		adoptedFw := map[string]any{
			"None": cli.EntryPoint,
			"WordPress": interfaces.EntryPoint,
		}

		if _, ok := adoptedFw[fw]; ok {
			adoptedFw[fw].(func())()
		} else {
			log.Fatal("Unsupported framework selected.")
		}

	} else {
		fwPrompt := &survey.Select{
			Message: "Select the framework you want to use: ",
			Options: []string{"None", "WordPress"},
		}

		if err := survey.AskOne(fwPrompt, &fw); err != nil {
			log.Fatal(err)
		}

		switch fw {
		case "None":
			cli.EntryPoint()
		case "WordPress":
			interfaces.EntryPoint()
		default:
			log.Fatal("Unsupported framework selected.")
		}
	}
}
