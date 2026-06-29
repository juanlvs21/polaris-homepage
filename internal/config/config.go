// Package config carga y valida la configuración de la aplicación.
//
// La fuente de verdad es un archivo YAML (por defecto `config.yaml`). Los
// secretos (tokens, API keys) NO se escriben en el YAML: se referencian con la
// sintaxis `${VAR_NAME}` y se resuelven desde variables de entorno en tiempo de
// carga. Esto mantiene el archivo de configuración seguro para versionar.
//
// Todo el branding y el theming (nombre del sitio, colores, radios) viven aquí,
// de modo que el mismo binario sirve cualquier marca con solo cambiar el YAML.
package config

import (
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/viper"
)

// Config es la raíz de la configuración de la aplicación.
type Config struct {
	Server   ServerConfig    `mapstructure:"server"`
	Branding BrandingConfig  `mapstructure:"branding"`
	Theme    ThemeConfig     `mapstructure:"theme"`
	Weather  WeatherConfig   `mapstructure:"weather"`
	Calendar CalendarConfig  `mapstructure:"calendar"`
	Proxmox  []ProxmoxConfig `mapstructure:"proxmox"`
	Arcane   []ArcaneConfig  `mapstructure:"arcane"`
	Unifi    []UnifiConfig   `mapstructure:"unifi"`
	Services []ServiceConfig `mapstructure:"services"`
	Icons    IconsConfig     `mapstructure:"icons"`
}

// IconsConfig controla el proxy con caché de iconos de servicios. Los iconos se
// descargan del CDN bajo demanda y se guardan en disco para servirlos local en
// las siguientes cargas. Así no hay que versionar archivos: basta con poner el
// nombre del icono (slug de dashboardicons.com) en cada servicio.
type IconsConfig struct {
	CacheDir string `mapstructure:"cache_dir"` // dónde guardar los iconos cacheados
	CDNURL   string `mapstructure:"cdn_url"`   // plantilla con %s para el slug del icono
}

// ServerConfig controla cómo escucha el servidor HTTP.
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// BrandingConfig define la identidad visible del dashboard. Es la pieza central
// del whitelabeling: cambiando estos campos el dashboard se "convierte" en otro
// producto sin tocar código.
type BrandingConfig struct {
	Name        string `mapstructure:"name" json:"name"`                 // Nombre mostrado en header y <title>
	Tagline     string `mapstructure:"tagline" json:"tagline"`           // Subtítulo opcional
	LogoURL     string `mapstructure:"logo_url" json:"logo_url"`         // Ruta/URL del logo (vacío = inicial del nombre)
	FaviconURL  string `mapstructure:"favicon_url" json:"favicon_url"`   // Ruta/URL del favicon
	SearchURL   string `mapstructure:"search_url" json:"search_url"`     // Motor de búsqueda (default Google)
	SearchLabel string `mapstructure:"search_label" json:"search_label"` // Placeholder del input de búsqueda
}

// ThemeConfig define el sistema de diseño consumible por el frontend. El
// frontend lo recibe vía /api/branding e inyecta cada valor como CSS custom
// property en :root. Permite cambiar paleta y redondeces sin recompilar.
type ThemeConfig struct {
	Mode   string            `mapstructure:"mode" json:"mode"`     // dark | light (v1 solo dark)
	Colors map[string]string `mapstructure:"colors" json:"colors"` // token -> color hex/rgb
	Radius map[string]string `mapstructure:"radius" json:"radius"` // token -> valor css (ej "12px")
	Font   FontConfig        `mapstructure:"font" json:"font"`
}

// FontConfig permite sobreescribir las familias tipográficas.
type FontConfig struct {
	Sans string `mapstructure:"sans" json:"sans"`
	Mono string `mapstructure:"mono" json:"mono"`
}

// WeatherConfig configura el widget de clima (wttr.in).
type WeatherConfig struct {
	Location string `mapstructure:"location"`
	Units    string `mapstructure:"units"` // metric | imperial
}

// CalendarConfig configura el calendario estático del frontend.
type CalendarConfig struct {
	FirstDayOfWeek string `mapstructure:"first_day_of_week" json:"first_day_of_week"` // monday | sunday
}

// ProxmoxConfig define un nodo Proxmox a monitorear.
type ProxmoxConfig struct {
	Name        string `mapstructure:"name"`
	Host        string `mapstructure:"host"`
	TokenID     string `mapstructure:"token_id"`
	TokenSecret string `mapstructure:"token_secret"`
	VerifyTLS   bool   `mapstructure:"verify_tls"`
}

// ArcaneConfig define la conexión a la API de Arcane (gestión Docker).
type ArcaneConfig struct {
	Name   string `mapstructure:"name"`
	Host   string `mapstructure:"host"`
	APIKey string `mapstructure:"api_key"`
}

