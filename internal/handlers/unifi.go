package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/sync/errgroup"

	"polaris-dashboard/internal/clients/unifi"
)

// Unifi devuelve el estado de la(s) consola(s) UniFi: velocidad WAN, latencia,
// clientes conectados y salud del gateway. Las consolas se consultan en paralelo
// (errgroup); si una falla, las demás siguen. Cachea 20s con fallback stale.
func (h *Handlers) Unifi(c *fiber.Ctx) error {
	// Dummy mientras se desarrolla, o cuando se fuerza la vista de prueba sin
	// una consola UniFi configurada.
	if useDummyData || (len(h.unifi) == 0 && showUnifiDummy) {
		return c.JSON(fiber.Map{"routers": dummyUnifiRouters(), "stale": false, "updated_at": time.Now().Unix()})
	}

	if routers, ok := h.unifiCache.Get(); ok {
		return c.JSON(fiber.Map{"routers": routers, "stale": false, "updated_at": time.Now().Unix()})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	routers := make([]unifi.Router, len(h.unifi))
	g, _ := errgroup.WithContext(ctx)
	for i, client := range h.unifi {
		i, client := i, client
		g.Go(func() error {
			routers[i] = client.Fetch() // Fetch nunca retorna error: lo expresa en Router.Error
			return nil
		})
	}
	_ = g.Wait()

	h.unifiCache.Set(routers)
	return c.JSON(fiber.Map{"routers": routers, "stale": false, "updated_at": time.Now().Unix()})
}
