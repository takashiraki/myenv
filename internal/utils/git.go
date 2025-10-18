package utils

import (
	"errors"
	"os/exec"
)

func CloneRepo(repoUrl string, targetPath string) error {
	command := exec.Command("git", "clone", repoUrl, targetPath)

	output, err := command.CombinedOutput()

	if err != nil {
		return errors.New("error cloning repository: " + string(output) + `\nPlease make sure that the repository URL is correct and you have access to it.\n` + err.Error())
	}

	return nil
}
