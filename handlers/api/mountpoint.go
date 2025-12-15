package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manosriram/floppy/internal/config"
)

type MountPointsRequest struct {
	Root string `json:"root"`
}

type FiberApiMountpointHandler struct {
	M config.Mountpoints
}

func NewFiberApiMountpointHandler(M config.Mountpoints) FiberApiMountpointHandler {
	return FiberApiMountpointHandler{
		M: M,
	}
}

func (h *FiberApiMountpointHandler) ListMountPointsHandler(c *fiber.Ctx) error {
	var reqData MountPointsRequest
	if err := c.BodyParser(&reqData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	return c.JSON(fiber.Map{
		"mountPoints": h.M.ListMountPoints(),
	})
}
