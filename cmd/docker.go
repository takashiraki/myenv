package cmd

import (
	"fmt"
	"os/exec"
)

func bootContainer(ymlPath string) error {
	cmd := exec.Command("docker", "compose", "up", "-d", "--build")

	cmd.Dir = ymlPath

	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("error booting container: %v\nOutput: %s", err, output)
	}

	return nil
}
