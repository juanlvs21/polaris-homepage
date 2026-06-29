# Sistema de diseño — Polaris Dashboard

Guía del lenguaje visual y, sobre todo, del **mecanismo de theming configurable**
que hace el dashboard whitelabel. Léelo junto a [config.example.yaml](config.example.yaml).

---

## 1. Filosofía

Dashboard de infraestructura. El look es **dark, limpio y pulido**, inspirado en
componentes tipo HeroUI: superficies suaves, controles redondeados, bordes
claros y elevación contenida. Prioridad a la legibilidad de datos y a la
densidad de información sin sentirse apretado. Responsive completo.

Principios:

1. **Elevación suave.** Las cards se separan del fondo con borde, highlight
   interno y sombra contenida; en hover el borde sube a `accent`.
2. **Pocos colores vivos.** `accent` guía interacción y branding; `success`
   comunica salud. El resto de la superficie se mantiene oscuro, sobrio y
   legible.
3. **El color comunica estado.** `success` / `warning` / `danger` codifican
   salud de servicios y umbrales de uso, nunca decoración.
4. **Configurable, no hardcodeado.** Ningún color ni radio se escribe en el
   markup. Todo pasa por *design tokens* (CSS custom properties) alimentados
   desde el `config.yaml`.

---

## 2. Cómo funciona el theming configurable

Esta es la pieza central del whitelabeling. El flujo completo:

```
config.yaml (theme:)                          ← fuente de verdad
   │  Viper + defaults (internal/config)
   ▼
GET /api/branding  ──►  { theme: { colors, radius, font, mode }, branding }
   │  fetch al arrancar (frontend/src/theme.js)
   ▼
document.documentElement.style.setProperty('--color-accent', '#…')
   │  CSS custom properties en :root
   ▼
Tailwind v4 (bg-accent → var(--color-accent))  ← @theme genera utilidades+vars
   │
   ▼
UI renderizada con la marca activa
```

Puntos clave (Tailwind v4, config CSS-first):

- **`internal/config/config.go`** define `defaultColors` y `defaultRadius`. Si el
  YAML omite un token, se rellena con el default → el frontend siempre recibe un
  set completo.
- **`frontend/src/style.css`** declara los tokens en el bloque `@theme`. En
  Tailwind v4 esto hace dos cosas a la vez: (1) genera las utilidades
  (`bg-accent`, `rounded-lg`, `text-2xl`…) y (2) expone cada token como CSS
  custom property en `:root`. Las utilidades generadas **referencian la
  variable** (`background-color: var(--color-accent)`), no el valor literal.
- **`frontend/src/theme.js`** (`applyTheme`) sobreescribe en runtime cada token
  (`--color-<nombre>`, `--radius-<nombre>`, `--font-<sans|mono>`) en el elemento
  raíz con el valor de `/api/branding`. Como las utilidades usan `var(...)`,
  TODA la UI cambia en vivo. También fija `document.title` y el favicon
  (`applyIdentity`).
- No hay `tailwind.config.js` ni `postcss.config.js`: Tailwind v4 se integra con
  el plugin oficial de Vite (`@tailwindcss/vite`, en `vite.config.js`).

> Para añadir un token nuevo: agrégalo a `defaultColors`/`defaultRadius` (Go) y
> al bloque `@theme` de `style.css`. Dos lugares, mismo nombre.

---

## 3. Design tokens

### Colores (paleta por defecto)

| Token | Valor | Uso |
|---|---|---|
| `bg-base` | `#131418` | Fondo base del dashboard |
| `bg-surface` | `#1E1F24` | Cards, paneles |
| `bg-elevated` | `#2B2D33` | Hover, inputs |
| `border` | `#3A3C44` | Bordes de cards y divisores |
| `border-subtle` | `#26282E` | Bordes secundarios |
| `text-primary` | `#F2F3F5` | Texto principal |
| `text-secondary` | `#B5BAC1` | Labels, metadata |
| `text-muted` | `#80848E` | Texto deshabilitado |
| `accent` | `#5865F2` | Primary blurple: interactivos, branding y selección |
| `accent-glow` | `#2B2F77` | Indigo profundo para glows y profundidad |
| `success` | `#57F287` | Estado saludable / online |
| `warning` | `#FEE75C` | Uso elevado |
| `danger` | `#ED4245` | Offline / crítico |
| `info` | `#7984F5` | Estados informativos |

