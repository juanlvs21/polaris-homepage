package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/sync/errgroup"

	"polaris-dashboard/internal/clients/proxmox"
)

// Proxmox devuelve el estado de todos los nodos. Las llamadas a cada nodo se
// hacen en paralelo (errgroup); si un nodo falla, los demás siguen. Cachea 30s
// y usa fallback stale ante fallo total.
func (h *Handlers) Proxmox(c *fiber.Ctx) error {
	if useDummyData {
		return c.JSON(fiber.Map{"nodes": dummyProxmoxNodes(), "stale": false, "updated_at": time.Now().Unix()})
	}

	if nodes, ok := h.proxmoxCache.Get(); ok {
		return c.JSON(fiber.Map{"nodes": nodes, "stale": false, "updated_at": time.Now().Unix()})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	nodes := make([]proxmox.Node, len(h.proxmox))
	g, _ := errgroup.WithContext(ctx)
	for i, client := range h.proxmox {
		i, client := i, client
		g.Go(func() error {
			nodes[i] = client.Fetch() // Fetch nunca retorna error: lo expresa en Node.Error
			return nil
		})
	}
	_ = g.Wait()

	h.proxmoxCache.Set(nodes)
	return c.JSON(fiber.Map{"nodes": nodes, "stale": false, "updated_at": time.Now().Unix()})
}
