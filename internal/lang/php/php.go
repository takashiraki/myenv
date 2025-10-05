package php

import (
	"fmt"
	"log"
	"myenv/internal/config"
	"myenv/internal/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

type PHPProjectDetail struct {
	Name    string            `json:"container_name"`
	Port    int               `json:"container_port"`
	Path    string            `json:"path"`
	Lang    string            `json:"lang"`
	Fw      string            `json:"framework"`
	Options map[string]string `json:"options"`
}

type PHPProject struct {
	Project PHPProjectDetail `json:"project"`
}

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
		cloneProject()
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
		survey.WithValidator(utils.ValidateDirectory),
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

	fw := "none"
	lang := "php"
	options := map[string]string{
		"type": "new",
	}

	homeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatal(err)
	}

	path := filepath.Join(homeDir, "dev", containerName)

	createConfigFile(containerName, containerPort, path, lang, fw, options)

	targetRepo := "https://github.com/takashiraki/docker_php.git"

	// Clone repository
	done := make(chan bool)

	go utils.ShowLoadingIndicator("Cloning repository", done)

	cmd := exec.Command("git", "clone", targetRepo, path)

	output, err := cmd.CombinedOutput()

	if err != nil {
		done <- true
		log.Fatalf("\r\033[Kerror cloning repository: %v\nOutput: %s", err, output)
	}

	done <- true

	fmt.Printf("\r\033[KCloning repository completed ✓\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Creating .env file", done)

	envExampleFilePath := filepath.Join(path, ".env.example")

	if _, err := os.Stat(envExampleFilePath); os.IsNotExist(err) {
		done <- true
		log.Fatalf("\r\033[Kerror: .env.example file does not exist in the repository")
	}

	envFilePath := filepath.Join(path, ".env")

	if _, err := os.Stat(envFilePath); err == nil {
		done <- true
		log.Fatalf("\r\033[Kerror: .env file already exists")
	}

	cmd = exec.Command("cp", envExampleFilePath, envFilePath)

	if _, err := cmd.CombinedOutput(); err != nil {
		done <- true
		log.Fatalf("\r\033[Kerror creating .env file: %v", err)
	}

	done <- true

	fmt.Printf("\r\033[KCreating .env file completed ✓\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Setup env file", done)

	repositoryPath := "src"
	port := 80

	content, err := os.ReadFile(envFilePath)

	if err != nil {
		done <- true
		log.Fatalf("\r\033[Kerror reading .env file: %v", err)
	}

	updateContent := string(content)
	updateContent = strings.ReplaceAll(updateContent, "REPOSITORY_PATH=", fmt.Sprintf("REPOSITORY_PATH=%s", repositoryPath))
	updateContent = strings.ReplaceAll(updateContent, "CONTAINER_NAME=", fmt.Sprintf("CONTAINER_NAME=%s", containerName))
	updateContent = strings.ReplaceAll(updateContent, "HOST_PORT=", fmt.Sprintf("HOST_PORT=%d", containerPort))
	updateContent = strings.ReplaceAll(updateContent, "CONTAINER_PORT=", fmt.Sprintf("CONTAINER_PORT=%d", port))

	if err := os.WriteFile(envFilePath, []byte(updateContent), 0644); err != nil {
		done <- true
		log.Fatalf("\r\033[Kerror writing .env file: %v", err)
	}

	done <- true

	fmt.Printf("\r\033[KSetup .env file completed ✓\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Starting Docker containers", done)

	cmd = exec.Command("docker", "compose", "up", "-d")

	cmd.Dir = path

	if _, err := cmd.CombinedOutput(); err != nil {
		done <- true
		log.Fatalf("\r\033[Kerror starting Docker containers: %v", err)
	}

	done <- true

	fmt.Printf("\r\033[KStarting Docker containers completed ✓\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Creating container workspace", done)

	devContainerExamplePath := filepath.Join(path, ".devcontainer", "devcontainer.json.example")
	devContainerPath := filepath.Join(path, ".devcontainer", "devcontainer.json")

	if _, err := os.Stat(devContainerExamplePath); os.IsNotExist(err) {
		log.Fatalf("error: .devcontainer.json.example file does not exist in the repository")
	}

	if _, err := os.Stat(devContainerPath); err == nil {
		log.Fatalf("error: .devcontainer.json file already exists")
	}

	cmd = exec.Command("cp", devContainerExamplePath, devContainerPath)

	if _, err := cmd.CombinedOutput(); err != nil {
		done <- true
		log.Fatalf("\r\033[Kerror creating .devcontainer.json file: %v", err)
	}

	devContainerContents, err := os.ReadFile(devContainerPath)

	if err != nil {
		done <- true
		log.Fatalf("\r\033[Kerror reading .devcontainer.json file: %v", err)
	}

	updateDevContainerContents := string(devContainerContents)

	updateDevContainerContents = strings.ReplaceAll(updateDevContainerContents, `"name": "php debug",`, fmt.Sprintf(`"name": "%s",`, containerName))

	if err := os.WriteFile(devContainerPath, []byte(updateDevContainerContents), 0644); err != nil {
		done <- true
		log.Fatalf("\r\033[Kerror writing .devcontainer.json file: %v", err)
	}

	done <- true
	fmt.Printf("\r\033[KCreating container workspace completed ✓\n")

	utils.ClearTerminal()

	fmt.Print(`
  ____  ___  __  __ ____  _     _____ _____ _____      /\   /\
 / ___|/ _ \|  \/  |  _ \| |   | ____|_   _| ____|    (  ._. )
| |   | | | | |\/| | |_) | |   |  _|   | | |  _|       > ^ <
| |___| |_| | |  | |  __/| |___| |___  | | | |___     /     \
 \____|\___/|_|  |_|_|   |_____|_____| |_| |_____|   /_______\

`)

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║                   🎉 SETUP COMPLETE! 🎉                ║")
	fmt.Println("╠════════════════════════════════════════════════════════╣")
	fmt.Printf("║ 📦 Container Name : %-34s ║\n", containerName)
	fmt.Printf("║ 📂 Repository Path: %-34s ║\n", path)
	fmt.Printf("║ 🌐 Port          : %-35d ║\n", containerPort)
	fmt.Println("╠════════════════════════════════════════════════════════╣")
	fmt.Println("║                     Next Steps:                        ║")
	fmt.Printf("║  • Open VS Code: code %-32s ║\n", path)
	fmt.Printf("║  • Access app  : http://localhost:%-20d ║\n", containerPort)
	fmt.Println("║  • Start coding in the devcontainer! 🚀                ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	codeVersionCommand := exec.Command("code", "--version")

	if _, err = codeVersionCommand.CombinedOutput(); err == nil {

		var openInVSCode bool

		openInVSCodePrompt := &survey.Confirm{
			Message: "Do you want to open the project in VS Code?",
		}

		if err := survey.AskOne(openInVSCodePrompt, &openInVSCode); err != nil {
			log.Fatal(err)
		}

		if openInVSCode {
			openCommand := exec.Command("code", path)

			if _, err := openCommand.CombinedOutput(); err != nil {
				log.Fatalf("error opening project in VS Code: %v", err)
			}
		}
	}
}

