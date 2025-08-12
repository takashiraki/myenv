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

func validatePort(val interface{}) error {
	str := val.(string)

	port, err := strconv.Atoi(str)

	if err != nil {
		return errors.New("invalid port")
	}

	if port < 1 || port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}

	if port < 1024 {
		return errors.New("port must be greater than 1024")
	}

	return nil
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
	if os.IsNotExist(err) {
		return false
	}
	return true
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
	go ShowLoadingIndicator("環境ファイルを作成中", done)

	cpCmd := exec.Command("cp", envExampleFilePath, envFilePath)
	err := cpCmd.Run()

	if err != nil {
		done <- true

		return fmt.Errorf("failed to copy .env.example to .env: %w", err)
	}

	done <- true

	fmt.Printf("\r\033[k環境ファイルの作成が完了しました ✓\n")

	done = make(chan bool)
	go ShowLoadingIndicator("環境ファイルを設定中", done)

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
	fmt.Printf("\r\033[K環境ファイルの設定が完了しました ✓\n")

	return nil
}
