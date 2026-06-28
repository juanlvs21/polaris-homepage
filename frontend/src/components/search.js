// Barra de búsqueda. Redirige al motor configurado en branding.search_url.

export function search(branding) {
  return {
    query: "",
    searchUrl: branding?.search_url || "https://www.google.com/search?q=",
    label: branding?.search_label || "Buscar en la web…",
    submit() {
      const q = this.query.trim();
      if (!q) return;
      window.open(this.searchUrl + encodeURIComponent(q), "_blank");
      this.query = "";
    },
    focusOnDesktop() {
      if (window.matchMedia("(min-width: 768px)").matches) {
        this.$nextTick(() => this.$refs.input?.focus());
      }
    },
  };
}