func createConfigFile(containerName string, containerPort int, path string, lang string, fw string, options map[string]string) {
	err := config.AddProjectConfig(containerName, containerPort, path, lang, fw, options)

	if err != nil {
		log.Fatal(err)
	}
}

func cloneProject() {
	gitRepo := ""
	gitRepoPrompt := &survey.Input{
		Message: "Enter the Git repository URL of PHP project : ",
	}

	err := survey.AskOne(
		gitRepoPrompt, &gitRepo,
		survey.WithValidator(survey.Required),
		survey.WithValidator(utils.ValidateGitRepoUrl),
		survey.WithValidator(utils.ValidateGitRepoProjectExists),
	)

	if err != nil {
		log.Fatal(err)
	}

	containerPort := 0
	containerPortPrompt := &survey.Input{
		Message: "Enter the port of PHP : ",
	}

	portErr := survey.AskOne(
		containerPortPrompt, &containerPort,
		survey.WithValidator(survey.Required),
		survey.WithValidator(utils.ValidatePort),
	)

	if portErr != nil {
		log.Fatal(portErr)
	}

	utils.ClearTerminal()

	repoName := utils.ExtractionRepoName(gitRepo)

	homeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatal(err)
	}

	path := filepath.Join(homeDir, "dev", repoName)
	containerName := repoName

	fmt.Print(`
 _____ _      ___  _   _ _____   ____  _   _ ____
/ ____| |    / _ \| \ | | ____| |  _ \| | | |  _ \
| |   | |   | | | |  \| |  _|   | |_) | |_| | |_) |
| |___| |___| |_| | |\  | |___  |  __/|  _  |  __/
\_____|_____|\___/|_| \_|_____| |_|   |_| |_|_|

`)

	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║              Configuration                       ║")
	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Printf("║ Container name : %-31s ║\n", containerName)
	fmt.Printf("║ Clone path : %-35s ║\n", path)
	fmt.Printf("║ Port           : %-31d ║\n", containerPort)
	fmt.Println("║ Framework      : None                            ║")
	fmt.Println("║ Language       : PHP                             ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
}
