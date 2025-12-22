package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/template/html/v2"
	"github.com/manosriram/floppy/handlers"
	"github.com/manosriram/floppy/internal/config"
	"github.com/manosriram/floppy/internal/fs"
	"go.uber.org/zap"
)

func generateThumbnails(m config.Mountpoints) {
	for _, mp := range m.MountPoints {
		go fs.GenerateThumbnails(mp)
	}
	zap.S().Infow("Completed thumbnail generation job")
}

func generateHlsSegments(m config.Mountpoints) {
	h := fs.NewHLS(m)
	h.GenerateHLSSegmentsForMountPoints()

	zap.S().Infow("HLS generation completed")
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	zap.ReplaceGlobals(logger)

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

	// Middlewares
	// app.Use(logger.New())

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
		c.Response().Header.Add("Access-Control-Allow-Origin", "*")
		return c.Next()
	})

	// Serve static assets from web/ (e.g. /styles.css)
	app.Static("/", "./static")
	app.Static("/hls", workingDir+".hls")

	// TODO: Make the API naming convention better
	app.Get("/api/thumb", h.ApiThumbnailHandler.GetThumbnailHandler)
	app.Post("/api/v1/fs/list", h.ApiFileHandler.ReadDirHandler)
	app.Post("/api/v1/fs/upload", h.ApiFileHandler.UploadDirHandler)
	app.Post("/api/v1/mountpoints/list", h.ApiMountpointHandler.ListMountPointsHandler)

	// Render template at root
	app.Get("/", w.WebHandler.Home)
	app.Get("/*", w.WebHandler.ReadMountDir)

	go func() {
		time.Sleep(1 * time.Second)
		go generateThumbnails(m)
		// go generateHlsSegments(m)
	}()

	log.Fatal(app.Listen(":5050"))
}
