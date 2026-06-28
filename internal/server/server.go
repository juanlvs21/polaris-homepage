// Package server arma el servidor HTTP de Fiber: monta la API y sirve el
// frontend embebido (SPA), con fallback a index.html para rutas del cliente.
package server

import (
	"io/fs"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"polaris-dashboard/internal/config"
	"polaris-dashboard/internal/handlers"
)

// New construye la app de Fiber lista para escuchar.
func New(cfg *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               cfg.Branding.Name,
		DisableStartupMessage: true,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(compress.New())

	// API
	handlers.New(cfg).Register(app)

	// Frontend embebido
	mountFrontend(app)

	return app
}

// mountFrontend sirve el contenido de `dist` embebido. Las rutas no resueltas
// por la API ni por un archivo estático devuelven index.html (SPA fallback).
func mountFrontend(app *fiber.App) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Fatalf("no se pudo abrir el frontend embebido: %v", err)
	}

	app.Use("/", filesystem.New(filesystem.Config{
		Root:         http.FS(sub),
		Index:        "index.html",
		NotFoundFile: "index.html",
	}))
}
