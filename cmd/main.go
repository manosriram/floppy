package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/manosriram/floppy/handlers"
)

func main() {
	app := fiber.New()

	h := handlers.NewFiberHandler()

	app.Post("/dir", h.ReadDirHandler)

	log.Fatal(app.Listen(":3000"))
}
