// Grid de accesos directos. Consume /api/config y refresca estado cada 60s.

import { getJSON } from "../utils/fetch.js";

const REFRESH_MS = 60 * 1000;

export function services() {
  return {
    items: [],
    loading: true,
    error: false,

    async init() {
      await this.load();
      setInterval(() => this.load(), REFRESH_MS);
      window.addEventListener("polaris:refresh", () => this.load());
    },
    async load() {
      try {
        const res = await getJSON("/api/config");
        this.items = res.services || [];
        this.error = false;
        window.dispatchEvent(new CustomEvent("polaris:updated"));
      } catch {
        this.error = true;
      } finally {
        this.loading = false;
      }
    },
    // Servicios agrupados por categoría, preservando orden de aparición.
    get grouped() {
      const groups = {};
      for (const s of this.items) {
        const cat = s.category || "General";
        (groups[cat] ||= []).push(s);
      }
      return Object.entries(groups).map(([category, items]) => ({ category, items }));
    },
    get sortedItems() {
      return [...this.items].sort((a, b) => {
        const category = (a.category || "General").localeCompare(b.category || "General");
        return category || (a.name || "").localeCompare(b.name || "");
      });
    },
    iconURL(icon) {
      // Sin icono definido -> icono por defecto.
      if (!icon) return "/api/icons/default";
      // URL absoluta (p. ej. un CDN propio) -> tal cual.
      if (/^https?:\/\//.test(icon)) return icon;
      // Slug/nombre -> endpoint con caché en backend (descarga del CDN y guarda local).
      return `/api/icons/${encodeURIComponent(icon.replace(/\.svg$/i, ""))}`;
    },
    serviceInitial(name) {
      return (name || "?").trim().charAt(0).toUpperCase();
    },
    dotClass(status) {
      return { online: "dot-online", offline: "dot-offline" }[status] || "dot-unknown";
    },
  };
}
