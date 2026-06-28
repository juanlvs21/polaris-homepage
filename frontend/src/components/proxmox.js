// Cards de nodos Proxmox + lista de VMs/LXC. Consume /api/proxmox cada 30s.

import { getJSON } from "../utils/fetch.js";
import { formatBytes, formatUptime, pct, thresholdColor } from "../utils/format.js";

const REFRESH_MS = 30 * 1000;
const CHART_W = 640;
const CHART_H = 168;
const SERIES_COLORS = [
  "var(--color-accent)",
  "var(--color-success)",
  "var(--color-info)",
  "var(--color-warning)",
  "var(--color-danger)",
];

export function proxmox() {
  return {
    nodes: [],
    loading: true,
    error: false,
    stale: false,
    updatedAt: null,

    // Sidepanel de detalle de nodo (guardamos el nombre para sobrevivir refrescos)
    selectedName: null,

    async init() {
      await this.load();
      setInterval(() => this.load(), REFRESH_MS);
      window.addEventListener("polaris:refresh", () => this.load());
      window.addEventListener("keydown", (e) => {
        if (e.key === "Escape") this.close();
      });
    },

    get selected() {
      return this.selectedName ? this.nodes.find((n) => n.name === this.selectedName) || null : null;
    },
    open(node) {
      this.selectedName = node.name;
      document.body.style.overflow = "hidden";
    },
    close() {
      this.selectedName = null;
      document.body.style.overflow = "";
    },
    async load() {
      try {
        const res = await getJSON("/api/proxmox");
        this.nodes = res.nodes || [];
        this.stale = !!res.stale;
        this.updatedAt = res.updated_at;
        this.error = false;
        window.dispatchEvent(new CustomEvent("polaris:updated"));
      } catch {
        this.error = true;
      } finally {
        this.loading = false;
      }
    },

    // Guests ordenados: running primero, stopped al final.
    sortedGuests(node) {
      return [...(node.guests || [])].sort((a, b) => {
        const ra = a.status === "running" ? 0 : 1;
        const rb = b.status === "running" ? 0 : 1;
        return ra - rb;
      });
    },
    runningGuestsForNode(node) {
      return (node.guests || []).filter((g) => g.status === "running").length;
    },

    get onlineNodes() {
      return this.nodes.filter((node) => node.online && !node.error).length;
    },
    get totalGuests() {
      return this.nodes.reduce((sum, node) => sum + (node.guests || []).length, 0);
    },
    get runningGuests() {
      return this.nodes.reduce(
        (sum, node) => sum + (node.guests || []).filter((g) => g.status === "running").length,
        0,
      );
    },
    get avgCpuPct() {
      if (!this.nodes.length) return 0;
      return Math.round(this.nodes.reduce((sum, node) => sum + this.cpuPct(node), 0) / this.nodes.length);
    },
    get totalMemUsed() {
      return this.nodes.reduce((sum, node) => sum + (node.mem_used || 0), 0);
    },
    get totalMemTotal() {
      return this.nodes.reduce((sum, node) => sum + (node.mem_total || 0), 0);
    },
    get totalDiskUsed() {
      return this.nodes.reduce((sum, node) => sum + (node.disk_used || 0), 0);
    },
    get totalDiskTotal() {
      return this.nodes.reduce((sum, node) => sum + (node.disk_total || 0), 0);
    },
    get memPercent() {
      return pct(this.totalMemUsed, this.totalMemTotal);
    },
    get diskPercent() {
      return pct(this.totalDiskUsed, this.totalDiskTotal);
    },
    get longestUptime() {
      return Math.max(0, ...this.nodes.map((node) => node.uptime || 0));
    },
    get qemuCount() {
      return this.nodes.reduce((sum, node) => sum + (node.guests || []).filter((g) => g.type === "qemu").length, 0);
    },
    get lxcCount() {
      return this.nodes.reduce((sum, node) => sum + (node.guests || []).filter((g) => g.type === "lxc").length, 0);
    },
    metricRows() {
      return [
        { key: "cpu", label: "CPU", value: (node) => this.cpuPct(node), text: (node) => `${this.cpuPct(node)}%` },
        {
          key: "ram",
          label: "RAM",
          value: (node) => this.memPct(node),
          text: (node) => `${formatBytes(node.mem_used)} / ${formatBytes(node.mem_total)}`,
        },
        {
          key: "disk",
          label: "Disco",
          value: (node) => this.diskPct(node),
          text: (node) => `${formatBytes(node.disk_used)} / ${formatBytes(node.disk_total)}`,
        },
      ];
    },

    cpuPct: (node) => Math.round((node.cpu || 0) * 100),
    memPct: (node) => pct(node.mem_used, node.mem_total),
    diskPct: (node) => pct(node.disk_used, node.disk_total),
    barColor: (percent) => `var(--color-${thresholdColor(percent)})`,
    seriesColor(index) {
      return SERIES_COLORS[index % SERIES_COLORS.length];
    },
    historyPoints(node, metric) {
      const base = metric === "cpu" ? this.cpuPct(node) : metric === "mem" ? this.memPct(node) : this.diskPct(node);
      const nameSeed = [...(node.name || "")].reduce((sum, ch) => sum + ch.charCodeAt(0), 0);
      return Array.from({ length: 28 }, (_, i) => {
        const wave = Math.sin((i + nameSeed) * 0.85) * 7 + Math.cos((i + nameSeed) * 0.32) * 5;
        const drift = metric === "disk" ? i * 0.12 : Math.sin(i * 0.18) * 3;
        return Math.max(2, Math.min(96, Math.round(base + wave + drift)));
      });
    },
    pathFor(points, width = CHART_W, height = CHART_H) {
      if (!points.length) return "";
      const step = width / Math.max(1, points.length - 1);
      return points
        .map((value, index) => {
          const x = Math.round(index * step);
          const y = Math.round(height - (value / 100) * height);
          return `${index === 0 ? "M" : "L"}${x},${y}`;
        })
        .join(" ");
    },
    areaFor(points, width = CHART_W, height = CHART_H) {
      const line = this.pathFor(points, width, height);
      return line ? `${line} L${width},${height} L0,${height} Z` : "";
    },
    chartSeries(metric) {
      return this.nodes.map((node, index) => {
        const points = this.historyPoints(node, metric);
        return {
          name: node.name,
          color: this.seriesColor(index),
          points,
          path: this.pathFor(points),
          area: this.areaFor(points),
          last: points[points.length - 1] || 0,
        };
      });
    },
    metricValue(node, metric) {
      return metric === "cpu" ? this.cpuPct(node) : metric === "mem" ? this.memPct(node) : this.diskPct(node);
    },
    // Mini sparkline para las tarjetas compactas de nodo.
    sparkSvg(node, metric = "cpu") {
      const W = 120;
      const H = 32;
      const points = this.historyPoints(node, metric);
      const path = this.pathFor(points, W, H);
      const area = this.areaFor(points, W, H);
      const color = this.barColor(this.metricValue(node, metric));
      return `<svg viewBox="0 0 ${W} ${H}" class="w-full h-8" preserveAspectRatio="none" aria-hidden="true"><path d="${area}" fill="${color}" fill-opacity="0.12"></path><path d="${path}" fill="none" stroke="${color}" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"></path></svg>`;
    },
    // Gráfica de un solo nodo para el sidepanel de detalle.
    nodeChartSvg(node, metric = "cpu") {
      if (!node) return "";
      const points = this.historyPoints(node, metric);
      const path = this.pathFor(points);
      const area = this.areaFor(points);
      const color = this.barColor(this.metricValue(node, metric));
      const grid = this.chartYTicks()
        .map((tick) => {
          const y = 20 + ((100 - tick) / 100) * CHART_H;
          return `<line x1="0" x2="${CHART_W}" y1="${y}" y2="${y}" stroke="var(--color-border-subtle)" stroke-dasharray="4 6"/><text x="4" y="${y - 4}" fill="var(--color-text-muted)" font-size="10">${tick}</text>`;
        })
        .join("");
      return `<svg viewBox="0 0 ${CHART_W} 220" class="absolute inset-0 w-full h-full" preserveAspectRatio="none" aria-hidden="true"><defs><linearGradient id="proxmoxNodeFill" x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stop-color="${color}" stop-opacity="0.24"/><stop offset="100%" stop-color="${color}" stop-opacity="0.02"/></linearGradient></defs>${grid}<path d="${area}" transform="translate(0 20)" fill="url(#proxmoxNodeFill)"></path><path d="${path}" transform="translate(0 20)" fill="none" stroke="${color}" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"></path></svg>`;
    },
    chartYTicks() {
      return [100, 75, 50, 25, 0];
    },
    chartXLabels() {
      return ["-30m", "-20m", "-10m", "ahora"];
    },
    chartSvg(metric) {
      const grid = this.chartYTicks()
        .map((tick) => {
          const y = 20 + ((100 - tick) / 100) * CHART_H;
          return `<line x1="0" x2="${CHART_W}" y1="${y}" y2="${y}" stroke="var(--color-border-subtle)" stroke-dasharray="4 6"/><text x="4" y="${y - 4}" fill="var(--color-text-muted)" font-size="10">${tick}</text>`;
        })
        .join("");
      const series = this.chartSeries(metric)
        .map((serie, index) => {
          const area =
            index === 0
              ? `<path d="${serie.area}" transform="translate(0 20)" fill="url(#proxmoxChartFill)"></path>`
              : "";
          return `${area}<path d="${serie.path}" transform="translate(0 20)" fill="none" stroke="${serie.color}" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"></path>`;
        })
        .join("");
      return `<svg viewBox="0 0 ${CHART_W} 220" class="absolute inset-0 w-full h-full" preserveAspectRatio="none" aria-hidden="true"><defs><linearGradient id="proxmoxChartFill" x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stop-color="var(--color-accent)" stop-opacity="0.24"/><stop offset="100%" stop-color="var(--color-accent)" stop-opacity="0.02"/></linearGradient></defs>${grid}${series}</svg>`;
    },
    donutStyle(percent) {
      return `background: conic-gradient(var(--color-accent) ${percent}%, var(--color-bg-elevated) 0)`;
    },
    nodeHealth(node) {
      if (node.error || !node.online) return "offline";
      const maxMetric = Math.max(this.cpuPct(node), this.memPct(node), this.diskPct(node));
      if (maxMetric >= 80) return "critical";
      if (maxMetric >= 60) return "warning";
      return "healthy";
    },
    healthClass(node) {
      return {
        healthy: "text-success",
        warning: "text-warning",
        critical: "text-danger",
        offline: "text-danger",
      }[this.nodeHealth(node)];
    },
    healthLabel(node) {
      return {
        healthy: "saludable",
        warning: "atencion",
        critical: "critico",
        offline: "offline",
      }[this.nodeHealth(node)];
    },
    bytes: formatBytes,
    uptime: formatUptime,
    guestCpu: (g) => `${Math.round((g.cpu || 0) * 100)}%`,
  };
}
