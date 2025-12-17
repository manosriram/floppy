package api

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/manosriram/floppy/internal/config"
	"github.com/manosriram/floppy/internal/utils"
)

// TODO: Do not use hardcoded path
const (
	DEFAULT_THUMBS_IMG_PATH = "static/default_thumb.png"
)

type FiberApiThumbnailHandler struct {
	M config.Mountpoints
}

func NewFiberApiThumbnailHandler(M config.Mountpoints) *FiberApiThumbnailHandler {
	return &FiberApiThumbnailHandler{
		M: M,
	}
}

// TODO: Add path handling and checking
func (h *FiberApiThumbnailHandler) GetThumbnailHandler(c *fiber.Ctx) error {
	fileName, err := url.PathUnescape(c.Query("file"))
	if err != nil {
		return err
	}

	root, err := url.PathUnescape(c.Query("root"))
	if err != nil {
		return err
	}

	realRoot := h.M.MountPoints[root]

	ext := strings.Split(fileName, ".")
	extn := ext[len(ext)-1]
	c.Response().Header.Add("Content-Type", "image/jpeg")

	if extn == "jpg" || extn == "jpeg" || extn == "png" {
		thumbnailFile := fmt.Sprintf("%s/%s", realRoot, fileName)
		name, err := utils.ThumbSHA256(thumbnailFile)
		if err != nil {
			return err
		}
		onlyFileName := strings.Split(fileName, "/")

		thumbnailFilePath := fmt.Sprintf("%s/%s_%s", c.Locals("thumbsDir").(string), name, onlyFileName[len(onlyFileName)-1])
		return c.SendFile(thumbnailFilePath)
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	return c.SendFile(fmt.Sprintf("%s/%s", wd, DEFAULT_THUMBS_IMG_PATH))
}
