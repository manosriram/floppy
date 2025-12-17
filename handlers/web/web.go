package web

import (
	"errors"
	"net/url"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/manosriram/floppy/internal/config"
	"github.com/manosriram/floppy/internal/fs"
	"go.uber.org/zap"
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
	mountPath, err := url.PathUnescape(c.Params("*"))
	if err != nil {
		zap.S().Errorw("Error reading dir: ", "err", err.Error())
		return c.Render("mount", fiber.Map{
			"Error": err.Error(),
		})
	}

	mountPathSplit := strings.Split(mountPath, "/")
	rootName := mountPathSplit[0]
	relDir := strings.Join(mountPathSplit[1:], "/")

	if len(mountPathSplit) < 1 {
		return c.Render("mount", fiber.Map{
			"Files":        nil,
			"FileData":     nil,
			"MountPath":    mountPathSplit[0],
			"RequestPath":  c.Path(),
			"RequestParam": c.Params("*"),
			"Error":        errors.New("Error reading Dir"),
			"RootName":     rootName,
			"RelDir":       relDir,
		})
	}

	mountPath = f.M.MountPoints[mountPathSplit[0]] + "/" + strings.Join(mountPathSplit[1:], "/")

	d, err := os.ReadFile(mountPath)
	if err == nil {
		return c.SendFile(mountPath)
	}

	files, err := fs.NewFS().ReadDir(mountPath)
	if err != nil {
		zap.S().Errorw("Error reading dir: ", "err", err.Error())
		return c.Render("mount", fiber.Map{
			"Files":        files,
			"FileData":     nil,
			"MountPath":    mountPathSplit[0],
			"RequestPath":  c.Path(),
			"RequestParam": c.Params("*"),
			"Error":        errors.New("Error reading Dir"),
			"RootName":     rootName,
			"RelDir":       relDir,
		})
	}

	return c.Render("mount", fiber.Map{
		"Files":        files,
		"FileData":     d,
		"MountPath":    mountPathSplit[0],
		"RequestPath":  c.Path(),
		"RequestParam": c.Params("*"),
		"RootName":     rootName,
		"RelDir":       relDir,
	})
}
