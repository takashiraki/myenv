package infrastructure

import (
	"errors"
	"os/exec"
)

type DockerContainer struct{}

func NewDockerContainer() *DockerContainer {
	return &DockerContainer{}
}

func (d *DockerContainer) CreateContainer(path string) error {
	cmd := exec.Command("docker", "compose", "up", "-d", "--build")

	cmd.Dir = path

	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.New("Error running docker compose up -d --build: " + err.Error() + ", output: " + string(output))
	}

	return nil
}
