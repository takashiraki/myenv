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
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", err)
		return
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
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", portErr)
		return
	}

	utils.ClearTerminal()

	repoName := utils.ExtractionRepoName(gitRepo)

	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", err)
		return
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

	fmt.Println("╔═════════════════════════════════════════════════════╗")
	fmt.Println("║                 Configuration                       ║")
	fmt.Println("╠═════════════════════════════════════════════════════╣")
	fmt.Printf("║ Container name : %-34s ║\n", containerName)
	fmt.Printf("║ Clone path     : %-34s ║\n", path)
	fmt.Printf("║ Port           : %-34d ║\n", containerPort)
	fmt.Println("║ Framework      : None                               ║")
	fmt.Println("║ Language       : PHP                                ║")
	fmt.Println("╚═════════════════════════════════════════════════════╝")

	var confirmResult bool

	confirmPrompt := &survey.Confirm{
		Message: "Is it okay to start building the environment with this configuration?",
	}

	if err := survey.AskOne(confirmPrompt, &confirmResult); err != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m %v\n", err)
		return
	}

	if !confirmResult {
		fmt.Println("\n\033[33mSetup cancelled.\033[0m Returning to configuration...")
		return
	}

	fw := "none"
	lang := "php"
	options := map[string]string{
		"type": "clone",
		"repo": gitRepo,
	}

	createConfigFile(
		containerName,
		containerPort,
		path,
		lang,
		fw,
		options,
	)

	targetRepo := "https://github.com/takashiraki/docker_php.git"

	done := make(chan bool)

	go utils.ShowLoadingIndicator("Cloning repository", done)

	if err := utils.CloneRepo(targetRepo, path); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to clone PHP project\n")

		// エラーの種類に応じた具体的なガイダンス
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "Repository not found") ||
			strings.Contains(errMsg, "fatal: repository"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Possible causes:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Repository URL is incorrect\n")
			fmt.Fprintf(os.Stderr, "   • Repository is private and requires authentication\n")
			fmt.Fprintf(os.Stderr, "   • Repository does not exist\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Check the repository URL and try again\n\n")

		case strings.Contains(errMsg, "Could not resolve host") ||
			strings.Contains(errMsg, "network"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Network issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Check your internet connection\n")
			fmt.Fprintf(os.Stderr, "   • Verify DNS settings\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Check network andtry again\n\n")

		case strings.Contains(errMsg, "git: command not found") ||
			strings.Contains(errMsg, "executable file not found"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Git is not installed:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Ubuntu/Debian: \033[36msudo apt install git\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • macOS: \033[36mbrew install git\033[0m\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Install git and try again\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\nDetails: %v\n", err)
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Fix the issue above and try again\n\n")
		}

		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)

		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	done <- true

	fmt.Printf("\r\033[KCloning repository completed \033[32m✓\033[0m\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Creating .env file", done)

	if err := utils.CreateEnvFile(path); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to create .env file\n")
		fmt.Fprintf(os.Stderr, "Details: %v\n\n", err)
		cleanUpFailedSetup(containerName, path)
		return
	}

	done <- true

	fmt.Printf("\r\033[KCreating .env file completed \033[32m✓\033[0m\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Clone your PHP project.", done)

	srcTargetPath := filepath.Join(path, "src", containerName)

	if err := utils.CloneRepo(gitRepo, srcTargetPath); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to clone PHP project\n")

		// エラーの種類に応じた具体的なガイダンス
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "Repository not found") ||
			strings.Contains(errMsg, "fatal: repository"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Possible causes:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Repository URL is incorrect\n")
			fmt.Fprintf(os.Stderr, "   • Repository is private and requires authentication\n")
			fmt.Fprintf(os.Stderr, "   • Repository does not exist\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Check the repository URL and try again\n\n")

		case strings.Contains(errMsg, "Authentication failed") ||
			strings.Contains(errMsg, "Permission denied"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Authentication required:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Set up SSH keys: ssh-keygen && cat ~/.ssh/id_rsa.pub\n")
			fmt.Fprintf(os.Stderr, "   • Or use HTTPS with personal access token\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Configure git credentials and try again\n\n")

		case strings.Contains(errMsg, "Could not resolve host") ||
			strings.Contains(errMsg, "network"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Network issue:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Check your internet connection\n")
			fmt.Fprintf(os.Stderr, "   • Verify DNS settings\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Check network andtry again\n\n")

		case strings.Contains(errMsg, "git: command not found") ||
			strings.Contains(errMsg, "executable file not found"):
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Git is not installed:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Ubuntu/Debian: \033[36msudo apt install git\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • macOS: \033[36mbrew install git\033[0m\n")
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Install git and try again\n\n")

		default:
			fmt.Fprintf(os.Stderr, "\nDetails: %v\n", err)
			fmt.Fprintf(os.Stderr, "\n\033[36m→ Next steps:\033[0m Fix the issue above and try again\n\n")
		}

		// クリーンアップの実行を明示
		fmt.Printf("\033[33mℹ Info:\033[0m Cleaning up partial setup...\n")
		cleanUpFailedSetup(containerName, path)

		// 再試行方法を明示
		fmt.Printf("\n\033[32m✓ Cleanup complete.\033[0m You can safely run this command again.\n\n")
		return
	}

	done <- true

	fmt.Printf("\r\033[KCloning your PHP project completed \033[32m✓\033[0m\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Setup env file", done)

	envFilePath := filepath.Join(path, ".env")

	repositoryPath := fmt.Sprintf("src/%s", containerName)
	hostPort := 80

	content, err := os.ReadFile(envFilePath)

	if err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to read .env file\n")
		fmt.Fprintf(os.Stderr, "Details: %v\n\n", err)
		return
	}

	updateContent := string(content)

	replacements := map[string]interface{}{
		"REPOSITORY_PATH=": fmt.Sprintf("REPOSITORY_PATH=%s", repositoryPath),
		"CONTAINER_NAME=":  fmt.Sprintf("CONTAINER_NAME=%s", containerName),
		"HOST_PORT=":       fmt.Sprintf("HOST_PORT=%d", containerPort),
		"CONTAINER_PORT=":  fmt.Sprintf("CONTAINER_PORT=%d", hostPort),
	}

	if err := utils.ReplaceAllValue(&updateContent, replacements); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to update .env file\n")
		fmt.Fprintf(os.Stderr, "Details: %v\n\n", err)
		return
	}

	if err := os.WriteFile(envFilePath, []byte(updateContent), 0644); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to write .env file\n")
		fmt.Fprintf(os.Stderr, "Details: %v\n\n", err)
		return
	}

	done <- true

	fmt.Printf("\r\033[KSetup .env file completed \033[32m✓\033[0m\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Creating container workspace", done)

	devContainerPath := filepath.Join(path, ".devcontainer", "devcontainer.json")

	devContainerExamplePath := filepath.Join(path, ".devcontainer", "devcontainer.json.example")

	if _, err := os.Stat(devContainerExamplePath); os.IsNotExist(err) {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m .devcontainer.json.example file does not exist\n")
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m .devcontainer.json.example file does not exist\n\n")
		cleanUpFailedSetup(containerName, path)
		return
	}

	utils.CopyFile(devContainerExamplePath, devContainerPath)

	devContainerContents, err := os.ReadFile(devContainerPath)

	if err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to read .devcontainer.json file\n")
		fmt.Fprintf(os.Stderr, "Details: %v\n\n", err)
		cleanUpFailedSetup(containerName, path)
		return
	}

	updateDevContainerContents := string(devContainerContents)

	replacements = map[string]interface{}{
		`"name": "php debug",`: fmt.Sprintf(`"name": "%s",`, containerName),
	}

	if err := utils.ReplaceAllValue(&updateDevContainerContents, replacements); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to update devcontainer.json file\n")
		fmt.Fprintf(os.Stderr, "Details: %v\n\n", err)
		cleanUpFailedSetup(containerName, path)
		return
	}

	if err := os.WriteFile(devContainerPath, []byte(updateDevContainerContents), 0644); err != nil {
		done <- true
		fmt.Printf("\r\033[K\033[31m✗ Error:\033[0m Failed to write devcontainer.json file\n")
		fmt.Fprintf(os.Stderr, "Details: %v\n\n", err)
		cleanUpFailedSetup(containerName, path)
		return
	}

	done <- true

	fmt.Printf("\r\033[KCreating container workspace completed \033[32m✓\033[0m\n")

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Starting Docker containers", done)

	if err := utils.UpWithBuild(path); err != nil {
		done <- true
		fmt.Printf("\r\033[KStarting Docker containers failed \033[31m✗\033[0m\n\n")

		errMsg := err.Error()

		cleanUpFailedSetup(containerName, path)

		switch {
		case strings.Contains(errMsg, "Cannot connect to the Docker daemon"):
			fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m Docker daemon is not running\n")
			fmt.Fprintf(os.Stderr, "\n\033[33m💡 Solution:\033[0m\n")
			fmt.Fprintf(os.Stderr, "   • Start Docker Desktop, or\n")
			fmt.Fprintf(os.Stderr, "   • Run: \033[36msudo systemctl start docker\033[0m\n\n")

		case strings.Contains(errMsg, "port is already allocated") || strings.Contains(errMsg, "address already in use"):
			fmt.Fprintf(os.Stderr, "Error: Port is already in use\n")
			fmt.Fprintf(os.Stderr, "Solution: Stop the conflicting container or choose a different port\n")

		case strings.Contains(errMsg, "permission denied"):
			fmt.Fprintf(os.Stderr, "Error: Permission denied\n")
			fmt.Fprintf(os.Stderr, "Solution: Add your user to the docker group or run with sudo\n")

		case strings.Contains(errMsg, "no space left"):
			fmt.Fprintf(os.Stderr, "Error: Insufficient disk space\n")
			fmt.Fprintf(os.Stderr, "Solution: Free up disk space or run 'docker system prune'\n")

		default:
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "\nTroubleshooting:\n")
			fmt.Fprintf(os.Stderr, "  • Check Docker: docker ps\n")
			fmt.Fprintf(os.Stderr, "  • View logs: cd %s && docker compose logs\n",
				path)
			fmt.Fprintf(os.Stderr, "  • Check .env: cat %s/.env\n", path)
		}

		return
	}

	done <- true
	fmt.Printf("\r\033[KStarting Docker containers completed \033[32m✓\033[0m\n")

	utils.ClearTerminal()

	fmt.Print(`
  ____  ___  __  __ ____  _     _____ _____ _____      /\   /\
 / ___|/ _ \|  \/  |  _ \| |   | ____|_   _| ____|    (  ._. )
| |   | | | | |\/| | |_) | |   |  _|   | | |  _|       > ^ <
| |___| |_| | |  | |  __/| |___| |___  | | | |___     /     \
 \____|\___/|_|  |_|_|   |_____|_____| |_| |_____|   /_______\

`)

	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                   🎉 SETUP COMPLETE! 🎉                          ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════════╣")
	fmt.Printf("║ 📦 Container Name : %-44s ║\n", containerName)
	fmt.Printf("║ 📂 Repository Path: %-44s ║\n", path)
	fmt.Printf("║ 🌐 Port          : %-45d ║\n", containerPort)
	fmt.Println("╠══════════════════════════════════════════════════════════════════╣")
	fmt.Println("║                          Next Steps:                             ║")
	fmt.Printf("║  • Open VS Code: code %-42s ║\n", path)
	fmt.Printf("║  • Access app  : http://localhost:%-30d ║\n", containerPort)
	fmt.Println("║  • Start coding in the devcontainer! 🚀                          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")

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

func cleanUpFailedSetup(containerName string, path string) {
	done := make(chan bool)

	go utils.ShowLoadingIndicator("Cleaning up config", done)
	if err := config.DeleteProjectConfig(containerName); err != nil {
		done <- true
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m Failed to remove project configuration: %v\n", err)
	} else {
		done <- true
		fmt.Printf("\r\033[KRemoved project configuration \033[32m✓\033[0m\n")
	}

	done = make(chan bool)

	go utils.ShowLoadingIndicator("Removing cloned repository", done)
	if err := os.RemoveAll(path); err != nil {
		done <- true
		fmt.Fprintf(os.Stderr, "\n\033[31m✗ Error:\033[0m Failed to remove cloned repository: %v\n", err)
	} else {
		done <- true
		fmt.Printf("\r\033[KRemoved cloned repository \033[32m✓\033[0m\n")
	}
}
