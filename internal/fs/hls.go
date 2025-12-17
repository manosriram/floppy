package fs

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/manosriram/floppy/internal/config"
	"github.com/manosriram/floppy/internal/utils"
	"go.uber.org/zap"
)

type HLS struct {
	M config.Mountpoints
}

func NewHLS(M config.Mountpoints) HLS {
	return HLS{
		M: M,
	}
}

func (h *HLS) GenerateHLSSegmentsForMountPoint(mountPoint string) {
	wd, err := os.Getwd()
	if err != nil {
	}

	hlsDir := fmt.Sprintf("%s/.hls", wd)

	_ = filepath.WalkDir(mountPoint, func(path string, d fs.DirEntry, err error) error {
		args := []string{
			"-i", path,
			"-c:v", "libx264",
			"-c:a", "libmp3lame",
			"-b:a", "128k",
			"-map", "0", // Map all streams (or use 0:v:0, 0:a:0)
			"-f", "hls",
			"-hls_time", "10", // Length of each segment
			"-hls_list_size", "0", // Keep all segments in the playlist
			"-hls_segment_filename", fmt.Sprintf("%s/%s/output%%03d.ts", hlsDir, d.Name()),
			fmt.Sprintf("%s/%s/%s.m3u8", hlsDir, d.Name(), d.Name()), // Output playlist path
		}

		if d.IsDir() || path[0] == '.' {
			return nil
		}

		ext := strings.Split(path, ".")
		extn := ext[len(ext)-1]
		if extn == "mp4" || extn == "webm" || extn == "ogg" || extn == "mov" || extn == "m4v" {
			zap.S().Infow("Generating HLS segments for media file", "filename", d.Name())
			hlsFilePath := fmt.Sprintf("%s/%s", hlsDir, d.Name())
			pathExists, err := utils.PathExists(hlsFilePath)
			if err != nil {
				return err
			}

			if !pathExists {
				err = os.Mkdir(hlsFilePath, 0o755)
				if err != nil {
					zap.S().Errorw("Error creating .hls dir", "err", err.Error())
					return err
				}
			}

			_, err = exec.Command("ffmpeg", args...).CombinedOutput()
			if err != nil {
				zap.S().Errorw("Error from ffmpeg shell call", "err", err.Error())
				return err
			}
		}
		return nil
	})
}

func (h *HLS) GenerateHLSSegmentsForMountPoints() {
	for _, mountPoint := range h.M.MountPoints {
		h.GenerateHLSSegmentsForMountPoint(mountPoint)
	}
}
