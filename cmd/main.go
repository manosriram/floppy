package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"github.com/manosriram/floppy/handlers"
	"github.com/manosriram/floppy/internal/config"
	"github.com/manosriram/floppy/internal/fs"
)

func generateThumbnails(m config.Mountpoints) {
	fmt.Println("Started thumbnail generation job")
	for _, mp := range m.MountPoints {
		go fs.GenerateThumbnails(mp)
	}
	fmt.Println("Completed thumbnail generation job")
}

func main() {
	// TODO: Add uber zap logger

	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working dir\n")
	}
	thumbsDir := workingDir + "/.thumbs"

	engine := html.New("./web", ".html")

	// Add template helper functions used by templates (e.g. mount.html breadcrumb)
	engine.AddFunc("split", strings.Split)
	engine.AddFunc("trim", strings.Trim)

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	m := config.NewMountPoints(fmt.Sprintf("%s/%s", workingDir, "config"))
	err = m.ReadMountPointsFromConfig()
	if err != nil {
		log.Fatalf("Error reading config file\n")
	}

	h := handlers.NewApiHandler(m)
	w := handlers.NewWebHandler(m)

	go generateThumbnails(m)

	// Middlewares
	app.Use(logger.New())

	app.Use(cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Query("noCache") == "true"
		},
		Expiration:   30 * time.Minute,
		CacheControl: true,
	}))

	// Middlware to set request context vars
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("mountpoints", m.MountPoints)
		c.Locals("thumbsDir", thumbsDir)
		return c.Next()
	})

	// Serve static assets from web/ (e.g. /styles.css)
	app.Static("/", "./static")

	// TODO: Make the API naming convention better
	app.Get("/api/thumb", h.ApiThumbnailHandler.GetThumbnailHandler)
	app.Post("/api/v1/fs/list", h.ApiFileHandler.ReadDirHandler)
	app.Post("/api/v1/mountpoints/list", h.ApiMountpointHandler.ListMountPointsHandler)

	// Render template at root
	app.Get("/", w.WebHandler.Home)
	app.Get("/*", w.WebHandler.ReadMountDir)

	log.Fatal(app.Listen(":5050"))
}
