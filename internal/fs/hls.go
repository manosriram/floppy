package fs

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
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

func generateNSegmentsViaFfmpeg(inputFilePath, outDir, seek string, duration float64, N int) ([]byte, error) {
	seekNum, _ := strconv.ParseInt(seek, 10, 64)
	startNum := seekNum / 6

	ss := strconv.Itoa(int(startNum))
	fmt.Println("outDir = ", outDir)

	fmt.Println("generating to ", outDir)
	args := []string{
		"-hide_banner",
		"-y",
		"-ss", seek,
		"-i", "/Users/manosriram/Desktop/ffmpegtest/chess.mp4",

		"-t", strconv.Itoa(N * 6),

		// segment muxer
		"-c", "copy",
		"-f", "segment",
		"-segment_time", "6",
		"-reset_timestamps", "1",
		"-segment_start_number", ss,
		filepath.Join(outDir, "output%03d.ts"),
		// "/Users/manosriram/go/src/floppy/.hls/chess.mp4/output%03d.ts",
		// filepath.Join(outDir, "output%03d.ts"),
	}
	return exec.Command("ffmpeg", args...).CombinedOutput()
}

var segRe = regexp.MustCompile(`^output(\d+)\.ts$`)

func SegmentNumberFromTS(path string) (string, error) {
	base := filepath.Base(path) // e.g. "output005.ts"
	m := segRe.FindStringSubmatch(base)
	if m == nil {
		return "0", fmt.Errorf("not a segment filename: %q", base)
	}
	return m[1], nil
	// return strconv.Atoi(m[1]) // "005" -> 5
}

func (h *HLS) CreateM3U8(c *fiber.Ctx) error {
	f := c.Query("f")
	rootVsFile := strings.Split(f, ":")
	root, fileName := rootVsFile[0], rootVsFile[1]

	// TODO: handle this map access properly with thread-safe support
	// outputFileName := ss[len(ss)-1]

	videoSuffixes := []string{".mp4", ".mkv", ".webm", ".ogg", ".mov", ".m4v"}
	isVideoPath := false
	for _, videoSuffix := range videoSuffixes {
		isVideoPath = isVideoPath || strings.HasSuffix(fileName, videoSuffix)
	}

	wd, _ := os.Getwd()
	hlsPath := filepath.Join(wd, ".hls", fileName)
	// hlsPath := fmt.Sprintf("%s/.hls/%s", wd, fileName)
	fmt.Println(hlsPath)

	filePath := fmt.Sprintf("%s/%s", h.M.MountPoints[root], fileName)
	ss := strings.Split(filePath, "/")
	actualFilePath := strings.Join(ss[:len(ss)-1], "/")

	exists, err := utils.PathExists(filePath)
	if (!exists || err != nil) && isVideoPath {
		return c.JSON(fiber.Map{
			"error": "File does not exist",
		})
	}

	fmt.Println(filePath)
	getVideoDurationFromFfprobeArgs := []string{
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	}

	cmd := exec.Command("ffprobe", getVideoDurationFromFfprobeArgs...)
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	outString := strings.Trim(string(out), "\n")
	duration, err := strconv.ParseFloat(outString, 64)
	if err != nil {
		return err
	}

	fmt.Println("ffprobe out = ", duration)

	segmentDuration := 6

	segmentCount := int(duration) / segmentDuration

	m3u8FileTemplate := `
#EXTM3U
#EXT-X-VERSION:7
#EXT-X-TARGETDURATION:6
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-MAP:URI="%s"

	`

	m3u8FileTemplate = fmt.Sprintf(m3u8FileTemplate, filePath)
	for segment := range segmentCount {
		m3u8FileTemplate += "#EXTINF:6.000,\n"
		m3u8FileTemplate += fmt.Sprintf("%s:%s/output%d.ts\n", root, fileName, segment)
	}

	if isVideoPath {
		go generateNSegmentsViaFfmpeg(actualFilePath, hlsPath, "0", duration, 3)
	}

	// } else {
	// nn := strings.Split(fileName, "/")
	// f := fmt.Sprintf("%s/%s", h.M.MountPoints[root], nn[0])
	// f := strings.Join([]string{h.M.MountPoints[root], nn[0]}, "/")
	// f := nn[0]

	// wd, err := os.Getwd()
	// if err != nil {
	// }

	// segmentPath := fmt.Sprintf("%s/.hls/%s", wd, fileName)
	// cmd.Start()
	// if err != nil {
	// fmt.Println(err)
	// }

	// "ffmpeg -ss 30 -i chess.mp4 -c copy -f mpegts pipe:1"
	// TODO: figure out how to send segments
	// How to find which .ts files to send since we are naming it as segment*.ts
	if !isVideoPath {
		outFile := strings.Split(rootVsFile[1], "/")[1]
		fileP := filepath.Join(wd, ".hls", fileName)
		exists, _ := utils.PathExists(fileP)
		if exists {
			b, _ := os.ReadFile(fileP)
			return c.Send(b)
		}

		s, err := SegmentNumberFromTS(outFile)
		if err != nil {
			return err
		}
		// seek := getSeekSecondFromOutputNumber()
		generateNSegmentsViaFfmpeg(actualFilePath, hlsPath, s, duration, 3)

		c.Set("Content-Type", "video/MP2T")
		// var out []byte
		return c.SendString(outFile)
	}
	return c.SendString(m3u8FileTemplate)
}

func (h *HLS) GenerateHLSSegmentsForMountPoint(mountPoint string) {
	wd, err := os.Getwd()
	if err != nil {
	}

	fmt.Println(wd)
	// hlsDir := fmt.Sprintf("%s/.hls", wd)
	hlsDir := "/Users/manosriram/go/src/floppy/.hls"

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

		// args := []string{
		// "-i", path,
		// "-c:v", "libx264",
		// "-c:a", "libmp3lame",
		// "-b:a", "128k",
		// "-map", "0", // Map all streams (or use 0:v:0, 0:a:0)
		// "-f", "hls",
		// "-hls_time", "6", // Length of each segment
		// "-hls_list_size", "0", // Keep all segments in the playlist
		// "-hls_segment_filename", fmt.Sprintf("%s/%s/output%%03d.ts", hlsDir, d.Name()),
		// fmt.Sprintf("%s/%s/%s.m3u8", hlsDir, d.Name(), d.Name()), // Output playlist path
		// }

		if d.IsDir() || path[0] == '.' {
			return nil
		}

		ext := strings.Split(path, ".")
		extn := ext[len(ext)-1]
		if extn == "mp4" || extn == "webm" || extn == "ogg" || extn == "mov" || extn == "m4v" {
			zap.S().Infow("Generating HLS segments for media file", "filename", d.Name())
			hlsFilePath := fmt.Sprintf("%s/%s/%s.m3u8", hlsDir, d.Name(), d.Name())
			fmt.Println(hlsFilePath)
			// pathExists, err := utils.PathExists(hlsFilePath)
			// if err != nil {
			// return err
			// }

			hlsDir := fmt.Sprintf("%s/%s", hlsDir, d.Name())
			err = os.Mkdir(hlsDir, 0o755)
			if !os.IsExist(err) {
				// if err != nil {
				// zap.S().Errorw("Error creating .hls dir", "err", err.Error())
				// return err
				// }

				_, err = exec.Command("ffmpeg", args...).CombinedOutput()
				if err != nil {
					zap.S().Errorw("Error from ffmpeg shell call", "err", err.Error())
					return err
				}
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
