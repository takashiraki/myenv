package infrastructure

import (
	"errors"
	"os/exec"
)

type GitRepository struct{}

func NewGitRepository() *GitRepository {
	return &GitRepository{}
}

func (d *GitRepository) CloneRepo(repoUrl string, targetPath string) error {
	cmd := exec.Command("git", "clone", repoUrl, targetPath)

	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.New("Error running git clone: " + err.Error() + ", output: " + string(output))
	}

	return nil
}
