package infrastructure

import (
	"errors"
	"os/exec"
	"strings"
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

func (d *DockerContainer) ChechProxyNetworkExists() error {
	cmd := exec.Command("docker", "network", "ls", "--filter", "name=my_proxy_network")

	output, err := cmd.CombinedOutput()

	if err != nil {
		return errors.New("Error running docker network ls --filter name=my_proxy_network: " + err.Error() + ", output: " + string(output))
	}

	if !strings.Contains(string(output), "my_proxy_network") {
		return errors.New("my_proxy_network does not exist")
	}

	return nil
}

func (d *DockerContainer) ChechInfraNetworkExists() error {
	cmd := exec.Command("docker", "network", "ls", "--filter", "name=my_infra_network")

	output, err := cmd.CombinedOutput()

	if err != nil {
		return errors.New("Error running docker network ls --filter name=my_infra_network: " + err.Error() + ", output: " + string(output))
	}

	if !strings.Contains(string(output), "my_infra_network") {
		return errors.New("my_infra_network does not exist")
	}

	return nil
}

func (d *DockerContainer) CreateProxyNetwork() error {
	cmd := exec.Command("docker", "network", "create", "my_proxy_network")

	output, err := cmd.CombinedOutput()

	if err != nil {
		return errors.New("Error running docker network create my_proxy_network: " + err.Error() + ", output: " + string(output))
	}

	return nil
}

func (d *DockerContainer) CreateInfraNetwork() error {
	cmd := exec.Command("docker", "network", "create", "my_infra_network")

	output, err := cmd.CombinedOutput()

	if err != nil {
		return errors.New("Error running docker network create my_infra_network: " + err.Error() + ", output: " + string(output))
	}

	return nil
}