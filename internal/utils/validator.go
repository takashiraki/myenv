package utils

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
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

func ValidateProxy(val any) error {
	var proxy string

	switch v := val.(type) {
	case string:
		proxy = v
	case int:
		proxy = strconv.Itoa(v)
	default:
		return errors.New("invalid type: proxy must be a string or integer")
	}

	if !strings.HasSuffix(proxy, ".localhost") {
		return errors.New("proxy must end with .localhost")
	}

	domain := strings.TrimSuffix(proxy, ".localhost")

	if domain == "" {
		return errors.New("invalid proxy format")
	}

	usedProxies, err := getUsedProxy()

	if err != nil {
		return err
	}

	for _, usedusedProxy := range usedProxies {
		if proxy == usedusedProxy {
			return errors.New("proxy is already in use")
		}
	}

	return nil
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

func getUsedProxy() ([]string, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return nil, errors.New("error getting home directory")
	}

	targetPath := filepath.Join(homeDir, ".config", "myenv", "config.json")

	if os.Stat(targetPath); os.IsNotExist(err) {
		return nil, errors.New("config file does not exist")
	}

	data, err := os.ReadFile(targetPath)

	if err != nil {
		return nil, errors.New("error reading config file")
	}

	var config struct {
		Projects map[string]struct {
			ContainerProxy string `json:"container_proxy"`
		} `json:"projects"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.New("error parsing config file")
	}

	var usedProxys []string

	for _, project := range config.Projects {
		usedProxys = append(usedProxys, project.ContainerProxy)
	}

	return usedProxys, nil
}

func getUsedPort() ([]int, error) {
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

func ValidateGitRepoUrl(val any) error {
	var repo string

	switch v := val.(type) {
	case string:
		repo = v
	default:
		return errors.New("invalid type: git repository must be a string.")
	}

	u, err := url.Parse(repo)

	if err != nil {
		return errors.New("invalid git repository URL")
	}

	if u.Scheme != "https" {
		return errors.New("invalid git repository URL: must start with https://")
	}

	if !strings.HasSuffix(repo, ".git") {
		return errors.New("invalid git repository URL: must end with .git")
	}

	return nil
}

func ValidateGitRepoProjectExists(val any) error {
	var repo string

	switch v := val.(type) {
	case string:
		repo = v
	default:
		return errors.New("invalid type: git repository must be a string.")
	}

	repoName := ExtractionRepoName(repo)

	homeDIr, err := os.UserHomeDir()

	if err != nil {
		return errors.New("error getting home directory")
	}

	targetPath := filepath.Join(homeDIr, "dev", repoName)

	if DirIsExists(targetPath) {
		return errors.New("project with the same name already exists")
	}

	usedProjects, err := getUsedProject()

	if err != nil {
		return err
	}

	for _, usedProject := range usedProjects {
		if repoName == usedProject {
			return errors.New("project name is already in use")
		}
	}

	return nil
}

func ExtractionRepoName(repo string) string {
	url := strings.TrimSuffix(repo, ".git")

	repoName := path.Base(url)

	return repoName
}
