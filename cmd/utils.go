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

	fmt.Printf("\r\033[k.env file created ✓\n")

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
