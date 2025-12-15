package utils

import "strings"

func CleanFilePath(path string) string {
	return strings.TrimRight(path, "/")
}
