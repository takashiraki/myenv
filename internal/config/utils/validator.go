package utils

import (
	"errors"
	"fmt"
	"myenv/internal/config/application"
	"myenv/internal/infrastructure"
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

	container := infrastructure.NewDockerContainer()
	repository := infrastructure.NewGitRepository()
	configService, err := application.NewConfigService(container, repository)

	if err != nil {
		return err
	}

	usedProjects, err := configService.GetProjects()

	if err != nil {
		return err
	}

	for _, usedProject := range usedProjects {
		if name == usedProject.ContainerName {
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
		return errors.New("invalid type: directory must be a string")
	}

	homeDir, err := os.UserHomeDir()

	if err != nil {
		return errors.New("error getting home directory")
	}

	targetDir := filepath.Join(homeDir, "dev", dir)

	if _, err := os.Stat(targetDir); err == nil {
		return errors.New("directory does not exist: " + targetDir)
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

	container := infrastructure.NewDockerContainer()
	repository := infrastructure.NewGitRepository()
	configService, err := application.NewConfigService(container, repository)

	if err != nil {
		return err
	}

	projects, err := configService.GetProjects()

	if err != nil {
		return err
	}

	for _, project := range projects {
		if proxy == project.ContainerProxy {
			return errors.New("proxy is already in use")
		}
	}

	return nil
}

func ValidateGitRepoUrl(val any) error {
	var repo string

	switch v := val.(type) {
	case string:
		repo = v
	default:
		return errors.New("invalid type: git repository must be a string")
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
		return errors.New("invalid type: git repository must be a string")
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

	container := infrastructure.NewDockerContainer()
	repository := infrastructure.NewGitRepository()
	configService, err := application.NewConfigService(container, repository)

	if err != nil {
		return err
	}

	projects, err := configService.GetProjects()

	if err != nil {
		return err
	}

	for _, project := range projects {
		if repoName == project.ContainerName {
			return errors.New("project name is already in use")
		}
	}

	if err := ValidateContainerExists(repoName); err != nil {
		return err
	}

	return nil
}

func ValidateContainerExists(val any) error {
	var name string

	switch v := val.(type) {
	case string:
		name = v
	default:
		return errors.New("invalid type: container name must be a string")
	}

	container := infrastructure.NewDockerContainer()

	output, err := container.ExecDockerCommand(
		"ps",
		"-a",
		"--filter",
		fmt.Sprintf("name=%s", name),
	)

	if err != nil {
		if strings.Contains(err.Error(), "Is the docker daemon running") {
			return errors.New("docker daemon is not running. Please start Docker and try again.")
		}

		return err
	}

	if strings.Contains(output, name) {
		return errors.New("container with the same name already exists. Please choose a different name.")
	}

	return nil
}

func ValidateDatabaseExists(val any) error {
	var db string

	switch v := val.(type) {
	case string:
		db = v
	default:
		return errors.New("invalid type: database name must be a string")
	}

	homeDir, err := os.UserHomeDir()

	if err != nil {
		return errors.New("error getting home directory")
	}

	mysqlEnvPath := filepath.Join(homeDir, "dev", "docker_mysql", ".env")

	if os.Stat(mysqlEnvPath); os.IsNotExist(err) {
		return errors.New("MySQL container is not set up. Please set up the MySQL container first.")
	}

	content, err := os.ReadFile(mysqlEnvPath)

	if err != nil {
		return errors.New("error reading MySQL .env file")
	}

	lines := strings.Split(string(content), "\n")

	var password string

	for _, line := range lines {
		if strings.HasPrefix(line, "MYSQL_ROOT_PASSWORD=") {
			parts := strings.SplitN(line, "=", 2)

			if len(parts) == 2 {
				password = parts[1]
			}
		}
	}

	if password == "" {
		return errors.New("MYSQL_ROOT_PASSWORD not found in .env file")
	}

	container := infrastructure.NewDockerContainer()

	command := fmt.Sprintf(
		"mysql -uroot -p%s -e \"SHOW DATABASES;\" | grep -w '%s'",
		password,
		db,
	)

	_, err = container.ExecCommand(
		"my_database",
		"sh",
		"-c",
		command,
	)

	if err == nil {
		return errors.New("database with the same name already exists. Please choose a different name.")
	}

	errStr := err.Error()

	if strings.Contains(errStr, "exit status 1") {
		return nil
	}

	if strings.Contains(errStr, "Is the docker daemon running") {
		return errors.New("docker daemon is not running. Please start Docker and try again.")
	}

	return errors.New("Error checking database existence: " + errStr)
}

func ExtractionRepoName(repo string) string {
	url := strings.TrimSuffix(repo, ".git")

	repoName := path.Base(url)

	return repoName
}
