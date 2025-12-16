package utils

import (
	"os"
	"strings"
)

func CleanFilePath(path string) string {
	return strings.TrimRight(path, "/")
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err // permission error, etc.
}
