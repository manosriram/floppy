package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"github.com/manosriram/floppy/handlers"
	"github.com/manosriram/floppy/internal/config"
)

func main() {
	// Template engine: renders files in ./web with .html extension
	engine := html.New("./web", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	m := config.NewMountPoints("/Users/manosriram/go/src/floppy/config")
	err := m.ReadMountPointsFromConfig()
	if err != nil {
		// panic
		log.Fatalf("Error reading config file\n")
	}

	h := handlers.NewApiHandler(m)
	w := handlers.NewWebHandler(m)

	// Middlewares
	app.Use(logger.New())

	// Middlware to set request context vars
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("mountpoints", m.MountPoints)
		return c.Next()
	})

	// Serve static assets from web/ under /static (e.g. /static/styles.css)
	// This avoids overriding the "/" route with the raw index.html file.
	app.Static("/static", "./web")

	// Render template at root
	app.Get("/", w.WebHandler.Home)

	// TODO: Make the API naming convention better
	app.Post("/api/v1/fs/list", h.ApiFileHandler.ReadDirHandler)
	app.Post("/api/v1/mountpoints/list", h.ApiMountpointHandler.ListMountPointsHandler)

	log.Fatal(app.Listen(":5050"))
}
