package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// defaultIconSVG se sirve cuando un icono no existe en el CDN o falla la
// descarga. Glyph genérico de "app" (4 casillas) en gris neutro.
const defaultIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="#94a3b8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5"/></svg>`

// iconNamePattern descarta cualquier carácter fuera del slug seguro, evitando
// path traversal y nombres raros.
var iconNamePattern = regexp.MustCompile(`[^a-z0-9._-]`)

// iconFetchGroup evita que peticiones simultáneas del mismo icono disparen
// varias descargas al CDN a la vez.
var iconFetchGroup sync.Map // name -> *sync.Mutex

func sanitizeIconName(raw string) string {
	name := strings.ToLower(strings.TrimSpace(raw))
	name = strings.TrimSuffix(name, ".svg")
	name = iconNamePattern.ReplaceAllString(name, "")
	return name
}

// Icon sirve el icono de un servicio. Estrategia cache-first:
//  1. Si está en el directorio de caché, se sirve desde disco (local, rápido).
//  2. Si no, se descarga del CDN, se guarda en caché y se sirve.
//  3. Si el CDN no lo tiene o falla, se sirve un icono por defecto.
func (h *Handlers) Icon(c *fiber.Ctx) error {
	name := sanitizeIconName(c.Params("name"))
	if name == "" || name == "default" {
		return serveSVG(c, []byte(defaultIconSVG), true)
	}

	cacheDir := h.cfg.Icons.CacheDir
	cachePath := filepath.Join(cacheDir, name+".svg")

	// 1) Cache hit.
	if data, err := os.ReadFile(cachePath); err == nil {
		return serveSVG(c, data, false)
	}

	// Serializa la descarga por nombre para no golpear el CDN en paralelo.
	muAny, _ := iconFetchGroup.LoadOrStore(name, &sync.Mutex{})
	mu := muAny.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	// Otra goroutine pudo haberlo cacheado mientras esperábamos el lock.
	if data, err := os.ReadFile(cachePath); err == nil {
		return serveSVG(c, data, false)
	}

	// 2) Descarga del CDN.
	data, ok := h.fetchIcon(name)
	if !ok {
		// 3) Fallback: por defecto (no se cachea, para reintentar luego).
		return serveSVG(c, []byte(defaultIconSVG), true)
	}

	// Guarda en caché de forma atómica (best-effort).
	if err := os.MkdirAll(cacheDir, 0o755); err == nil {
		tmp := cachePath + ".tmp"
		if err := os.WriteFile(tmp, data, 0o644); err == nil {
			_ = os.Rename(tmp, cachePath)
		}
	}

	return serveSVG(c, data, false)
}

// fetchIcon descarga el SVG del CDN configurado.
func (h *Handlers) fetchIcon(name string) ([]byte, bool) {
	url := fmt.Sprintf(h.cfg.Icons.CDNURL, name)
	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, false
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024)) // tope de 512 KB
	if err != nil || len(data) == 0 {
		return nil, false
	}
	return data, true
}

func serveSVG(c *fiber.Ctx, data []byte, isDefault bool) error {
	c.Set("Content-Type", "image/svg+xml")
	if isDefault {
		// Corto, para reintentar pronto si el CDN se recupera.
		c.Set("Cache-Control", "public, max-age=300")
	} else {
		c.Set("Cache-Control", "public, max-age=86400")
	}
	return c.Send(data)
}
