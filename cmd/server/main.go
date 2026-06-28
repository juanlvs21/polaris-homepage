// Command server es el punto de entrada del dashboard.
//
// Carga la configuración, construye el servidor de Fiber (API + frontend
// embebido) y escucha. La ruta del config se puede sobreescribir con la
// variable de entorno CONFIG_PATH o el flag -config.
package main

import (
	"flag"
	"log"
	"os"

	"polaris-dashboard/internal/config"
	"polaris-dashboard/internal/server"
)

func main() {
	defaultPath := os.Getenv("CONFIG_PATH")
	if defaultPath == "" {
		defaultPath = "config.yaml"
	}
	configPath := flag.String("config", defaultPath, "ruta al archivo config.yaml")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("error de configuración: %v", err)
	}

	app := server.New(cfg)

	log.Printf("%s escuchando en http://%s", cfg.Branding.Name, cfg.Addr())
	if err := app.Listen(cfg.Addr()); err != nil {
		log.Fatalf("el servidor terminó: %v", err)
	}
}
