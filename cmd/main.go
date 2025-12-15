package main

import (
	"log"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/manosriram/floppy/handlers"
	"github.com/manosriram/floppy/internal/config"
)

func main() {
	app := fiber.New()

	m := config.NewMountPoints("/Users/manosriram/go/src/floppy/config")
	err := m.ReadMountPointsFromConfig()
	if err != nil {
		// panic
		log.Fatalf("Error reading config file\n")
	}

	h := handlers.NewApiHandler(m)

	// Middlewares
	app.Use(logger.New())

	// Middlware to set request context vars
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("mountpoints", m.MountPoints)
		return c.Next()
	})

	// Serve index.html at root
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join("web", "index.html"))
	})

	// TODO: Make the API naming convention better
	app.Post("/api/v1/fs/list", h.ApiFileHandler.ReadDirHandler)
	app.Post("/api/v1/mountpoints/list", h.ApiMountpointHandler.ListMountPointsHandler)

	log.Fatal(app.Listen(":5050"))
}
