package utils

import (
	"errors"
	"os"
	"path/filepath"
)

func ValidateProjectName(val any) error {
	str := val.(string)

	homeDir, err := os.UserHomeDir()

	if err != nil {
		return errors.New("error getting home directory")
	}

	devPath := filepath.Join(homeDir, "dev")

	targetPath := filepath.Join(devPath, str)

	if DirIsExists(targetPath) {
		return errors.New("project name already exists")
	}

	return nil
}

func DirIsExists(dir string) bool {
	_, err := os.Stat(dir)
	return !os.IsNotExist(err)
}
