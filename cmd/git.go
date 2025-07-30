package cmd

import "os/exec"

func CloneRepository(dir string) error {
	cmd := exec.Command("git", "clone", "https://github.com/laravel/laravel.git", dir)
	output, err := cmd.CombinedOutput()
	return nil
}
