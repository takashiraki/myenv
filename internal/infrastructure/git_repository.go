package infrastructure

import "os/exec"

type GitRepository struct{}

func NewGitRepository() *GitRepository {
	return &GitRepository{}
}

func (d *GitRepository) CloneRepo(repoUrl string, targetPath string) error {
	cmd := exec.Command("git", "clone", repoUrl, targetPath)

	_, err := cmd.CombinedOutput()

	if err != nil {
		return err
	}

	return nil
}
