package utils

import (
	"fmt"
	"os/exec"
)

func UpWithBuild(path string) error {
	cmd := exec.Command("docker", "compose", "up", "-d", "--build")
	cmd.Dir = path

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Error running docker compose up -d --build: %v, output: %s", err, string(output))
	}

	return nil
}
