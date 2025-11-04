package utils

import (
	"fmt"
	"myenv/internal/infrastructure"
)

func CloneRepoHandle(
	p infrastructure.RepositoryInterface,
	repoUrl string,
	path string,
) error {
	done := make(chan bool)

	go ShowLoadingIndicator("Cloning repository", done)

	if err := p.CloneRepo(repoUrl, path); err != nil {
		done <- true

		fmt.Printf("\r\033[31m✗ Error:\033[0m Failed to clone repository\n")

		errMsg := err.Error()
	}

	done <- true
	return nil
}
