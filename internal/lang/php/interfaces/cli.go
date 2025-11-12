package interfaces

import (
	"fmt"
	"log"
	"myenv/internal/framework"
	"myenv/internal/lang/php"
	"myenv/internal/lang/php/wordpress/interfaces"

	"github.com/AlecAivazis/survey/v2"
)

func EntryPoint(lang string, fw string) {
	if lang == "" {
			fmt.Println("Initializing your environment...")
			var selectedLang string

			langPrompt := &survey.Select{
				Message: "Select the language you want to use:",
				Options: []string{"PHP"},
			}

			if err := survey.AskOne(langPrompt, &selectedLang); err != nil {
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
					Options: []string{"None", "WordPress"},
				}

				if err := survey.AskOne(fwPrompt, &fw); err != nil {
					log.Fatal(err)
				}

				switch fw {
				case "None":
					php.PHP()
				case "WordPress":
					interfaces.EntryPoint()
				default:
					log.Fatal("Unsupported framework selected.")
				}
			}
		default:
			log.Fatal("Unsupported language selected.")
		}
}