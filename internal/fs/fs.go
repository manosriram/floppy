package fs

import (
	"fmt"
	"os"

	"github.com/manosriram/floppy/internal/utils"
)

type FS struct {
	// Root string
	// MountPoints []string
	ShowHidden bool
}

func NewFS() FS {
	return FS{
		// Root:       root,
		ShowHidden: false,
	}
}

func (f FS) ReadDir(path string) ([]FileMetadata, error) {
	path = utils.CleanFilePath(path) // TODO: only proceed after this if f.Root exists in mountpoints
	var filesMetadata []FileMetadata
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		info, _ := entry.Info()

		// dotfile
		if !f.ShowHidden && len(entry.Name()) > 0 && entry.Name()[0] == '.' {
			continue
		}

		filesMetadata = append(filesMetadata, FileMetadata{
			IsDir: entry.IsDir(),
			Path:  fmt.Sprintf("%s/%s", path, entry.Name()),
			Name:  entry.Name(),
			Size:  info.Size(),
			// Type:  entry.Info(),
		})
	}

	return filesMetadata, nil
}
