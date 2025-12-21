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

func generateNSegmentsViaFfmpeg(inputFilePath, outDir, seek string, N int) ([]byte, error) {
	seekNum, _ := strconv.ParseInt(seek, 10, 64)
	startNum := (6 * seekNum) / 6

	ss := strconv.Itoa(int(startNum))
	fmt.Println(seek, seekNum, startNum)
	fmt.Println("zz ", ss)
	fmt.Println("inDir = ", inputFilePath)
	fmt.Println("outDir = ", outDir)

	seeeek := strconv.Itoa(int(seekNum) * N)

	fmt.Println("generating to ", outDir)
	args := []string{
		"-hide_banner",
		"-y",
		"-ss", seeeek,
		"-i", "/Users/manosriram/Desktop/beatles.mp4",

		"-t", "6",

		// segment muxer
		"-c", "copy",
		"-f", "segment",
		"-vsync", "cfr", // Force constant frame rate
		"-map", "0:v:0", // Explicitly map first video stream
		"-map", "0:a:0", // Explicitly map first audio stream
		"-g", "60", // Keyframe every 2 seconds (crucial for HLS)
		"-r", "30", // Keyframe every 2 seconds (crucial for HLS)
		"-hls_time", "6", // Match your 6.000 duration in manifest
		"-hls_list_size", "0",
		"-async", "1", // Sync audio start
		"-ar", "44100", // Force audio sample rate to 44.1kHz (Standard)
		"-hls_playlist_type", "vod",
		"-segment_time", "6",
		"-reset_timestamps", "1",
		"-segment_start_number", ss,
		fmt.Sprintf("%s/%s", outDir, "output%3d.ts"),
		// outDir,
		// "/Users/manosriram/go/src/floppy/.hls/beatles.mp4",
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
	fmt.Println("isvid = ", isVideoPath)

	wd, _ := os.Getwd()
	hlsPath := filepath.Join(wd, ".hls", fileName)
	// hlsPath := fmt.Sprintf("%s/.hls/%s", wd, fileName)

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
	if !isVideoPath {
		l := strings.Split(filePath, "/")
		filePath = strings.Join(l[0:len(l)-1], "/")
	}

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

	fmt.Println(rootVsFile)

	fileP := filepath.Join(wd, ".hls", fileName)
	fmt.Println(fileP) // /Users/manosriram/go/src/floppy/.hls/beatles.mp4/output0.ts
	z := strings.Split(fileP, "/")
	if len(z) < 1 {
		return nil
	}
	// zz := strings.Join(z[:len(z)-1], "/")
	exists, _ = utils.PathExists(hlsPath)
	if exists && !isVideoPath {
		fmt.Println("hlll = ", hlsPath)
		b, _ := os.ReadFile(hlsPath)
		return c.Send(b)
	}

	var outFile string
	var s string = "0"
	if !isVideoPath {
		outFile = strings.Split(rootVsFile[1], "/")[1]
		s, err = SegmentNumberFromTS(outFile)
		if err != nil {
			return err
		}
	} else {
		outFile = ""
	}

	if isVideoPath {
		err = os.MkdirAll(hlsPath, 0o755)
		if err != nil {
			return err
		}

		m3u8FileTemplate := `
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:7
#EXT-X-MEDIA-SEQUENCE:0
`
		// m3u8FileTemplate = fmt.Sprintf(m3u8FileTemplate)
		for segment := range segmentCount {
			// for segment := range segmentCount + 1 {
			m3u8FileTemplate += "#EXTINF:6.000,\n"
			m3u8FileTemplate += fmt.Sprintf("http://localhost:5050/hls?f=%s:%s/output%03d.ts\n", root, fileName, segment)
		}
		m3u8FileTemplate += "#EXT-X-ENDLIST"

		for i := range 10 {
			si, _ := strconv.Atoi(s) // "005" -> 5
			si += 1
			s = strconv.Itoa(si)

			generateNSegmentsViaFfmpeg(actualFilePath, hlsPath, s, i)
		}
		return c.SendString(m3u8FileTemplate)
	}

	fmt.Println("not vid")

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

	// si, _ := strconv.Atoi(s) // "005" -> 5

	// startTime := si * 6
	// startTimeStr := strconv.Itoa(startTime)
	// fmt.Println("s ss = ", startTimeStr)

	// _, err = generateNSegmentsViaFfmpeg(actualFilePath, zz, s, 5)
	// if err != nil {
	// return err
	// }

	data, _ := os.ReadFile(hlsPath)

	return c.Send(data)
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