En `config.yaml` se redefinen por nombre (sin el prefijo `--color-`):

```yaml
theme:
  colors:
    accent: "#38BDF8"
    bg-base: "#0F172A"
```

### Radios

| Token | Default | Uso |
|---|---|---|
| `sm` | `8px` | Badges pequeños |
| `md` | `12px` | Cards internas, inputs |
| `lg` | `16px` | Cards principales |
| `xl` | `22px` | Modales, drawers |

```yaml
theme:
  radius:
    lg: "20px"     # toda la app, más redondeada
```

Los bloques tipo "terminal" (logs, tablas densas) usan superficies internas
compactas para conservar la lectura técnica.

### Tipografía

| Token | Default | Uso |
|---|---|---|
| `font.sans` | `Inter, Geist, system-ui, sans-serif` | Títulos y body |
| `font.mono` | `'Geist Mono', ui-monospace, monospace` | Métricas, IPs, logs |

Escala (fijada en `@theme` de `style.css`): `xs 11 · sm 13 · base 15 · lg 18 ·
xl 22 · 2xl 28` (px). Datos numéricos siempre en mono.

---

## 4. Dark mode

v1 es **dark only**. `theme.mode` (`dark` | `light`) togglea la clase `dark` en
`<html>` (Tailwind `darkMode: "class"`), dejando la infraestructura lista para
un light theme futuro: bastaría con un segundo set de defaults y la lógica de
cambio. Hoy, cambiar a "light" solo cambia la clase; define una paleta clara en
`theme.colors` si la quieres.

---

## 5. Patrón de componente

Cada módulo de UI sigue el mismo patrón, fácil de replicar:

```
frontend/src/components/<modulo>.js   →  función factory que devuelve un objeto
                                         Alpine (estado + métodos + polling)
index.html                            →  markup con x-data="<modulo>" + tokens
```

Convenciones:

- **Estado de carga:** cada componente expone `loading`, `error`, y cuando
  aplica `stale`. El markup muestra un `skeleton` mientras `loading`.
- **Polling:** el intervalo vive en el componente (`REFRESH_MS`). Clima 15 min;
  Proxmox/Docker 30 s; servicios 60 s.
- **Sin lógica de estilo en JS:** los componentes devuelven *tokens* de color
  (`thresholdColor → "warning"`), no valores hex. El markup los convierte en
  `var(--color-…)`.

### Card base

```html
<div class="card card-hover p-4"> … </div>
```

`.card` (en `style.css`): `bg-surface` + `border` + `rounded-lg`, highlight
interno, sombra contenida y transición de 150ms. `.card-hover:hover` sube el
borde a `accent` y aplica un lift de 1px.

### Dot de estado

```html
<span class="dot dot-online"></span>   <!-- pulse 2s -->
<span class="dot dot-offline"></span>  <!-- estático -->
<span class="dot dot-unknown"></span>
```

### Barra de progreso con umbral

El ancho y el color se calculan desde los datos:

```html
<div class="h-1 rounded-full bg-border overflow-hidden">
  <div :style="`width:${pct}%; background:${barColor(pct)}`"></div>
</div>
```

`barColor` mapea `<60% → accent`, `60–80% → warning`, `>80% → danger`.

---

## 6. Layout

```
Mobile  (<640px):   1 columna, todo apilado
Tablet  (640–1024): 2 columnas
Desktop (>1024px):  sidebar izquierda + contenido principal
```

Estructura (de arriba a abajo): **Header** sticky (logo · búsqueda · clima) →
**Sidebar visible** (clima · calendario) + contenido principal con
**Proxmox** → **Accesos rápidos** → **Docker/Arcane**. En mobile clima y
calendario se apilan primero para quedar visibles al entrar.

---

## 7. Animaciones

Solo donde aportan. Todas se desactivan con `prefers-reduced-motion: reduce`.

| Animación | Dónde | Detalle |
|---|---|---|
| `pulse-dot` | Dots online | opacity 1→0.5→1, 2s infinite |
| `fade-up` | Secciones al cargar | opacity + translateY 8px, 0.4s |
| Transición de superficie | Hover de cards | border-color, shadow y translate 150ms |
| `shimmer` | Skeletons | gradiente animado |
| slide-in | Drawer de Docker | translateX, 300ms ease |

Sin rotaciones, bounces ni partículas.
