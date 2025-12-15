package web

import (
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/manosriram/floppy/internal/config"
)

type FiberWebHandler struct {
	M config.Mountpoints
}

func NewFiberWebHandler(M config.Mountpoints) FiberWebHandler {
	return FiberWebHandler{
		M: M,
	}
}

func (f *FiberWebHandler) Home(c *fiber.Ctx) error {
	mountPoints := f.M.ListMountPoints()

	return c.SendFile(filepath.Join("web", "index.html"))
}
