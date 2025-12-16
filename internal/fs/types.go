package fs

type FileMetadata struct {
	Path      string
	Data      []byte
	IsDir     bool
	Name      string
	Size      int64
	MountPath string
	Type      string
}
