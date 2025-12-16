package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"github.com/manosriram/floppy/handlers"
	"github.com/manosriram/floppy/internal/config"
	fss "github.com/manosriram/floppy/internal/fs"
)

const (
	THUMB_DIR = "/Users/manosriram/go/src/floppy/.thumbs/"
)

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err // permission error, etc.
}

func walkdir(root string) {
	// root := "/some/path"

	thumbsDir := "/Users/manosriram/go/src/floppy/.thumbs"
	ok, _ := pathExists(thumbsDir)
	if !ok {
		if err := os.Mkdir(thumbsDir, 0o755); err != nil {
			log.Fatalf("Error creating .thumbs dir")
		}
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// e.g. permission error; decide whether to stop or skip
			return err
		}

		if d.IsDir() || path[0] == '.' {
			// fmt.Println("DIR :", path)
			return nil
		}

		ext := strings.Split(path, ".")
		extn := ext[len(ext)-1]
		if extn == "jpg" || extn == "jpeg" || extn == "svg" || extn == "png" {
			thb := fss.NewThumbnail(path, thumbsDir, extn, d.Name())
			// fmt.Println(thb.GenerateThumbnail())
			thb.GenerateThumbnail()
		}
		// fmt.Println(extn)
		// fmt.Println("FILE:", path)
		return nil
	})

	if err != nil {
		panic(err)
	}
}

func main() {

	// Template engine: renders files in ./web with .html extension
	engine := html.New("./web", ".html")

	// Add template helper functions used by templates (e.g. mount.html breadcrumb)
	engine.AddFunc("split", strings.Split)
	engine.AddFunc("trim", strings.Trim)

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
		c.Locals("thumbsDir", THUMB_DIR)
		return c.Next()
	})

	for _, mp := range m.MountPoints {
		walkdir(mp)
	}

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
