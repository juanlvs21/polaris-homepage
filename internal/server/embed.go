package server

import "embed"

// distFS contiene el build del frontend (frontend/dist) embebido en el binario.
//
// El directorio se incluye en compilación con `//go:embed`. Para que `go build`
// no falle cuando aún no existe el build, el repo versiona un `dist/.gitkeep`
// y un `index.html` de placeholder. Ejecuta `make build` (o `vite build`) para
// generar el frontend real antes de compilar el binario de producción.
//
//go:embed all:dist
var distFS embed.FS
