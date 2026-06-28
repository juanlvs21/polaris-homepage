package handlers

import "github.com/gofiber/fiber/v2"

// Weather devuelve el clima actual. Cachea 15 min; ante fallo de wttr.in
// devuelve el último valor conocido con `stale: true`.
func (h *Handlers) Weather(c *fiber.Ctx) error {
	if useDummyData {
		return c.JSON(fiber.Map{"weather": dummyWeather(h.cfg.Weather.Location, h.cfg.Weather.Units), "stale": false})
	}

	if w, ok := h.weatherCache.Get(); ok {
		return c.JSON(fiber.Map{"weather": w, "stale": false})
	}

	w, err := h.weather.Fetch()
	if err != nil {
		if stale, ok := h.weatherCache.Stale(); ok {
			return c.JSON(fiber.Map{"weather": stale, "stale": true})
		}
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
	}

	h.weatherCache.Set(w)
	return c.JSON(fiber.Map{"weather": w, "stale": false})
}
