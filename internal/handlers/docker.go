package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"polaris-dashboard/internal/clients/arcane"
)

type dockerInstance struct {
	Name       string             `json:"name"`
	Host       string             `json:"host,omitempty"`
	Containers []arcane.Container `json:"containers"`
	Error      string             `json:"error,omitempty"`
}

// Docker devuelve los contenedores desde Arcane. Cachea 30s; fallback stale.
func (h *Handlers) Docker(c *fiber.Ctx) error {
	if useDummyData {
		instances := dummyDockerInstances()
		return c.JSON(fiber.Map{"instances": instances, "containers": flattenContainers(instances), "stale": false})
	}

	if len(h.arcane) == 0 {
		return c.JSON(fiber.Map{"instances": []any{}, "containers": []any{}, "stale": false})
	}
	if instances, ok := h.dockerCache.Get(); ok {
		return c.JSON(fiber.Map{"instances": instances, "containers": flattenContainers(instances), "stale": false})
	}

	instances := make([]dockerInstance, 0, len(h.arcane))
	for _, a := range h.arcane {
		cs, err := a.client.Containers()
		instance := dockerInstance{Name: a.name, Containers: cs}
		if err != nil {
			instance.Error = err.Error()
		}
		instances = append(instances, instance)
	}

	h.dockerCache.Set(instances)
	return c.JSON(fiber.Map{"instances": instances, "containers": flattenContainers(instances), "stale": false})
}

// dockerActions son las acciones de ciclo de vida permitidas sobre un contenedor.
var dockerActions = map[string]bool{"start": true, "stop": true, "restart": true}

// DockerAction ejecuta start/stop/restart sobre un contenedor y luego invalida
// el cache para que el siguiente /api/docker refleje el nuevo estado.
func (h *Handlers) DockerAction(c *fiber.Ctx) error {
	action := c.Params("action")
	if !dockerActions[action] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "acción no soportada"})
	}

	if useDummyData {
		return c.JSON(fiber.Map{"ok": true})
	}
	if len(h.arcane) == 0 {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "arcane no configurado"})
	}

	id := c.Params("id")
	instanceName := c.Query("instance")

	var lastErr error
	for _, a := range h.arcane {
		if instanceName != "" && a.name != instanceName {
			continue
		}
		if err := a.client.Action(id, action); err == nil {
			h.dockerCache.Invalidate()
			return c.JSON(fiber.Map{"ok": true})
		} else {
			lastErr = err
		}
	}
	if lastErr != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": lastErr.Error()})
	}
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "instancia arcane no encontrada"})
}

// DockerLogs devuelve las últimas líneas de logs de un contenedor (sin cache).
func (h *Handlers) DockerLogs(c *fiber.Ctx) error {
	if useDummyData {
		tail := parseTail(c.Query("tail"))
		return c.JSON(fiber.Map{"logs": dummyLogs(c.Params("id"), tail)})
	}

	if len(h.arcane) == 0 {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "arcane no configurado"})
	}
	id := c.Params("id")
	instanceName := c.Query("instance")

	tail := parseTail(c.Query("tail"))

	var lastErr error
	for _, a := range h.arcane {
		if instanceName != "" && a.name != instanceName {
			continue
		}
		logs, err := a.client.Logs(id, tail)
		if err == nil {
			return c.JSON(fiber.Map{"logs": logs})
		}
		lastErr = err
	}
	if lastErr != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": lastErr.Error()})
	}
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "instancia arcane no encontrada"})
}

func parseTail(raw string) int {
	tail := 50
	if raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			tail = n
		}
	}
	if tail < 1 {
		tail = 50
	}
	if tail > 200 {
		tail = 200
	}
	return tail
}

func flattenContainers(instances []dockerInstance) []arcane.Container {
	total := 0
	for _, instance := range instances {
		total += len(instance.Containers)
	}
	out := make([]arcane.Container, 0, total)
	for _, instance := range instances {
		out = append(out, instance.Containers...)
	}
	return out
}
