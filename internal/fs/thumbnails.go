package fs

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/manosriram/floppy/internal/utils"
)

// TODO: Clean this struct
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

func (t *Thumbnail) GenerateThumbnail() error {
	src, err := imaging.Open(t.FilePath)
	if err != nil {
		return err
	}
	thumbnail := imaging.Thumbnail(src, 500, 500, imaging.Lanczos) // 500x500 dim thumbnail
	thumbSHA, err := utils.ThumbSHA256(t.FilePath)
	if err != nil {
		return err
	}

	thumbnailFilePath := fmt.Sprintf("%s/%s_%s", t.ThumbsPath, thumbSHA, t.Name)
	err = imaging.Save(thumbnail, thumbnailFilePath)
	if err != nil {
		return err
	}

	return nil
}

// TODO: Optimize this to lazy generate thumbnails
func GenerateThumbnails(root string) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}
	thumbsDir := workingDir + "/.thumbs"

	ok, err := utils.PathExists(thumbsDir)
	if err != nil {
		return err
	}

	if !ok {
		if err := os.Mkdir(thumbsDir, 0o755); err != nil {
			log.Fatalf("Error creating .thumbs dir")
		}
	}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || path[0] == '.' {
			return nil
		}

		ext := strings.Split(path, ".")
		extn := ext[len(ext)-1]
		if extn == "jpg" || extn == "jpeg" || extn == "png" {
			// Check if thumbnail exists
			hash, err := utils.ThumbSHA256(path)
			if err != nil {
				return err
			}

			filePath := fmt.Sprintf("%s/%s_%s", thumbsDir, hash, d.Name())
			if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
				thumbnail := NewThumbnail(path, thumbsDir, extn, d.Name())
				thumbnail.GenerateThumbnail()
				return nil
			}
			return nil
		}
		return nil
	})
	return nil
}
