package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"polaris-dashboard/internal/config"
)

// serviceStatus es un acceso directo con su estado de disponibilidad.
type serviceStatus struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Icon     string `json:"icon"`
	Category string `json:"category"`
	Status   string `json:"status"` // online | offline | unknown
}

// Services devuelve los accesos directos del config con su estado, comprobado
// con un ping HTTP en paralelo (una goroutine por servicio, timeout 3s).
func (h *Handlers) Services(c *fiber.Ctx) error {
	if useDummyData {
		return c.JSON(fiber.Map{"services": dummyServices(h.cfg.Services)})
	}

	services := h.cfg.Services
	out := make([]serviceStatus, len(services))

	var wg sync.WaitGroup
	client := &http.Client{Timeout: 3 * time.Second}

	for i, svc := range services {
		out[i] = serviceStatus{
			Name: svc.Name, URL: svc.URL, Icon: svc.Icon,
			Category: svc.Category, Status: "unknown",
		}
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			out[i].Status = pingStatus(client, url)
		}(i, svc.URL)
	}
	wg.Wait()

	return c.JSON(fiber.Map{"services": out})
}

// pingStatus hace un GET y clasifica el resultado.
func pingStatus(client *http.Client, url string) string {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "unknown"
	}
	resp, err := client.Do(req)
	if err != nil {
		return "offline"
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return "online"
	}
	if resp.StatusCode >= 400 && resp.StatusCode < 600 {
		// El servicio responde aunque devuelva error de auth/forbidden: está vivo.
		return "online"
	}
	return "unknown"
}

// dummyServices toma los servicios del config y les asigna un estado simulado
// (en modo demo no se hace ping real porque las URLs .casa no resuelven).
func dummyServices(services []config.ServiceConfig) []serviceStatus {
	out := make([]serviceStatus, 0, len(services))
	for i, svc := range services {
		status := "online"
		if i%7 == 5 {
			status = "unknown"
		}
		if i%7 == 6 {
			status = "offline"
		}
		out = append(out, serviceStatus{
			Name:     svc.Name,
			URL:      svc.URL,
			Icon:     svc.Icon,
			Category: svc.Category,
			Status:   status,
		})
	}
	return out
}