// UnifiConfig define una consola UniFi OS a monitorear vía la Integration API.
type UnifiConfig struct {
	Name      string `mapstructure:"name"`
	Host      string `mapstructure:"host"`
	APIKey    string `mapstructure:"api_key"`
	SiteID    string `mapstructure:"site_id"`
	VerifyTLS bool   `mapstructure:"verify_tls"`
}

// ServiceConfig define un acceso directo del app grid.
type ServiceConfig struct {
	Name     string `mapstructure:"name"`
	URL      string `mapstructure:"url"`
	Icon     string `mapstructure:"icon"`
	Category string `mapstructure:"category"`
}

// envRefPattern matchea referencias `${VAR}` para resolver secretos.
var envRefPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// Load lee, interpola y valida la configuración desde `path`.
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("no se pudo leer la configuración en %q: %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("configuración inválida: %w", err)
	}

	resolveEnvRefs(&cfg)
	applyThemeDefaults(&cfg)

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// setDefaults establece valores razonables para que el dashboard arranque
// incluso con un YAML mínimo.
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 3000)
	v.SetDefault("branding.name", "Polaris")
	v.SetDefault("branding.search_url", "https://www.google.com/search?q=")
	v.SetDefault("branding.search_label", "Buscar en la web…")
	v.SetDefault("theme.mode", "dark")
	v.SetDefault("weather.units", "metric")
	v.SetDefault("calendar.first_day_of_week", "monday")
	v.SetDefault("icons.cache_dir", "cache/icons")
	v.SetDefault("icons.cdn_url", "https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/svg/%s.svg")
}

// resolveEnvRefs reemplaza recursivamente las referencias ${VAR} en los campos
// sensibles por su valor en el entorno.
func resolveEnvRefs(cfg *Config) {
	for i := range cfg.Arcane {
		cfg.Arcane[i].APIKey = expandEnv(cfg.Arcane[i].APIKey)
	}
	for i := range cfg.Proxmox {
		cfg.Proxmox[i].TokenSecret = expandEnv(cfg.Proxmox[i].TokenSecret)
		cfg.Proxmox[i].TokenID = expandEnv(cfg.Proxmox[i].TokenID)
	}
	for i := range cfg.Unifi {
		cfg.Unifi[i].APIKey = expandEnv(cfg.Unifi[i].APIKey)
	}
}

// expandEnv resuelve todas las referencias ${VAR} dentro de s.
func expandEnv(s string) string {
	return envRefPattern.ReplaceAllStringFunc(s, func(match string) string {
		name := envRefPattern.FindStringSubmatch(match)[1]
		return os.Getenv(name)
	})
}

// defaultColors es la paleta base usada cuando el YAML no define un token.
// Garantiza que el frontend siempre reciba un set completo.
var defaultColors = map[string]string{
	"bg-base":        "#0F172A",
	"bg-surface":     "#1E293B",
	"bg-elevated":    "#263449",
	"border":         "#334155",
	"border-subtle":  "#263449",
	"text-primary":   "#E2E8F0",
	"text-secondary": "#94A3B8",
	"text-muted":     "#64748B",
	"accent":         "#38BDF8",
	"accent-glow":    "#004B50",
	"success":        "#A3E635",
	"warning":        "#F59E0B",
	"danger":         "#EF4444",
	"info":           "#7DD3FC",
}

var defaultRadius = map[string]string{
	"sm": "8px",
	"md": "12px",
	"lg": "16px",
	"xl": "22px",
}

// applyThemeDefaults completa los tokens de tema faltantes con los defaults,
// de modo que un YAML que solo redefine "accent" siga teniendo paleta completa.
func applyThemeDefaults(cfg *Config) {
	if cfg.Theme.Colors == nil {
		cfg.Theme.Colors = map[string]string{}
	}
	for k, def := range defaultColors {
		if _, ok := cfg.Theme.Colors[k]; !ok {
			cfg.Theme.Colors[k] = def
		}
	}
	if cfg.Theme.Radius == nil {
		cfg.Theme.Radius = map[string]string{}
	}
	for k, def := range defaultRadius {
		if _, ok := cfg.Theme.Radius[k]; !ok {
			cfg.Theme.Radius[k] = def
		}
	}
	if cfg.Theme.Font.Sans == "" {
		cfg.Theme.Font.Sans = "Inter, Geist, system-ui, sans-serif"
	}
	if cfg.Theme.Font.Mono == "" {
		cfg.Theme.Font.Mono = "'Geist Mono', ui-monospace, monospace"
	}
}

// validate verifica invariantes mínimas de la configuración.
func (c *Config) validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port inválido: %d", c.Server.Port)
	}
	if c.Branding.Name == "" {
		return fmt.Errorf("branding.name es obligatorio")
	}
	return nil
}

// Addr devuelve la dirección de escucha "host:port".
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
