package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

func ThumbSHA256(filePath string) (string, error) {
	abs, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}

	id := fmt.Sprintf("%s|%d|%d", abs, info.Size(), info.ModTime().UnixNano())
	sum := sha256.Sum256([]byte(id))
	name := hex.EncodeToString(sum[:])
	return name, nil
}
