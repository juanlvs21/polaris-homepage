// Package handlers contiene los handlers HTTP de la API /api/*.
//
// Cada handler es de solo lectura y, cuando consume una API externa, cachea el
// resultado en memoria para no saturar a Proxmox/Arcane/wttr.in con el polling
// del frontend. Ante un fallo de la API externa se devuelve el último valor
// conocido marcado con `"stale": true`.
package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"polaris-dashboard/internal/cache"
	"polaris-dashboard/internal/clients/arcane"
	"polaris-dashboard/internal/clients/proxmox"
	"polaris-dashboard/internal/clients/weather"
	"polaris-dashboard/internal/config"
)

// Handlers agrupa la configuración, los clientes externos y los caches.
type Handlers struct {
	cfg *config.Config

	weather *weather.Client
	proxmox []*proxmox.Client
	arcane  []arcaneInstance

	weatherCache *cache.Entry[*weather.Weather]
	proxmoxCache *cache.Entry[[]proxmox.Node]
	dockerCache  *cache.Entry[[]dockerInstance]
}

type arcaneInstance struct {
	name   string
	client *arcane.Client
}

// New construye los handlers a partir de la configuración cargada.
func New(cfg *config.Config) *Handlers {
	h := &Handlers{
		cfg:          cfg,
		weather:      weather.New(cfg.Weather.Location, cfg.Weather.Units),
		weatherCache: cache.New[*weather.Weather](15 * time.Minute),
		proxmoxCache: cache.New[[]proxmox.Node](30 * time.Second),
		dockerCache:  cache.New[[]dockerInstance](30 * time.Second),
	}
	for _, p := range cfg.Proxmox {
		h.proxmox = append(h.proxmox, proxmox.New(p.Name, p.Host, p.TokenID, p.TokenSecret, p.VerifyTLS))
	}
	for _, a := range cfg.Arcane {
		if a.Host == "" {
			continue
		}
		name := a.Name
		if name == "" {
			name = a.Host
		}
		h.arcane = append(h.arcane, arcaneInstance{name: name, client: arcane.New(a.Host, a.APIKey)})
	}
	return h
}

// Register monta todas las rutas de la API en el router de Fiber.
func (h *Handlers) Register(app *fiber.App) {
	api := app.Group("/api")
	api.Get("/health", h.Health)
	api.Get("/branding", h.Branding)
	api.Get("/config", h.Services)
	api.Get("/weather", h.Weather)
	api.Get("/proxmox", h.Proxmox)
	api.Get("/docker", h.Docker)
	api.Get("/docker/:id/logs", h.DockerLogs)
	api.Get("/icons/:name", h.Icon)
}

// Health responde un simple ping de liveness.
func (h *Handlers) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// Branding expone la identidad y el sistema de diseño (whitelabeling). El
// frontend lo consume al arrancar para fijar el nombre, el favicon, el título
// y las CSS custom properties (colores y radios).
func (h *Handlers) Branding(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"branding": h.cfg.Branding,
		"theme":    h.cfg.Theme,
		"calendar": h.cfg.Calendar,
		"weather": fiber.Map{
			"units": h.cfg.Weather.Units,
		},
	})
}
