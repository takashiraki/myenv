package cmd

import (
	"fmt"
	"os/exec"
)

func bootContainer(ymlPath string) error {
	ch := make(chan bool)

	go ShowLoadingIndicator("Booting container", ch)

	cmd := exec.Command("docker", "compose", "up", "-d")

	cmd.Dir = ymlPath

	output, err := cmd.CombinedOutput()

	if err != nil {
		ch <- true
		return fmt.Errorf("error booting container: %v\nOutput: %s", err, output)
	}

	ch <- true

	fmt.Printf("\r\033[kContainer booted ✓\n")

	return nil
}


func execCommand(command string, args []string, dir string,indicator string,completeMessage string)error {
	ch := make(chan bool)

	go ShowLoadingIndicator(indicator, ch)

	cmd := exec.Command(command,args...)
	cmd.Dir =dir

	output, err := cmd.CombinedOutput()

	if err != nil {
		ch <- true

		return fmt.Errorf("error executing command: %v\nOutput: %s", err, output)
	}

	ch <- true

	fmt.Printf("\r\033[k%s ✓\n",completeMessage)

	return nil
}