package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manosriram/floppy/internal/config"
	"github.com/manosriram/floppy/internal/fs"
)

type ReadDirRequest struct {
	Path        string   `json:"path"`
	Root        string   `json:"root"`
	MountPoints []string `json:"mountpoints"`
}

type FiberApiFileHandler struct {
	M config.Mountpoints
}

func NewFiberApiFileHandler(M config.Mountpoints) *FiberApiFileHandler {
	return &FiberApiFileHandler{
		M: M,
	}
}

func (h *FiberApiFileHandler) ReadDirHandler(c *fiber.Ctx) error {
	var reqData ReadDirRequest
	if err := c.BodyParser(&reqData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	reqData.MountPoints = c.Locals("mountpoints").([]string)

	fs := fs.NewFS()
	files, err := fs.ReadDir(reqData.Path)
	if err != nil {
	}

	return c.JSON(fiber.Map{
		"files": files,
	})
}
