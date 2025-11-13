package utils

import (
	"errors"
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
	configService, err := application.NewConfigService(container)

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

	container := infrastructure.NewDockerContainer()
	configService, err := application.NewConfigService(container)

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
	configService, err := application.NewConfigService(container)
	
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

	return nil
}

func ExtractionRepoName(repo string) string {
	url := strings.TrimSuffix(repo, ".git")

	repoName := path.Base(url)

	return repoName
}
