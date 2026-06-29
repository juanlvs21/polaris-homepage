// Grid de contenedores Docker (Arcane) + drawer de detalle con logs.

import { getJSON, postJSON } from "../utils/fetch.js";
import { formatBytes, formatUptime, pct, thresholdColor } from "../utils/format.js";

const REFRESH_MS = 30 * 1000;

export function docker() {
  return {
    instances: [],
    containers: [],
    loading: true,
    error: false,
    stale: false,
    updatedAt: null,
    now: Date.now(),

    // Acción en curso: { id, verb } mientras corre un start/stop/restart.
    acting: null,

    // Drawer
    selected: null,
    selectedInstance: null,
    logs: [],
    logsLoading: false,
    showEnv: false,

    async init() {
      await this.load();
      setInterval(() => this.load(), REFRESH_MS);
      // Mantiene viva la etiqueta relativa ("Updated Xs ago").
      setInterval(() => (this.now = Date.now()), 10 * 1000);
      window.addEventListener("polaris:refresh", () => this.load());
      window.addEventListener("keydown", (e) => {
        if (e.key === "Escape") this.close();
      });
    },
    async load() {
      try {
        const res = await getJSON("/api/docker");
        this.instances = res.instances || [{ name: "Arcane", containers: res.containers || [] }];
        this.containers = res.containers || [];
        this.stale = !!res.stale;
        this.error = false;
        this.updatedAt = Date.now();
        window.dispatchEvent(new CustomEvent("polaris:updated"));
      } catch {
        this.error = true;
      } finally {
        this.loading = false;
      }
    },

    get allContainers() {
      const fromInstances = this.instances.flatMap((instance) =>
        (instance.containers || []).map((container) => ({
          ...container,
          arcaneName: instance.name,
          arcaneHost: instance.host,
          arcaneError: instance.error || "",
        })),
      );
      return fromInstances.length ? fromInstances : this.containers;
    },
    get runningCount() {
      return this.allContainers.filter((c) => c.status === "running").length;
    },
    get totalCount() {
      return this.allContainers.length;
    },
    get totalCpu() {
      return this.allContainers.reduce((sum, c) => sum + (c.cpu || 0), 0);
    },
    get totalMem() {
      return this.allContainers.reduce((sum, c) => sum + (c.mem_used || 0), 0);
    },
    // Etiqueta relativa de la última actualización ("Updated Xs ago").
    get updatedLabel() {
      if (!this.updatedAt) return "—";
      const secs = Math.max(0, Math.round((this.now - this.updatedAt) / 1000));
      if (secs < 5) return "Updated just now";
      if (secs < 60) return `Updated ${secs}s ago`;
      const mins = Math.floor(secs / 60);
      if (mins < 60) return `Updated ${mins}m ago`;
      return `Updated ${Math.floor(mins / 60)}h ago`;
    },

    // Running primero, luego por Arcane y nombre.
    get sorted() {
      return [...this.allContainers].sort((a, b) => {
        const ra = a.status === "running" ? 0 : 1;
        const rb = b.status === "running" ? 0 : 1;
        return ra - rb || (a.arcaneName || "").localeCompare(b.arcaneName || "") || (a.name || "").localeCompare(b.name || "");
      });
    },
    sortedContainers(instance) {
      return [...(instance.containers || [])].sort((a, b) => {
        const ra = a.status === "running" ? 0 : 1;
        const rb = b.status === "running" ? 0 : 1;
        return ra - rb;
      });
    },
    instanceContainers(instance) {
      return instance.containers || [];
    },
    instanceRunning(instance) {
      return this.instanceContainers(instance).filter((c) => c.status === "running").length;
    },
    instanceCpu(instance) {
      return this.instanceContainers(instance).reduce((sum, c) => sum + (c.cpu || 0), 0);
    },
    instanceMemUsed(instance) {
      return this.instanceContainers(instance).reduce((sum, c) => sum + (c.mem_used || 0), 0);
    },
    instanceMemLimit(instance) {
      return this.instanceContainers(instance).reduce((sum, c) => sum + (c.mem_limit || 0), 0);
    },
    instanceMemPct(instance) {
      const total = this.instanceMemLimit(instance);
      return total ? Math.min(100, Math.round((this.instanceMemUsed(instance) / total) * 100)) : 0;
    },

    instanceFor(container) {
      if (!container?.arcaneName) return null;
      return this.instances.find((instance) => instance.name === container.arcaneName) || null;
    },

    async open(container, instance = null) {
      const sourceInstance = instance || this.instanceFor(container);
      this.selectedInstance = sourceInstance;
      this.selected = container;
      document.body.style.overflow = "hidden";
      this.logs = [];
      this.logsLoading = true;
      try {
        const instanceQuery = sourceInstance?.name ? `&instance=${encodeURIComponent(sourceInstance.name)}` : "";
        const res = await getJSON(`/api/docker/${container.id}/logs?tail=50${instanceQuery}`);
        this.logs = res.logs || [];
      } catch {
        this.logs = ["No se pudieron cargar los logs."];
      } finally {
        this.logsLoading = false;
      }
    },
    close() {
      this.selected = null;
      this.selectedInstance = null;
      this.showEnv = false;
      document.body.style.overflow = "";
    },

    // --- Acciones de ciclo de vida (start/stop/restart) ---
    async action(c, verb) {
      if (this.acting) return;
      const inst = this.instanceFor(c);
      const query = inst?.name ? `?instance=${encodeURIComponent(inst.name)}` : "";
      this.acting = { id: c.id, verb };
      try {
        await postJSON(`/api/docker/${c.id}/${verb}${query}`);
        await this.load();
      } catch {
        this.error = true;
      } finally {
        this.acting = null;
      }
    },
    isActing(c, verb) {
      return this.acting?.id === c.id && (!verb || this.acting?.verb === verb);
    },

    // Host corto: hostname de la URL de Arcane, o el nombre de la instancia.
    hostLabel(c) {
      if (c?.arcaneHost) {
        try {
          return new URL(c.arcaneHost).hostname;
        } catch {
          return c.arcaneHost;
        }
      }
      return c?.arcaneName || "local";
    },

    // Iconos por nombre de contenedor (mismo endpoint cacheado que los servicios).
    iconURL(c) {
      const slug = (c?.name || c?.image || "").split(/[:/]/).pop();
      return slug ? `/api/icons/${encodeURIComponent(slug)}` : "/api/icons/default";
    },

    // Color por umbral: CPU sobre 100% nominal, RAM sobre su límite.
    cpuColor(c) {
      return `text-${thresholdColor(Math.min(100, c.cpu || 0))}`;
    },
    memColor(c) {
      const percent = c.mem_limit ? pct(c.mem_used, c.mem_limit) : c.mem_used > 512 * 1024 * 1024 ? 70 : 0;
      return `text-${thresholdColor(percent)}`;
    },

    dotClass(status) {
      return { running: "dot-online", exited: "dot-offline", paused: "dot-unknown" }[status] || "dot-unknown";
    },
    statusLabel(status) {
      return status || "unknown";
    },
    statusPillClass(status) {
      return status === "running"
        ? "border-success/35 bg-success/15 text-success"
        : "border-danger/35 bg-danger/15 text-danger";
    },
    containerInitial(c) {
      return (c?.name || c?.image || "?").trim().charAt(0).toUpperCase();
    },
    mem: (c) => (c.mem_limit ? `${formatBytes(c.mem_used)} / ${formatBytes(c.mem_limit)}` : formatBytes(c.mem_used)),
    bytes: formatBytes,
    cpu: (c) => `${(c.cpu || 0).toFixed(1)}%`,
    uptime: formatUptime,
  };
}
