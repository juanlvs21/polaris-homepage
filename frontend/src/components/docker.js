// Grid de contenedores Docker (Arcane) + drawer de detalle con logs.

import { getJSON } from "../utils/fetch.js";
import { formatBytes, formatUptime } from "../utils/format.js";

const REFRESH_MS = 30 * 1000;

export function docker() {
  return {
    instances: [],
    containers: [],
    loading: true,
    error: false,
    stale: false,

    // Drawer
    selected: null,
    selectedInstance: null,
    logs: [],
    logsLoading: false,
    showEnv: false,

    async init() {
      await this.load();
      setInterval(() => this.load(), REFRESH_MS);
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
