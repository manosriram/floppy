package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manosriram/floppy/internal/fs"
)

type ReadDirRequest struct {
	Path string `json:"path"`
	Root string `json:"root"`
}

type FiberHandler struct{}

func NewFiberHandler() *FiberHandler {
	return &FiberHandler{}
}

func (h *FiberHandler) ReadDirHandler(c *fiber.Ctx) error {
	var reqData ReadDirRequest
	if err := c.BodyParser(&reqData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	fs := fs.NewFS(reqData.Root)
	files, _ := fs.ReadPath(reqData.Path)

	return c.JSON(fiber.Map{
		"files": files,
	})
}
