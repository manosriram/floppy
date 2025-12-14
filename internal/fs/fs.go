package fs

import (
	"fmt"
	"os"
)

type FS struct {
	Root string
}

func NewFS(root string) FS {
	return FS{
		Root: root,
	}
}

func (f FS) ReadPath(path string) ([]FileMetadata, error) {
	path = f.Root + path // TODO: only proceed after this if f.Root exists in mountpoints
	var filesMetadata []FileMetadata
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		info, _ := entry.Info()
		filesMetadata = append(filesMetadata, FileMetadata{
			IsDir: entry.IsDir(),
			Path:  fmt.Sprintf("%s/%s", path, entry.Name()),
			Name:  entry.Name(),
			Size:  info.Size(),
		})
	}

	return filesMetadata, nil
}
