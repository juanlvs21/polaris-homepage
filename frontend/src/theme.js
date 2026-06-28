// theme.js — motor de whitelabeling del frontend.
//
// Obtiene /api/branding (nombre, logo, colores y radios definidos en
// config.yaml) y aplica:
//   1. Las CSS custom properties en :root (paleta + redondeces + fuentes).
//   2. El <title> del documento y el favicon.
// Devuelve el objeto branding para que Alpine lo exponga al markup.

import { getJSON } from "./utils/fetch.js";

const FALLBACK = {
  branding: { name: "Dashboard", tagline: "", search_url: "https://www.google.com/search?q=", search_label: "Buscar…" },
  theme: { colors: {}, radius: {}, font: {} },
  calendar: { first_day_of_week: "monday" },
  weather: { units: "metric" },
};

export async function loadBranding() {
  let data;
  try {
    data = await getJSON("/api/branding");
  } catch {
    data = FALLBACK;
  }
  applyTheme(data.theme || {});
  applyIdentity(data.branding || {});
  return data;
}

// applyTheme escribe cada token como --color-* / --radius-* / --font-* en :root.
function applyTheme(theme) {
  const root = document.documentElement;
  for (const [name, value] of Object.entries(theme.colors || {})) {
    root.style.setProperty(`--color-${name}`, value);
  }
  for (const [name, value] of Object.entries(theme.radius || {})) {
    root.style.setProperty(`--radius-${name}`, value);
  }
  if (theme.font?.sans) root.style.setProperty("--font-sans", theme.font.sans);
  if (theme.font?.mono) root.style.setProperty("--font-mono", theme.font.mono);

  // v1 es darkmode; dejamos la clase lista para un futuro light theme.
  root.classList.toggle("dark", (theme.mode || "dark") === "dark");
}

// applyIdentity fija título y favicon a partir del branding.
function applyIdentity(branding) {
  if (branding.name) document.title = branding.name;
  if (branding.favicon_url) {
    let link = document.querySelector("link[rel~='icon']");
    if (!link) {
      link = document.createElement("link");
      link.rel = "icon";
      document.head.appendChild(link);
    }
    link.href = branding.favicon_url;
  }
}
