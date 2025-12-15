package web

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/manosriram/floppy/internal/config"
	"github.com/manosriram/floppy/internal/fs"
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

	return c.Render("index", fiber.Map{
		"Title":       "Floppy",
		"MountPoints": mountPoints,
	})
}

func (f *FiberWebHandler) ReadMountDir(c *fiber.Ctx) error {
	mountPath := c.Params("*")

	mountPathSplit := strings.Split(mountPath, "/")
	mountPath = f.M.MountPoints[mountPathSplit[0]] + "/" + strings.Join(mountPathSplit[1:], "/")
	fmt.Println(mountPath)
	files, _ := fs.NewFS().ReadDir(mountPath)

	return c.Render("mount", fiber.Map{
		"Files":     files,
		"MountPath": mountPathSplit[0],
	})
}
