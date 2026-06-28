# Polaris Dashboard

> Homepage / panel de control unificado para infraestructura self-hosted.
> Un único binario de Go sirve una SPA (Vite + Alpine.js + Tailwind) y expone
> una API de solo lectura hacia Proxmox, Docker (Arcane) y el clima (wttr.in).
>
> **100% whitelabel:** el nombre, el logo, la paleta de colores y los radios se
> definen en un archivo `config.yaml`. El mismo binario se convierte en
> cualquier marca sin recompilar. Ver [DESIGN.md](DESIGN.md).

---

## Tabla de contenido

- [Características](#características)
- [Arquitectura](#arquitectura)
- [Requisitos](#requisitos)
- [Configuración](#configuración)
- [Desarrollo local](#desarrollo-local)
- [Build de producción](#build-de-producción)
- [Deploy con Docker](#deploy-con-docker)
- [Whitelabeling](#whitelabeling)
- [Integraciones](#integraciones)
- [Comandos del Makefile](#comandos-del-makefile)

---

## Características

- **Accesos directos** a tus servicios con ping de estado (online/offline).
- **Métricas Proxmox**: CPU, RAM, disco, uptime y lista de VMs/LXC por nodo.
- **Contenedores Docker** vía Arcane, con drawer de detalle y logs.
- **Clima** actual + pronóstico de 3 días (wttr.in, sin API key).
- **Búsqueda** web y **calendario** mensual.
- **Dark mode técnico**, responsive, con tema configurable por YAML.
- **Un solo binario**: el frontend se embebe con `//go:embed`. Sin Node en producción.

---

## Arquitectura

Arquitectura limpia, por capas, fácil de escalar:

```
cmd/server/            Entry point (main): carga config y arranca el servidor
internal/
  config/              Carga + validación de config.yaml (Viper), env interpolation
  server/              Fiber app: middleware, rutas, frontend embebido (go:embed)
  handlers/            Capa HTTP: un handler por endpoint /api/*
  cache/               Cache en memoria genérico con TTL + fallback "stale"
  clients/
    proxmox/           Cliente REST de Proxmox (solo lectura)
    arcane/            Cliente REST de Arcane (Docker)
    weather/           Cliente de wttr.in
frontend/
  src/
    theme.js           Motor de whitelabeling (inyecta tokens + identidad)
    components/         Un componente Alpine por módulo (search, weather, …)
    utils/              fetch + formateo
```

**Principios:**
- Los *clients* no conocen HTTP de Fiber; los *handlers* no conocen el detalle
  de cada API externa. Agregar una integración nueva = un paquete en
  `internal/clients/` + un handler + un componente Alpine.
- Toda lectura externa pasa por `cache` para no saturar las APIs con el polling.
- Los secretos nunca tocan el YAML: se referencian como `${VAR}` y se resuelven
  desde el entorno.

```
Browser ── HTTP ──► Go + Fiber ──► wttr.in / Proxmox / Arcane
              (assets embebidos)   (cacheado, en paralelo con errgroup)
```

---

## Requisitos

| Herramienta | Versión | Para qué |
|---|---|---|
| Go | 1.22+ (probado en 1.26) | Backend |
| Node.js | 20.19+ / 22.12+ (Vite 8; probado en 24) | Build del frontend |
| Docker | opcional | Deploy de producción |

---

## Configuración

La fuente de verdad es **`config.yaml`**. Empieza copiando el ejemplo:

```bash
cp config.example.yaml config.yaml
cp .env.example .env        # secretos: tokens y API keys
```

- Edita `config.yaml` con tu branding, tema, servicios e integraciones.
- Pon los secretos en `.env` y referéncialos en el YAML con `${VAR}`.
- El binario busca `config.yaml` en el directorio actual, o donde indique
  `CONFIG_PATH` / el flag `-config`.

Ver [`config.example.yaml`](config.example.yaml) para el archivo completo comentado.

---

## Desarrollo local

Dos procesos: el backend de Go (`:3000`) y el dev server de Vite (`:5173`) con
HMR. Vite proxea `/api/*` al backend, así que no hay CORS.

```bash
make deps        # instala dependencias de Go y del frontend (una vez)
make dev         # levanta backend + frontend en paralelo
```

Abre **http://localhost:5173** (frontend con hot-reload).
El backend solo está en `:3000` (la API; sirve el frontend embebido en prod).

¿Prefieres correrlos por separado?

```bash
make dev-backend     # solo Go  (:3000)
make dev-frontend    # solo Vite (:5173)
```

---

## Build de producción

El frontend se compila **dentro** del árbol de Go (`internal/server/dist`) para
que `//go:embed` lo incluya. El resultado es un único binario autocontenido.

```bash
make build       # 1) vite build  2) go build -> ./polaris-dashboard
./polaris-dashboard
# -> sirve API + frontend en http://localhost:3000
```

El binario no necesita Node ni el directorio `frontend/`: todo va embebido.

---

## Deploy con Docker

Imagen multistage (Node → Go → distroless). La imagen final es mínima y no
contiene toolchain.

```bash
# Build de la imagen
make docker          # docker build -t polaris-dashboard:latest .

# Run rápido montando tu config
make docker-run
```

### docker-compose (recomendado para la VM de Docker)

```bash
cp config.example.yaml config.yaml   # ajusta
cp .env.example .env                 # rellena secretos
docker compose up -d
```

`docker-compose.yml` monta `config.yaml` como volumen de solo lectura, inyecta
los secretos como variables de entorno y expone el puerto `${DASHBOARD_PORT}`
(default `80`). `restart: unless-stopped`.

---

## Whitelabeling

Todo lo visible se controla desde `config.yaml`, sin tocar código ni recompilar.
El flujo: el backend expone `GET /api/branding`; al arrancar, el frontend lee la
respuesta y **inyecta los tokens como CSS custom properties** (`--color-*`,
`--radius-*`, `--font-*`) y fija el `<title>` y el favicon.

```yaml
branding:
  name: "Mi Panel"                  # header + <title>
  search_url: "https://duckduckgo.com/?q="
  favicon_url: "/static/icons/brand.svg"

theme:
  mode: dark
  colors:
    accent: "#38BDF8"               # solo redefines lo que quieras
    bg-base: "#0F172A"
  radius:
    sm: "8px"
    md: "12px"
    lg: "16px"
    xl: "22px"
  font:
    sans: "Inter, Geist, system-ui, sans-serif"
```

Los tokens que no definas heredan los valores por defecto de Polaris.
Detalle completo del sistema de diseño en [DESIGN.md](DESIGN.md).

---

## Integraciones

### Proxmox (API Token, solo lectura)

1. En Proxmox: **Datacenter → Permissions → API Tokens → Add**.
2. Crea un token para un usuario con rol **PVEAuditor** (solo lectura).
3. Pon el secreto en `.env` (`PROXMOX_NODE1_TOKEN=…`).
4. En `config.yaml`, bajo `proxmox:`, configura `host`, `token_id` y
   `token_secret: "${PROXMOX_NODE1_TOKEN}"`. Usa `verify_tls: false` para los
   certificados self-signed típicos de Proxmox.

### Arcane (Docker)

1. En Arcane: **Settings → API Keys**, genera una key.
2. `ARCANE_API_KEY=…` en `.env`.
3. En `config.yaml`, bajo `arcane:`, agrega una entrada por instancia con
   `name`, `host` y `api_key: "${ARCANE_API_KEY}"`.

> El shape exacto de la API de Arcane puede variar por versión; ajusta los tags
> `json` en `internal/clients/arcane/client.go` si tu instancia difiere.

### Clima (wttr.in)

Sin API key. Solo define `weather.location` y `weather.units` en el YAML.

### Íconos de servicios

Coloca los SVG/PNG en `frontend/public/icons/` (se sirven en
`/static/icons/`). Buen repositorio: [selfh.st/icons](https://selfh.st/icons).

---

## Comandos del Makefile

```bash
make help        # lista todos los targets
make deps        # instala dependencias
make dev         # backend + frontend en desarrollo
make build       # build de producción (binario único)
make docker      # construye la imagen Docker
make test        # tests de Go
make clean       # limpia binario y build del frontend
```

---

## Endpoints de la API

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/api/health` | Liveness |
| GET | `/api/branding` | Identidad + tema (whitelabeling) |
| GET | `/api/config` | Servicios + estado (ping) |
| GET | `/api/weather` | Clima actual + pronóstico (cache 15m) |
| GET | `/api/proxmox` | Métricas de nodos + VMs/LXC (cache 30s) |
| GET | `/api/docker` | Contenedores (cache 30s) |
| GET | `/api/docker/:id/logs?tail=50` | Logs de un contenedor |

Cuando una API externa falla, el endpoint devuelve el último dato cacheado con
`"stale": true` en lugar de un error.
