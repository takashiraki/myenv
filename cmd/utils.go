package cmd

import (
	"errors"
	"fmt"
	"strconv"
)

func clearTerminal() {
	fmt.Println("\033c")
}

func validatePort(val interface{})error {
	str := val.(string)

	port,err := strconv.Atoi(str)

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