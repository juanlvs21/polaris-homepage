// Indicador de "última actualización" + botón para refrescar toda la data.
//
// Coordinación entre componentes vía eventos globales:
//   - "polaris:refresh" -> cada componente de datos vuelve a hacer load().
//   - "polaris:updated" -> cada load() exitoso (auto o manual) lo emite, y aquí
//     registramos la marca de tiempo más reciente.

export function status() {
  return {
    lastUpdated: null,
    now: Date.now(),
    refreshing: false,

    init() {
      window.addEventListener("polaris:updated", () => {
        this.lastUpdated = Date.now();
        this.refreshing = false;
      });
      // Mantiene viva la etiqueta relativa ("hace Xs").
      setInterval(() => (this.now = Date.now()), 10 * 1000);
    },

    refresh() {
      if (this.refreshing) return;
      this.refreshing = true;
      window.dispatchEvent(new CustomEvent("polaris:refresh"));
      // Fallback: no dejar el spinner girando si alguna fuente falla.
      setTimeout(() => (this.refreshing = false), 5000);
    },

    get label() {
      if (!this.lastUpdated) return "Sin datos";
      const secs = Math.max(0, Math.round((this.now - this.lastUpdated) / 1000));
      if (secs < 5) return "ahora mismo";
      if (secs < 60) return `hace ${secs}s`;
      const mins = Math.floor(secs / 60);
      if (mins < 60) return `hace ${mins} min`;
      const hours = Math.floor(mins / 60);
      return `hace ${hours} h`;
    },
  };
}
