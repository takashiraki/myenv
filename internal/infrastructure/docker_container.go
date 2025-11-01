package infrastructure

import (
	"os/exec"
)

type DockerContainer struct{}

func NewDockerContainer() *DockerContainer {
	return &DockerContainer{}
}

func (d *DockerContainer) CreateContainer(path string) error {
	cmd := exec.Command("docker", "compose", "up", "-d", "--build")

	cmd.Dir = path

	if _, err := cmd.CombinedOutput(); err != nil {
		return err
	}

	return nil
}
