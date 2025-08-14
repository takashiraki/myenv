package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func clearTerminal() {
	fmt.Println("\033c")
}

func ShowLoadingIndicator(message string, done chan bool) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0

	for {
		select {
		case <-done:
			return
		default:
			fmt.Printf("\r%s %s", frames[i], message)
			i = (i + 1) % len(frames)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func DirIsExists(dir string) bool {
	_, err := os.Stat(dir)
	return !os.IsNotExist(err)
}

func CreateEnvFile(
	containerName string,
	targetPath string,
	srcPath string,
	dockerPath string,
	containerPort int,
	clientPort int,
) error {
	envExampleFilePath := filepath.Join(targetPath, ".env.example")
	envFilePath := filepath.Join(targetPath, ".env")

	if _, err := os.Stat(envExampleFilePath); os.IsNotExist(err) {
		return errors.New(".env.example file does not exist")
	}

	if _, err := os.Stat(envFilePath); err == nil {
		return nil
	}

	done := make(chan bool)
	go ShowLoadingIndicator("Creating .env file", done)

	cpCmd := exec.Command("cp", envExampleFilePath, envFilePath)
	err := cpCmd.Run()

	if err != nil {
		done <- true

		return fmt.Errorf("failed to copy .env.example to .env: %w", err)
	}

	done <- true

	fmt.Printf("\r\033[K.env file created ✓\n")

	done = make(chan bool)
	go ShowLoadingIndicator("Setting .env file", done)

	content, err := os.ReadFile(envFilePath)

	if err != nil {
		done <- true
		return fmt.Errorf("failed to read .env file: %w", err)
	}

	updateContent := string(content)

	updateContent = strings.ReplaceAll(updateContent, "CONTAINER_NAME=", "CONTAINER_NAME="+containerName)

	updateContent = strings.ReplaceAll(updateContent, "REPOSITORY=", "REPOSITORY="+srcPath)

	updateContent = strings.ReplaceAll(updateContent, "DOCKER_PATH=", "DOCKER_PATH="+dockerPath)

	updateContent = strings.ReplaceAll(updateContent, "HOST_PORT=", "HOST_PORT="+strconv.Itoa(clientPort))

	updateContent = strings.ReplaceAll(updateContent, "CONTAINER_PORT=", "CONTAINER_PORT="+strconv.Itoa(containerPort))

	err = os.WriteFile(envFilePath, []byte(updateContent), 0644)

	if err != nil {
		done <- true
		return fmt.Errorf("failed to write to .env file: %w", err)
	}

	done <- true
	fmt.Printf("\r\033[K.env file set ✓\n")

	return nil
}


func createDevContainerFile(containerName string, targetPath string) error {
	ch := make(chan bool)

	go ShowLoadingIndicator("Creating devcontainer.json file", ch)

	if !DirIsExists(targetPath) {
		ch <- true
		fmt.Printf("\r\033[Ktarget path does not exist ✓\n")
		return errors.New("target path does not exist")
	}

	devContainerExampleFilePath := filepath.Join(targetPath, "devcontainer.json.example")
	devContainerFilePath := filepath.Join(targetPath, "devcontainer.json")

	if _,err := os.Stat(devContainerExampleFilePath); os.IsNotExist(err){
		ch <- true
		fmt.Printf("\r\033[Kdevcontainer.json.example file does not exist ✓\n")
		return errors.New("devcontainer.json.example file does not exist")
	}

	if _,err := os.Stat(devContainerFilePath); err == nil {
		ch <- true
		fmt.Printf("\r\033[Kdevcontainer.json file already exists ✓\n")
		return nil
	}

	cpCmd := exec.Command("cp", devContainerExampleFilePath, devContainerFilePath)
	output,err := cpCmd.CombinedOutput()

	if err != nil {
		ch <- true
		fmt.Printf("\r\033[Kfailed to copy devcontainer.json.example to devcontainer.json: %v\nOutput: %s", err, output)
		return fmt.Errorf("failed to copy devcontainer.json.example to devcontainer.json: %w", err)
	}

	content, err := os.ReadFile(devContainerFilePath)

	if err != nil {
		ch <- true
		fmt.Printf("\r\033[Kfailed to read devcontainer.json file: %v\nOutput: %s", err, output)
		return fmt.Errorf("failed to read devcontainer.json file: %w", err)
	}

	contentStr := string(content)

	lines := strings.Split(contentStr, "\n")

	for i, line := range lines {

		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, `"name":`) {
			lines[i] = strings.Replace(line,trimmed, fmt.Sprintf(`"name": "%s",`,containerName),1)
			break
		}
	}

	modifiedContent := strings.Join(lines, "\n")
	if err := os.WriteFile(devContainerFilePath,[]byte(modifiedContent),0644); err != nil {
		ch <- true
		fmt.Printf("\r\033[Kfailed to write devcontainer.json file: %v\nOutput: %s", err, output)
		return fmt.Errorf("failed to write devcontainer.json file: %w", err)
	}

	ch <- true
	fmt.Printf("\r\033[Kdevcontainer.json file created ✓\n")

	return nil
}

func updateEnvSrcPath(projectPath string, srcPath string)error {
	envFilePath := filepath.Join(projectPath, ".env")

	ch := make(chan bool)

	go ShowLoadingIndicator("Updating .env file", ch)

	content, err := os.ReadFile(envFilePath)

	if err != nil {
		ch <- true
		fmt.Printf("\r\033[Kfailed to read .env file: %v\n", err)
		return fmt.Errorf("failed to read .env file: %w", err)
	}

	envContent := string(content)

	envContent = strings.ReplaceAll(envContent,"REPOSITORY=src","REPOSITORY="+srcPath)

	if err := os.WriteFile(envFilePath,[]byte(envContent),0644); err !=nil {
		ch <- true
		fmt.Printf("\r\033[kfailed to write .env file: %v\n",err)
		return fmt.Errorf("failed to write .env file: %w",err)
	}

	ch <- true
	fmt.Printf("\r\033[K.env file updated ✓\n")

	return nil
}