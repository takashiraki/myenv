package utils

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
)

func ValidateProjectName(val any) error {
	var name string

	switch v := val.(type) {
	case string:
		name = v
	case int:
		name = strconv.Itoa(v)
	default:
		return errors.New("invalid type: project name must be a string or integer")
	}

	usedProjects, err := getUsedProject()

	if err != nil {
		return err
	}

	for _, usedProject := range usedProjects {
		if name == usedProject {
			return errors.New("project name is already in use")
		}
	}

	return nil
}

func ValidateDirectory(val any) error {
	var dir string

	switch v := val.(type) {
	case string:
		dir = v
	default:
		return errors.New("invalid type: directory must be a string.")
	}

	homeDir, err := os.UserHomeDir()

	if err != nil {
		return errors.New("error getting home directory")
	}

	targetDir := filepath.Join(homeDir, "dev", dir)

	if err := checkDir(targetDir); err != nil {
		return err
	}

	return nil
}

func DirIsExists(dir string) bool {
	_, err := os.Stat(dir)
	return !os.IsNotExist(err)
}

func ValidatePort(val any) error {
	var port int

	switch v := val.(type) {
	case int:
		port = v
	case string:
		var err error
		port, err = strconv.Atoi(v)
		if err != nil {
			return errors.New("invalid port format: must be a valid integer")
		}
	default:
		return errors.New("invalid type: port must be an integer or string")
	}

	if port < 1 || port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}

	if port < 1024 {
		return errors.New("port must be greater than 1024")
	}

	usedPorts, err := getUsedPort()

	if err != nil {
		return err
	}

	for _, usedPort := range usedPorts {
		if port == usedPort {
			return errors.New("port is already in use")
		}
	}

	return nil
}

func getUsedPort() ([]int, error) {
	homeDIr, err := os.UserHomeDir()

	if err != nil {
		return nil, errors.New("error getting home directory")
	}

	targetPath := filepath.Join(homeDIr, ".config", "myenv", "config.json")

	data, err := os.ReadFile(targetPath)

	if err != nil {
		return nil, errors.New("error reading config file")
	}

	var config struct {
		Projects map[string]struct {
			ContainerPort int `json:"container_port"`
		} `json:"projects"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.New("error parsing config file")
	}

	var usedPorts []int

	for _, project := range config.Projects {
		usedPorts = append(usedPorts, project.ContainerPort)
	}

	return usedPorts, nil
}

func getUsedProject() ([]string, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return nil, errors.New("error getting home directory")
	}

	targetPath := filepath.Join(homeDir, ".config", "myenv", "config.json")

	data, err := os.ReadFile(targetPath)

	if err != nil {
		return nil, errors.New("error reading config file")
	}

	var config struct {
		Projects map[string]struct {
			ContainerName string `json:"container_name"`
		} `json:"projects"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.New("error parsing config file")
	}

	var usedProjects []string

	for _, project := range config.Projects {
		usedProjects = append(usedProjects, project.ContainerName)
	}

	return usedProjects, nil
}
