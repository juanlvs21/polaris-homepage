# Polaris Dashboard — Product Requirements Document

> Homepage personal para la red interna `webapp.casa`. Panel de control unificado para servicios self-hosted, métricas de infraestructura, clima local y utilidades del día a día. Construido en Go + Fiber con frontend Vite + Alpine.js + Tailwind CSS.

---

## Tabla de contenido

1. [Propósito y contexto](#1-propósito-y-contexto)
2. [Infraestructura objetivo](#2-infraestructura-objetivo)
3. [Stack tecnológico](#3-stack-tecnológico)
4. [Línea de diseño](#4-línea-de-diseño)
5. [Arquitectura del sistema](#5-arquitectura-del-sistema)
6. [Módulos y features](#6-módulos-y-features)
7. [Integraciones externas](#7-integraciones-externas)
8. [Configuración](#9-configuración)
9. [Tareas trazables](#10-tareas-trazables)

---

## 1. Propósito y contexto

Polaris Dashboard es la homepage que se sirve desde la raíz del dominio interno `webapp.casa`. Su función es actuar como punto de entrada único a toda la infraestructura self-hosted: accesos directos a aplicaciones, estado de la infraestructura en tiempo real (nodos Proxmox, VMs, LXC, contenedores Docker), clima local y utilidades básicas del día a día (búsqueda, calendario).

**No es** un panel de administración. No permite acciones destructivas sobre la infraestructura. Es de solo lectura en todo lo que no sea la propia configuración del dashboard.

**Audiencia:** uso personal / familiar en red interna. Sin autenticación en v1.

---

## 2. Infraestructura objetivo

| Componente | Descripción |
|---|---|
| **Proxmox Node 1** | Servidor físico principal. Corre VMs y contenedores LXC. |
| **Proxmox Node 2** | Segundo servidor físico. Corre VMs y contenedores LXC. |
| **VM Docker** | Máquina virtual dedicada a contenedores Docker, gestionada con Arcane. |
| **Dominio interno** | `webapp.casa` — resuelto por DNS interno (Pi-hole u otro resolver). |
| **Dashboard** | Corre como contenedor Docker en la VM de Docker. Accesible en `webapp.casa`. |

---

## 3. Stack tecnológico

### Backend

| Tecnología | Rol | Justificación |
|---|---|---|
| **Go 1.22+** | Lenguaje principal | Binario único, bajo consumo de recursos, concurrencia nativa |
| **Fiber v2** | HTTP framework | API similar a Express/Hono, familiar para desarrolladores JS |
| **Viper** | Gestión de configuración | Lectura de `config.yaml`, soporte de env vars |
| **go-resty** | Cliente HTTP | Llamadas a APIs externas (Proxmox, Arcane, wttr.in) |
| **`//go:embed`** | Embedding de frontend | Incluye el `dist/` del frontend dentro del binario de Go |
| **`errgroup`** | Concurrencia | Llamadas paralelas a Proxmox API para múltiples nodos/VMs |

### Frontend

| Tecnología | Rol | Justificación |
|---|---|---|
| **Vite** | Bundler / dev server | Build rápido, HMR, output optimizado para embedding en Go |
| **Alpine.js v3** | Reactividad UI | Ligero, sin build complejo, apropiado para dashboards de datos |
| **Tailwind CSS v3** | Estilos | Purge automático en build, responsive utilities, DX familiar |
| **Lucide Icons** | Iconografía del sistema | SVG inline, tree-shakeable, consistente |
| **selfh.st/icons** | Íconos de aplicaciones | Repositorio de íconos para apps self-hosted |

### Infraestructura / Deploy

| Tecnología | Rol |
|---|---|
| **Docker** | Contenedor de producción |
| **Dockerfile multistage** | Stage 1: build del frontend (Node). Stage 2: build de Go. Stage 3: imagen final mínima |
| **Makefile** | Comandos de desarrollo y build |

---

## 4. Línea de diseño

### Filosofía

Dashboard de infraestructura personal. El look es **dark, técnico y limpio** — como una terminal moderna, no como una app corporativa. Prioridad absoluta a la legibilidad de datos y la densidad de información sin sentirse apretado. Responsive completo: usable en móvil con una mano.

### Paleta de colores

```
--color-bg-base:      #0F172A   /* Fondo base del dashboard */
--color-bg-surface:   #1E293B   /* Cards, panels */
--color-bg-elevated:  #263449   /* Hover states, inputs */
--color-border:       #334155   /* Bordes de cards y divisores */
--color-border-subtle:#263449   /* Bordes secundarios */

--color-text-primary: #E2E8F0   /* Texto vivo */
--color-text-secondary:#94A3B8  /* Labels, metadata */
--color-text-muted:   #64748B   /* Texto deshabilitado */

--color-accent:       #38BDF8   /* Luz guía / primary */
--color-accent-glow:  #004B50   /* Variante teal para glows */

--color-success:      #A3E635   /* Estado saludable */
--color-warning:      #F59E0B   /* Amarillo — uso elevado de recursos */
--color-danger:       #EF4444   /* Rojo — servicios offline, uso crítico */
--color-info:         #7DD3FC   /* Azul claro — estados informativos */
```

**Nota de diseño:** la paleta usa un lienzo azul-negro profundo con paneles fríos, una luz guía cyan para interacción y verde esmeralda para salud de infraestructura. Se usa con contención para que los datos sigan siendo lo primero.

### Tipografía

```
Display / títulos:   "Inter" con fallback Geist — limpio, moderno, legible
Body / datos:        "Inter" con fallback Geist — consistencia de producto SaaS
Monospace / métricas: "Geist Mono" — porcentajes, IPs, versiones, logs
```

Inter/Geist pueden cargarse vía `fontsource` o CDN. Peso en display: 600–700. Peso en body: 400. Datos numéricos en Geist Mono weight 500.

Escala tipográfica:
```
text-xs:   11px — labels de métricas, metadata
text-sm:   13px — body de cards, descripciones
text-base: 15px — texto principal
text-lg:   18px — títulos de sección
text-xl:   22px — valores de métricas destacadas
text-2xl:  28px — números grandes (CPU %, temperatura)
```

### Bordes y radios

```
radius-sm:  8px   — badges y controles pequeños
radius-md:  12px  — cards internas, inputs
radius-lg:  16px  — cards principales
radius-xl:  22px  — modales, drawers
```

Los elementos de tipo "terminal" (bloques de logs, tablas de métricas densas) usan superficies internas compactas para mantener lectura técnica.

### Sombras y elevación

El sistema de elevación usa borde, highlight interno y sombras sutiles para acercarse a una apariencia tipo HeroUI sin perder densidad de dashboard. Las cards tienen borde `--color-border` y en hover el borde sube a `--color-accent` con transición de 150ms.

```css
/* Card en reposo */
border: 1px solid var(--color-border);
box-shadow: 0 1px 2px rgba(0,0,0,.28), 0 12px 30px rgba(0,0,0,.14);

/* Card en hover / focus */
border: 1px solid var(--color-accent);
transform: translateY(-1px);
transition: border-color 150ms ease, box-shadow 150ms ease, transform 150ms ease;
```

### Componentes de UI

**Cards de servicio**
- Fondo: `--color-bg-surface`
- Border radius: `radius-lg`
- Padding: `16px`
- Icono: 32×32px, centrado
- Nombre del servicio: `text-sm`, `text-primary`, `font-medium`
- Indicador de estado: dot de 8px en esquina superior derecha (`success` / `danger`)
- En hover: borde cambia a accent, cursor pointer, leve lift con `translateY(-1px)`

**Cards de métricas (Proxmox / Docker)**
- Header: nombre del nodo/VM en `text-xs` `text-secondary` uppercase + letter-spacing
- Valor principal: `text-2xl` Geist Mono, color según umbral
- Barra de progreso: `height: 4px`, border-radius pill, fondo `--color-border`, fill con degradado `accent → warning → danger` según porcentaje
- Footer: uptime / IP en `text-xs` `text-muted`

**Barra de búsqueda**
- Ocupa el ancho completo del header (max-width contenida)
- Fondo: `--color-bg-elevated`
- Border: `--color-border`
- Focus: border cambia a `--color-accent` + glow sutil `box-shadow: 0 0 0 3px rgba(88,166,255,0.15)`
- Ícono de lupa a la izquierda en `--color-text-secondary`
- Submit on Enter — redirige a `google.com/search?q=`

**Indicadores de estado (dot)**
```
● online   → #A3E635 con pulse animation sutil (opacity 1→0.5→1, 2s infinite)
● offline  → #f85149 estático
● unknown  → #484f58 estático
```

### Layout y grid

```
Mobile  (< 640px):  1 columna
Tablet  (640–1024px): 2 columnas
Desktop (> 1024px): 12 columnas grid, widgets con colspan variable
```

Estructura de la página (de arriba a abajo):

```
┌─────────────────────────────────────────┐
│  HEADER: Logo · Búsqueda Google · Clima  │  h: 64px fijo
├─────────────────────────────────────────┤
│                                         │
│  SECCIÓN: Accesos rápidos (app grid)    │  colspan 12
│                                         │
├──────────────────┬──────────────────────┤
│                  │                      │
│  Proxmox Node 1  │   Proxmox Node 2     │  colspan 6 c/u
│  (métricas nodo) │   (métricas nodo)    │
│                  │                      │
│  VMs / LXC list  │   VMs / LXC list     │
│                  │                      │
├──────────────────┴──────────────────────┤
│                                         │
│  Docker (Arcane) — contenedores grid    │  colspan 12
│                                         │
├────────────────────┬────────────────────┤
│                    │                    │
│  Calendario        │  Widget Clima      │  colspan 6 c/u
│  (estático)        │  (detalle)         │
│                    │                    │
└────────────────────┴────────────────────┘
```

En mobile todo se apila en una columna. El header colapsa la búsqueda debajo del logo. El clima se mueve al header como dato compacto (temperatura + ícono).

### Animaciones

Solo donde agregan valor real:
- **Pulse** en dots de servicios online (sutil, 2s infinite)
- **Fade-in + slide-up** al cargar cada sección (`opacity 0→1`, `translateY 8px→0`, staggered 50ms por sección)
- **Transición de borde** en hover de cards (150ms)
- **Skeleton loading** mientras se cargan las métricas (barras grises animadas con shimmer)
- Nada más. Sin rotaciones, sin bounces, sin particles.

---

## 5. Arquitectura del sistema

```
Browser
  │
  │  HTTP (static assets)
  ▼
┌─────────────────────────────────────────┐
│           Go + Fiber Server             │
│                                         │
│  GET /          → index.html (embed)    │
│  GET /static/*  → assets (embed)        │
│                                         │
│  GET /api/weather      → wttr.in proxy  │
│  GET /api/proxmox      → Proxmox API    │
│  GET /api/docker       → Arcane API     │
│  GET /api/config       → servicios cfg  │
└─────────────────────────────────────────┘
         │              │              │
         ▼              ▼              ▼
    wttr.in API    Proxmox API    Arcane API
    (internet)     (red interna)  (red interna)
```

### Flujo de datos en el frontend

El frontend es una SPA mínima. Alpine.js hace `fetch` a los endpoints `/api/*` al cargar la página y cada N segundos (polling). No hay WebSockets en v1.

Intervalos de refresco:
- Clima: cada 15 minutos
- Proxmox métricas: cada 30 segundos
- Docker contenedores: cada 30 segundos
- Estado de servicios (ping): cada 60 segundos

### Estructura de carpetas del proyecto

```
polaris-dashboard/
├── main.go                    # Entry point, setup de Fiber
├── config.yaml                # Configuración del usuario
├── Makefile                   # Comandos dev, build, docker
├── Dockerfile                 # Multistage build
├── .env.example               # Variables de entorno de ejemplo
│
├── internal/
│   ├── config/
│   │   └── config.go          # Structs de config + carga con Viper
│   ├── handlers/
│   │   ├── weather.go         # GET /api/weather
│   │   ├── proxmox.go         # GET /api/proxmox
│   │   ├── docker.go          # GET /api/docker
│   │   └── services.go        # GET /api/config (servicios del yaml)
│   ├── proxmox/
│   │   └── client.go          # Cliente Proxmox REST API
│   ├── arcane/
│   │   └── client.go          # Cliente Arcane REST API
│   └── weather/
│       └── client.go          # Cliente wttr.in
│
└── frontend/
    ├── package.json
    ├── vite.config.js
    ├── tailwind.config.js
    ├── postcss.config.js
    ├── index.html
    └── src/
        ├── main.js            # Entry point Alpine.js
        ├── style.css          # Tailwind directives + custom CSS vars
        ├── components/
        │   ├── search.js      # Barra de búsqueda
        │   ├── weather.js     # Widget de clima
        │   ├── calendar.js    # Calendario estático
        │   ├── services.js    # Grid de accesos directos
        │   ├── proxmox.js     # Cards de nodos Proxmox
        │   └── docker.js      # Grid de contenedores Docker
        └── utils/
            ├── fetch.js       # Wrapper de fetch con error handling
            └── format.js      # Formateo de bytes, uptime, etc.
```

---

## 6. Módulos y features

### 6.1 Header

**Descripción:** Barra superior fija. Visible en todas las vistas. Contiene el logo/nombre del dashboard, la barra de búsqueda y el clima compacto.

**Comportamiento:**
- Posición `sticky top-0`, z-index alto, fondo con `backdrop-blur` sutil y borde inferior
- En mobile: logo + temperatura/ícono de clima en la misma fila; búsqueda en fila debajo
- En desktop: logo a la izquierda, búsqueda centrada (max-width 480px), clima a la derecha

---

### 6.2 Búsqueda Google

**Descripción:** Campo de texto que redirige a Google con la query del usuario.

**Comportamiento:**
- Al hacer Enter o click en el ícono de lupa: `window.open('https://www.google.com/search?q=' + encodeURIComponent(query), '_blank')`
- Placeholder: `"Buscar en Google…"`
- Focus automático al cargar la página en desktop (no en mobile para no abrir teclado)
- Limpia el campo después de hacer submit

**No tiene:** historial de búsquedas, autocompletado, integración con motores alternativos (v1).

---

### 6.3 Widget de clima

**Descripción:** Muestra el clima actual de la ubicación configurada. Datos provenientes de `wttr.in`.

**Vista compacta (header, mobile y desktop):**
- Ícono del clima (emoji o SVG del set de wttr.in)
- Temperatura actual en °C
- Descripción breve ("Parcialmente nublado")

**Vista expandida (card en la sección inferior):**
- Temperatura actual (grande, Geist Mono)
- Sensación térmica
- Humedad
- Viento (km/h + dirección)
- Pronóstico de 3 días: ícono + temp máx/mín

**API:** `wttr.in/{ciudad}?format=j1` — devuelve JSON sin necesidad de API key.

**Config requerida en `config.yaml`:**
```yaml
weather:
  location: "Ciudad, País"
  units: metric  # metric | imperial
```

---

### 6.4 Calendario estático

**Descripción:** Vista de calendario del mes actual. Sin integración con calendarios externos. Solo orientación de fechas.

**Comportamiento:**
- Renderizado 100% en el frontend con Alpine.js, sin llamadas al backend
- Muestra el mes actual al cargar
- Botones `<` y `>` para navegar entre meses
- El día de hoy resaltado con el color accent
- Grid clásico: columnas Sun–Sat (o Lun–Dom, configurable)
- Semana en la que está el día de hoy resaltada sutilmente

**No tiene:** eventos, integración con Google Calendar, recordatorios, notificaciones.

---

### 6.5 Accesos directos (App Grid)

**Descripción:** Grid de cards con accesos directos a las aplicaciones self-hosted y servicios de la red.

**Estructura de cada card:**
- Ícono de la app (PNG/SVG de selfh.st/icons, servido desde `/static/icons/`)
- Nombre de la app
- URL de destino (abre en nueva pestaña)
- Dot de estado (online/offline basado en ping HTTP)
- Tag de categoría opcional (visible en desktop, oculto en mobile)

**Comportamiento:**
- Las apps se definen en `config.yaml`
- El backend hace un ping HTTP GET a la URL de cada servicio y devuelve el estado en `/api/config`
- El ping se hace en paralelo con goroutines (una por servicio)
- Timeout de ping: 3 segundos
- Las cards se ordenan por categoría (configurable el orden en yaml)

**Config requerida:**
```yaml
services:
  - name: "Gitea"
    url: "http://gitea.webapp.casa"
    icon: "gitea.svg"
    category: "Dev"
  - name: "Jellyfin"
    url: "http://jellyfin.webapp.casa"
    icon: "jellyfin.svg"
    category: "Media"
```

---

### 6.6 Métricas Proxmox

**Descripción:** Cards de estado para cada nodo Proxmox. Muestra métricas del nodo físico y lista de VMs / contenedores LXC con sus recursos.

#### Card de nodo (por cada servidor Proxmox)

**Métricas del nodo:**
- Nombre del nodo
- Estado (online/offline)
- CPU: porcentaje de uso + barra de progreso
- RAM: usado / total en GB + barra de progreso
- Disco (storage root): usado / total en GB + barra de progreso
- Temperatura de CPU (si disponible vía endpoint de sensores)
- Uptime formateado ("14d 3h 22m")
- Versión de Proxmox

**Colores de barras según umbral:**
```
< 60%  → --color-accent  (azul)
60–80% → --color-warning (amarillo)
> 80%  → --color-danger  (rojo)
```

#### Lista de VMs y LXC

Por cada nodo, debajo de las métricas del nodo, una lista compacta de todas las VMs y contenedores LXC con:
- ID + Nombre
- Tipo (VM / LXC) — badge
- Estado (running / stopped) — dot de color
- CPU %
- RAM usado / asignado
- IP (si disponible vía Proxmox guest agent)

Las VMs/LXC stopped aparecen al final y en `opacity: 0.5`.

**Integración Proxmox:**

Autenticación via API Token (no usuario/contraseña):
```
Header: Authorization: PVEAPIToken=USER@REALM!TOKENID=TOKEN_SECRET
```

Endpoints consumidos:
```
GET /api2/json/nodes                                          → lista nodos
GET /api2/json/nodes/{node}/status                            → métricas nodo
GET /api2/json/nodes/{node}/qemu                              → lista VMs
GET /api2/json/nodes/{node}/qemu/{vmid}/status/current        → métricas VM
GET /api2/json/nodes/{node}/lxc                               → lista LXC
GET /api2/json/nodes/{node}/lxc/{vmid}/status/current         → métricas LXC
```

Las llamadas a los endpoints de estado de cada VM/LXC se hacen en paralelo con `errgroup`.

**Config requerida:**
```yaml
proxmox:
  - name: "Node 1"
    host: "https://192.168.1.10:8006"
    token_id: "dashboard@pve!polaris"
    token_secret: "${PROXMOX_NODE1_TOKEN}"
    verify_tls: false
  - name: "Node 2"
    host: "https://192.168.1.11:8006"
    token_id: "dashboard@pve!polaris"
    token_secret: "${PROXMOX_NODE2_TOKEN}"
    verify_tls: false
```

---

### 6.7 Docker / Arcane

**Descripción:** Grid de contenedores Docker obtenidos desde la API de Arcane.

**Por cada contenedor:**
- Nombre del contenedor
- Imagen (nombre + tag)
- Estado (running / exited / paused) — dot de color
- CPU % de uso
- RAM: usado / límite (o usado / total host si no hay límite)
- Puertos expuestos (badges compactos)
- Uptime / tiempo desde último inicio

**Comportamiento:**
- Los contenedores `running` aparecen primero
- Los `exited` aparecen abajo con `opacity: 0.5`
- Click en un contenedor abre un drawer/modal lateral con:
  - Detalle extendido del contenedor
  - Variables de entorno (opcional, toggle para mostrar/ocultar)
  - Últimas 50 líneas de logs del contenedor (con fuente monospace, scroll)
  - Botón para ir a Arcane directamente

**Integración Arcane:**

Arcane expone una REST API. La autenticación se hace con API key.

Endpoints consumidos:
```
GET /api/containers           → lista de contenedores con stats básicos
GET /api/containers/{id}/logs → logs del contenedor
```

**Config requerida:**
```yaml
arcane:
  - name: "Arcane Core"
    host: "http://arcane.webapp.casa"
    api_key: "${ARCANE_API_KEY}"
  - name: "Arcane Media"
    host: "http://arcane-media.webapp.casa"
    api_key: "${ARCANE_MEDIA_API_KEY}"
```

---

## 7. Integraciones externas

| Integración | Protocolo | Auth | Dirección |
|---|---|---|---|
| **wttr.in** | HTTP REST | Ninguna | Outbound (internet) |
| **Proxmox Node 1** | HTTPS REST | API Token header | Inbound (red interna) |
| **Proxmox Node 2** | HTTPS REST | API Token header | Inbound (red interna) |
| **Arcane** | HTTP REST | API Key header | Inbound (red interna) |

---

## 8. Configuración

El archivo `config.yaml` es la fuente de verdad de la configuración. Los secrets (tokens, API keys) se inyectan como variables de entorno y se referencian en el yaml con la sintaxis `${VAR_NAME}`.

```yaml
# config.yaml — ejemplo completo

server:
  port: 3000
  host: "0.0.0.0"

weather:
  location: "Caracas, Venezuela"
  units: metric

calendar:
  first_day_of_week: monday  # monday | sunday

proxmox:
  - name: "Polaris-1"
    host: "https://192.168.1.10:8006"
    token_id: "dashboard@pve!polaris"
    token_secret: "${PROXMOX_NODE1_TOKEN}"
    verify_tls: false
  - name: "Polaris-2"
    host: "https://192.168.1.11:8006"
    token_id: "dashboard@pve!polaris"
    token_secret: "${PROXMOX_NODE2_TOKEN}"
    verify_tls: false

arcane:
  - name: "Arcane Core"
    host: "http://arcane.webapp.casa"
    api_key: "${ARCANE_API_KEY}"
  - name: "Arcane Media"
    host: "http://arcane-media.webapp.casa"
    api_key: "${ARCANE_MEDIA_API_KEY}"

services:
  - name: "Gitea"
    url: "http://gitea.webapp.casa"
    icon: "gitea.svg"
    category: "Dev"
  - name: "Jellyfin"
    url: "http://jellyfin.webapp.casa"
    icon: "jellyfin.svg"
    category: "Media"
  - name: "Proxmox"
    url: "https://proxmox.webapp.casa:8006"
    icon: "proxmox.svg"
    category: "Infra"
  - name: "Arcane"
    url: "http://arcane.webapp.casa"
    icon: "docker.svg"
    category: "Infra"
```

---

## 9. Tareas trazables

Las tareas están organizadas por épica. Cada tarea tiene un identificador, descripción, criterios de aceptación y dependencias.

---

### ÉPICA 1 — Setup y scaffolding del proyecto

#### TASK-001 — Inicializar módulo Go y dependencias base
- **Descripción:** Crear el repositorio, inicializar `go mod`, instalar dependencias base (Fiber, Viper, go-resty, errgroup).
- **Criterios de aceptación:**
  - [ ] `go.mod` y `go.sum` creados correctamente
  - [ ] `main.go` con servidor Fiber levantando en `:3000`
  - [ ] `GET /health` devuelve `{"status": "ok"}`
- **Dependencias:** ninguna

#### TASK-002 — Inicializar proyecto frontend con Vite
- **Descripción:** Crear el proyecto frontend en `/frontend` con Vite, Alpine.js, Tailwind CSS y Lucide.
- **Criterios de aceptación:**
  - [ ] `npm run dev` levanta el dev server en `:5173`
  - [ ] `npm run build` genera `/frontend/dist/`
  - [ ] Tailwind procesa correctamente y purga clases no usadas en build
  - [ ] Alpine.js disponible globalmente
- **Dependencias:** ninguna

#### TASK-003 — Configurar embedding del frontend en Go
- **Descripción:** Usar `//go:embed` para embeber el `dist/` del frontend dentro del binario de Go. Fiber sirve los archivos estáticos desde ahí.
- **Criterios de aceptación:**
  - [ ] `go build` produce un único binario
  - [ ] El binario sirve `index.html` en `GET /`
  - [ ] El binario sirve assets en `GET /static/*`
  - [ ] Sin el directorio `/frontend/dist/` presente, el binario sigue funcionando
- **Dependencias:** TASK-001, TASK-002

#### TASK-004 — Implementar sistema de configuración con Viper
- **Descripción:** Cargar `config.yaml` al inicio con Viper. Soportar interpolación de variables de entorno.
- **Criterios de aceptación:**
  - [ ] La config se carga correctamente al arrancar el servidor
  - [ ] Los campos `${VAR_NAME}` se resuelven desde env vars
  - [ ] Si falta la config, el servidor loguea un error descriptivo y termina
  - [ ] Struct tipados en Go para cada sección del yaml
- **Dependencias:** TASK-001

#### TASK-005 — Configurar Makefile de desarrollo
- **Descripción:** Makefile con targets para las tareas comunes del proyecto.
- **Criterios de aceptación:**
  - [ ] `make dev` levanta el backend Go y el frontend Vite en paralelo
  - [ ] `make build` hace `vite build` y luego `go build`
  - [ ] `make docker` construye la imagen Docker
  - [ ] `make clean` elimina binario y `dist/`
- **Dependencias:** TASK-001, TASK-002

#### TASK-006 — Configurar proxy de Vite hacia el backend Go en desarrollo
- **Descripción:** En desarrollo, las llamadas a `/api/*` del frontend deben ser redirigidas por Vite al backend Go en `:3000`.
- **Criterios de aceptación:**
  - [ ] `vite.config.js` tiene configurado `proxy: { '/api': 'http://localhost:3000' }`
  - [ ] Las llamadas a `/api/*` desde el frontend llegan al backend sin errores de CORS
- **Dependencias:** TASK-001, TASK-002

#### TASK-007 — Dockerfile multistage
- **Descripción:** Dockerfile con tres stages: build del frontend (Node.js), build del binario Go, imagen final mínima (distroless o alpine).
- **Criterios de aceptación:**
  - [ ] `docker build` produce una imagen funcional
  - [ ] La imagen final pesa menos de 30MB
  - [ ] El contenedor arranca y sirve el dashboard correctamente
  - [ ] La config se inyecta via volumen (`-v ./config.yaml:/app/config.yaml`)
- **Dependencias:** TASK-003, TASK-005

---

### ÉPICA 2 — Diseño base y layout

#### TASK-008 — Implementar design tokens y estilos globales
- **Descripción:** Definir las CSS custom properties del sistema de diseño. Configurar Tailwind para usar los tokens como colores custom.
- **Criterios de aceptación:**
  - [ ] Todas las variables de color del sistema de diseño definidas en `:root`
  - [ ] `tailwind.config.js` extiende los colores con los tokens (`bg-surface`, `accent`, etc.)
  - [ ] Fuentes Geist y Geist Mono cargadas y aplicadas al `body`
  - [ ] Reset CSS base aplicado
- **Dependencias:** TASK-002

#### TASK-009 — Implementar layout principal responsive
- **Descripción:** Estructura HTML/CSS del layout completo: header fijo + grid de secciones. Sin contenido real aún, solo la estructura.
- **Criterios de aceptación:**
  - [ ] Header con `position: sticky` y `z-index` correcto
  - [ ] Grid de 12 columnas en desktop, 2 en tablet, 1 en mobile
  - [ ] Layout funciona sin scroll horizontal en todos los breakpoints
  - [ ] Testado en viewport de 375px (iPhone SE) y 1440px (desktop estándar)
- **Dependencias:** TASK-008

#### TASK-010 — Implementar componente Card base
- **Descripción:** Componente reutilizable de card con todos sus estados (default, hover, loading skeleton).
- **Criterios de aceptación:**
  - [ ] Estilos de card en reposo, hover y focus correctos según el sistema de diseño
  - [ ] Skeleton loading con animación shimmer
  - [ ] Card se adapta al contenido (no altura fija)
- **Dependencias:** TASK-008

---

### ÉPICA 3 — Búsqueda Google

#### TASK-011 — Implementar barra de búsqueda
- **Descripción:** Componente de búsqueda en el header que redirige a Google.
- **Criterios de aceptación:**
  - [ ] Enter en el input abre `https://www.google.com/search?q=<query>` en nueva pestaña
  - [ ] El campo se limpia después del submit
  - [ ] En desktop, el input recibe focus automático al cargar la página
  - [ ] Ícono de lupa visible a la izquierda
  - [ ] Focus ring con glow de color accent
  - [ ] Responsive: ocupa el ancho disponible en mobile
- **Dependencias:** TASK-009

---

### ÉPICA 4 — Widget de clima

#### TASK-012 — Implementar cliente wttr.in en Go
- **Descripción:** Función en `internal/weather/client.go` que consulta `wttr.in` y devuelve una struct tipada con los datos del clima.
- **Criterios de aceptación:**
  - [ ] `GET wttr.in/{location}?format=j1` parseado correctamente
  - [ ] Struct de respuesta con: temperatura actual, sensación térmica, humedad, viento, descripción, código de condición, pronóstico de 3 días
  - [ ] Error handling apropiado (timeout, respuesta inválida)
  - [ ] Tests unitarios con respuesta mockeada
- **Dependencias:** TASK-004

#### TASK-013 — Implementar endpoint `GET /api/weather`
- **Descripción:** Handler de Fiber que llama al cliente wttr.in y devuelve los datos como JSON.
- **Criterios de aceptación:**
  - [ ] Devuelve JSON con los datos del clima
  - [ ] Cache en memoria de 15 minutos (no llamar a wttr.in en cada request del frontend)
  - [ ] Si wttr.in falla, devuelve el último dato cacheado con un campo `stale: true`
  - [ ] Responde en menos de 2 segundos
- **Dependencias:** TASK-012

#### TASK-014 — Implementar widget de clima en el frontend
- **Descripción:** Componente Alpine.js que consulta `/api/weather` y renderiza el clima.
- **Criterios de aceptación:**
  - [ ] Vista compacta en el header: ícono + temperatura + descripción breve
  - [ ] Card expandida en la sección inferior con todos los datos
  - [ ] Pronóstico de 3 días con ícono + temp máx/mín
  - [ ] Se refresca automáticamente cada 15 minutos
  - [ ] Muestra skeleton mientras carga
  - [ ] Muestra estado de error si la API falla
- **Dependencias:** TASK-009, TASK-013

---

### ÉPICA 5 — Calendario estático

#### TASK-015 — Implementar componente de calendario
- **Descripción:** Componente Alpine.js de calendario mensual, 100% frontend, sin llamadas al backend.
- **Criterios de aceptación:**
  - [ ] Muestra el mes y año actual al cargar
  - [ ] Navegación entre meses con botones `<` y `>`
  - [ ] Primer día de la semana configurable (lunes o domingo)
  - [ ] Día de hoy resaltado con el color accent
  - [ ] Días de meses adyacentes visibles pero en `text-muted`
  - [ ] Responsive: se ve bien en mobile con los 7 días en una fila
- **Dependencias:** TASK-009

---

### ÉPICA 6 — Accesos directos (App Grid)

#### TASK-016 — Implementar endpoint `GET /api/config`
- **Descripción:** Handler que lee los servicios del `config.yaml` y devuelve la lista con el estado de cada servicio (ping HTTP).
- **Criterios de aceptación:**
  - [ ] Devuelve JSON con la lista de servicios definidos en config
  - [ ] El ping a cada servicio se hace en paralelo con goroutines
  - [ ] Timeout de ping: 3 segundos por servicio
  - [ ] Estado: `online` (2xx/3xx), `offline` (timeout o error de conexión), `unknown` (respuesta inesperada)
  - [ ] Si un ping falla, no bloquea a los demás
  - [ ] Los íconos se sirven desde `/static/icons/`
- **Dependencias:** TASK-004

#### TASK-017 — Implementar grid de accesos directos
- **Descripción:** Grid de cards de servicios en el frontend.
- **Criterios de aceptación:**
  - [ ] Grid adaptable: 2 col mobile, 4 col tablet, 6+ col desktop
  - [ ] Cada card tiene: ícono, nombre, dot de estado
  - [ ] Click abre la URL en nueva pestaña
  - [ ] Dot de estado con animación pulse cuando está online
  - [ ] Agrupación por categoría con separador de sección
  - [ ] Se refresca el estado cada 60 segundos
  - [ ] Skeleton loading mientras carga
- **Dependencias:** TASK-009, TASK-016

#### TASK-018 — Descargar y organizar íconos de servicios
- **Descripción:** Descargar los SVGs/PNGs de los servicios configurados desde selfh.st/icons y colocarlos en `/frontend/public/icons/`.
- **Criterios de aceptación:**
  - [ ] Íconos disponibles para todos los servicios en `config.yaml`
  - [ ] Íconos optimizados (SVG < 10KB por ícono)
  - [ ] Naming convention consistente: `servicio.svg` en minúsculas
- **Dependencias:** TASK-002

---

### ÉPICA 7 — Métricas Proxmox

#### TASK-019 — Implementar cliente Proxmox API en Go
- **Descripción:** Cliente HTTP en `internal/proxmox/client.go` que se autentica con API Token y expone métodos para cada endpoint necesario.
- **Criterios de aceptación:**
  - [ ] Autenticación via header `Authorization: PVEAPIToken=...`
  - [ ] Soporte para `verify_tls: false` (self-signed certs de Proxmox)
  - [ ] Métodos: `GetNodeStatus()`, `GetVMs()`, `GetVMStatus()`, `GetLXCs()`, `GetLXCStatus()`
  - [ ] Todos los tipos de respuesta con structs tipados
  - [ ] Error handling con mensajes descriptivos
  - [ ] Tests unitarios con mocks
- **Dependencias:** TASK-004

#### TASK-020 — Implementar fetch paralelo de métricas multi-nodo
- **Descripción:** Lógica que usa `errgroup` para hacer todas las llamadas a la API de Proxmox en paralelo (múltiples nodos, múltiples VMs/LXC por nodo).
- **Criterios de aceptación:**
  - [ ] Las llamadas a ambos nodos se hacen en paralelo
  - [ ] Dentro de cada nodo, las llamadas de estado de cada VM/LXC se hacen en paralelo
  - [ ] Si un nodo falla, el otro sigue funcionando y se devuelve su data
  - [ ] El tiempo total de respuesta es aprox. el de la llamada más lenta, no la suma
  - [ ] Timeout global de 10 segundos para todo el batch
- **Dependencias:** TASK-019

#### TASK-021 — Implementar endpoint `GET /api/proxmox`
- **Descripción:** Handler de Fiber que orquesta las llamadas paralelas y devuelve el estado completo de la infraestructura Proxmox.
- **Criterios de aceptación:**
  - [ ] Devuelve JSON con: lista de nodos, métricas de cada nodo, lista de VMs y LXC con métricas
  - [ ] Cache en memoria de 30 segundos
  - [ ] Si Proxmox no responde, devuelve el último dato cacheado con `stale: true`
  - [ ] La respuesta incluye el timestamp de la última actualización
- **Dependencias:** TASK-020

#### TASK-022 — Implementar cards de nodos Proxmox en el frontend
- **Descripción:** Componente Alpine.js que renderiza las cards de nodos y la lista de VMs/LXC.
- **Criterios de aceptación:**
  - [ ] Card por nodo con: nombre, estado, CPU%, RAM, disco, uptime
  - [ ] Barras de progreso con colores según umbral (azul/amarillo/rojo)
  - [ ] Lista de VMs/LXC debajo de cada nodo con estado, CPU%, RAM, IP
  - [ ] VMs/LXC detenidas aparecen al final con opacity reducida
  - [ ] Badge de tipo VM vs LXC visible
  - [ ] Se refresca cada 30 segundos
  - [ ] Timestamp "Actualizado hace X seg" visible
  - [ ] Skeleton loading mientras carga
- **Dependencias:** TASK-009, TASK-021

---

### ÉPICA 8 — Docker / Arcane

#### TASK-023 — Implementar cliente Arcane API en Go
- **Descripción:** Cliente HTTP en `internal/arcane/client.go` que consume la API de Arcane.
- **Criterios de aceptación:**
  - [ ] Autenticación con API key via header
  - [ ] Método `GetContainers()` que devuelve lista con stats
  - [ ] Método `GetContainerLogs(id string, tail int)` que devuelve logs
  - [ ] Structs tipados para todas las respuestas
  - [ ] Error handling apropiado
- **Dependencias:** TASK-004

#### TASK-024 — Implementar endpoint `GET /api/docker`
- **Descripción:** Handler que consulta Arcane y devuelve el estado de los contenedores.
- **Criterios de aceptación:**
  - [ ] Devuelve JSON con lista de contenedores: nombre, imagen, estado, CPU%, RAM, puertos, uptime
  - [ ] Cache de 30 segundos
  - [ ] Si Arcane no responde, devuelve último dato cacheado con `stale: true`
- **Dependencias:** TASK-023

#### TASK-025 — Implementar endpoint `GET /api/docker/:id/logs`
- **Descripción:** Handler que devuelve los logs de un contenedor específico.
- **Criterios de aceptación:**
  - [ ] Acepta parámetro `?tail=50` (default 50, máx 200)
  - [ ] Devuelve los logs como array de strings
  - [ ] Sin cache (siempre fresh)
- **Dependencias:** TASK-023

#### TASK-026 — Implementar grid de contenedores Docker en el frontend
- **Descripción:** Sección que muestra los contenedores Docker obtenidos de Arcane.
- **Criterios de aceptación:**
  - [ ] Grid de cards: 1 col mobile, 2 col tablet, 3 col desktop
  - [ ] Cada card: nombre, imagen+tag, dot de estado, CPU%, RAM, puertos
  - [ ] Contenedores `running` primero, `exited` al final con opacity reducida
  - [ ] Se refresca cada 30 segundos
  - [ ] Skeleton loading mientras carga
- **Dependencias:** TASK-009, TASK-024

#### TASK-027 — Implementar drawer de detalle de contenedor
- **Descripción:** Panel lateral que se abre al hacer click en un contenedor con detalle extendido y logs.
- **Criterios de aceptación:**
  - [ ] Se abre desde el lateral derecho con animación slide-in (300ms ease)
  - [ ] Overlay oscuro en el contenido de atrás
  - [ ] Muestra todos los datos del contenedor
  - [ ] Sección de logs: últimas 50 líneas en fuente Geist Mono, fondo más oscuro, scroll vertical
  - [ ] Toggle "Mostrar variables de entorno" oculto por default
  - [ ] Botón "Abrir en Arcane" que lleva a la URL de Arcane
  - [ ] Cierra al hacer click en el overlay o presionar Escape
  - [ ] En mobile: ocupa el 100% del ancho de pantalla
- **Dependencias:** TASK-026, TASK-025

---

### ÉPICA 9 — Pulido y producción

#### TASK-028 — Implementar animaciones y micro-interacciones
- **Descripción:** Agregar las animaciones definidas en el sistema de diseño.
- **Criterios de aceptación:**
  - [ ] Fade-in + slide-up staggered en las secciones al cargar (50ms entre secciones)
  - [ ] Pulse animation en dots de servicios online
  - [ ] Transición de borde en hover de cards (150ms)
  - [ ] Animación shimmer en skeletons
  - [ ] `prefers-reduced-motion` respetado (todas las animaciones desactivadas)
- **Dependencias:** TASK-017, TASK-022, TASK-026

#### TASK-029 — Testing de responsividad completo
- **Descripción:** Verificar el layout en todos los breakpoints objetivo.
- **Criterios de aceptación:**
  - [ ] 375px (iPhone SE) — sin scroll horizontal, sin elementos cortados
  - [ ] 768px (tablet portrait)
  - [ ] 1024px (tablet landscape / laptop pequeño)
  - [ ] 1440px (desktop estándar)
  - [ ] Drawer de Docker funciona en mobile (100% width)
  - [ ] Barra de búsqueda usable con teclado virtual en mobile
- **Dependencias:** todas las épicas de UI

#### TASK-030 — Documentar `config.yaml` y variables de entorno
- **Descripción:** `README.md` con instrucciones de instalación, configuración y deploy.
- **Criterios de aceptación:**
  - [ ] Instrucciones de creación del API Token en Proxmox
  - [ ] Instrucciones de obtención del API Key en Arcane
  - [ ] `config.yaml` de ejemplo comentado
  - [ ] Ejemplo de `docker-compose.yml` con el dashboard
  - [ ] Variables de entorno documentadas en `.env.example`
- **Dependencias:** TASK-007

#### TASK-031 — docker-compose.yml de producción
- **Descripción:** Archivo `docker-compose.yml` listo para deploy en la VM de Docker.
- **Criterios de aceptación:**
  - [ ] Monta `config.yaml` como volumen
  - [ ] Inyecta secrets como env vars
  - [ ] Puerto configurable (default `80`)
  - [ ] `restart: unless-stopped`
  - [ ] Conectado a la red interna de Docker si aplica
- **Dependencias:** TASK-007

---

## Resumen de épicas y estimado de tareas

| Épica | Tareas | Complejidad |
|---|---|---|
| Setup y scaffolding | TASK-001 a TASK-007 | Media |
| Diseño base y layout | TASK-008 a TASK-010 | Media |
| Búsqueda Google | TASK-011 | Baja |
| Widget de clima | TASK-012 a TASK-014 | Media |
| Calendario estático | TASK-015 | Baja |
| Accesos directos | TASK-016 a TASK-018 | Media |
| Métricas Proxmox | TASK-019 a TASK-022 | Alta |
| Docker / Arcane | TASK-023 a TASK-027 | Alta |
| Pulido y producción | TASK-028 a TASK-031 | Media |

**Total: 31 tareas**

---

*Documento generado para el proyecto Polaris Dashboard — `webapp.casa`*
