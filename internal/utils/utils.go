package utils

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ClearTerminal() {
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

func checkDir(dir string) error {
	if DirIsExists(dir) {
		return fmt.Errorf("directory %s already exists", dir)
	}

	return nil
}

func CreateEnvFile(projectPath string) error {
	envExampleFilePath := filepath.Join(projectPath, ".env.example")

	if _, err := os.Stat(envExampleFilePath); os.IsNotExist(err) {
		return errors.New(".env.example file does not exist")
	}

	envFilePath := filepath.Join(projectPath, ".env")

	if _, err := os.Stat(envFilePath); !os.IsNotExist(err) {
		fmt.Println(".env file already exists, skipping creation.")
		return nil
	}

	src, err := os.Open(envExampleFilePath)

	if err != nil {
		return errors.New("error opening .env.example file: " + err.Error())
	}

	defer src.Close()

	dst, err := os.Create(envFilePath)

	if err != nil {
		return errors.New("error creating .env file: " + err.Error())
	}

	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return errors.New("error copying .env.example to .env: " + err.Error())
	}

	return nil
}

func ReplaceValue(content *string, replacement map[string]interface{}) error {
	for key, value := range replacement {
		switch v := value.(type) {
		case string:
			*content = strings.ReplaceAll(*content, key, fmt.Sprintf("%s%s", key, v))
		case int:
			*content = strings.ReplaceAll(*content, key, fmt.Sprintf("%s%d", key, v))
		default:
			return errors.New("invalid type: value must be a string or integer")
		}
	}

	return nil
}
