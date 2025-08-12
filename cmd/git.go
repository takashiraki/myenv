package cmd

import (
	"fmt"
	"os/exec"
)

func CloneRepository(repo string, dir string) error {
	done := make(chan bool)

	go ShowLoadingIndicator("コンテナを起動中", done)

	cmd := exec.Command("git", "clone", repo, dir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		done <- true
		return fmt.Errorf("\r\033[kerror cloning repository: %v\nOutput: %s", err, output)
	}

	done <- true

	fmt.Printf("\r\033[kコンテナの起動が完了しました ✓\n")
	return nil
}
