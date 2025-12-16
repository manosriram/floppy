package fs

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
)

type Thumbnail struct {
	FilePath   string
	ThumbsPath string
	Extension  string
	Name       string
}

func NewThumbnail(filePath, thumbsPath, extn, name string) Thumbnail {
	return Thumbnail{
		FilePath:   filePath,
		ThumbsPath: thumbsPath,
		Extension:  extn,
		Name:       name,
	}
}

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

	// // ext should be something like ".jpg" or ".png"
	// return filepath.Join(thumbsDir, name+"."+ext), nil
}

func (t *Thumbnail) GenerateThumbnail() error {
	src, err := imaging.Open(t.FilePath)
	if err != nil {
		return err
	}
	thumbnail := imaging.Thumbnail(src, 50, 50, imaging.Lanczos)
	// name, err := thumbName(t.FilePath, t.ThumbsPath, t.Extension)
	thumbSHA, _ := ThumbSHA256(t.FilePath)
	err = imaging.Save(thumbnail, t.ThumbsPath+"/"+thumbSHA+"_"+t.Name)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (t *Thumbnail) GetThumbnailPathForFilePath(path, ext string) string {
	thumbSHA, _ := ThumbSHA256(path)
	return t.ThumbsPath + "/" + thumbSHA + "_" + t.Name
}
