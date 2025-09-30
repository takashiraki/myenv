package utils

import (
	"fmt"
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
