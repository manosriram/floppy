package api

import (
	"fmt"
	"strings"

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

func (h *FiberApiFileHandler) UploadDirHandler(c *fiber.Ctx) error {
	// Get first file from form field "document":
	file, err := c.FormFile("file")
	path := c.FormValue("path")

	path = strings.Trim(path, " ")
	splitPath := strings.Split(path, "/")

	root := splitPath[1]
	actualRoot := h.M.MountPoints[root]
	actualPath := fmt.Sprintf("%s/%s", actualRoot, strings.Join(splitPath[2:], "/"))

	if err != nil {
		return err
	}
	return c.SaveFile(file, actualPath)
}
