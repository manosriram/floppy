package web

import (
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

type FiberWebHandler struct{}

func NewFiberWebHandler() FiberWebHandler {
	return FiberWebHandler{}
}

func (f *FiberWebHandler) Home(c *fiber.Ctx) error {
	return c.SendFile(filepath.Join("web", "index.html"))
}
