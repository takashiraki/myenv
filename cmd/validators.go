package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
)


func validatePort(val interface{}) error {
	str := val.(string)

	port, err := strconv.Atoi(str)

	if err != nil {
		return errors.New("invalid port")
	}

	if port < 1 || port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}

	if port < 1024 {
		return errors.New("port must be greater than 1024")
	}

	return nil
}

func validateProjectName(val interface{}) error {
	str := val.(string)

	homeDir, err := os.UserHomeDir();

	if(err != nil){
		return errors.New("error getting home directory")
	}

	devPath := filepath.Join(homeDir,"dev")
	
	targetPath := filepath.Join(devPath, str)

	if DirIsExists(targetPath) {
		return  errors.New("project name already exists")
	}

	return nil
}