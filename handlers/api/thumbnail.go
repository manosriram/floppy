package api

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/manosriram/floppy/internal/config"
	"github.com/manosriram/floppy/internal/fs"
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

	fileName, _ := url.PathUnescape(c.Query("file"))
	root, _ := url.PathUnescape(c.Query("root"))
	// root := c.Query("root")
	realRoot := h.M.MountPoints[root]

	ext := strings.Split(fileName, ".")
	extn := ext[len(ext)-1]
	if extn == "jpg" || extn == "jpeg" || extn == "svg" || extn == "png" {
		name, err := fs.ThumbSHA256(realRoot + "/" + fileName)
		if err != nil {
			return err
		}
		onlyFileName := strings.Split(fileName, "/")
		name = c.Locals("thumbsDir").(string) + name + "_" + onlyFileName[len(onlyFileName)-1]

		c.Response().Header.Add("Content-Type", "image/jpeg")
		return c.SendFile(name)
	}

	return c.JSON(nil)
}
