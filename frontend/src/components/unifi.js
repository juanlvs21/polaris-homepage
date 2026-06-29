// Panel UniFi: velocidad WAN, throughput en vivo, clientes y salud del gateway.
// Consume /api/unifi cada 15s. Pensado para la columna lateral (compacto).

import { getJSON } from "../utils/fetch.js";
import { formatBitrate, formatUptime, thresholdColor } from "../utils/format.js";

const REFRESH_MS = 15 * 1000;

export function unifi() {
  return {
    routers: [],
    loading: true,
    error: false,
    stale: false,
    updatedAt: null,

    async init() {
      await this.load();
      setInterval(() => this.load(), REFRESH_MS);
      window.addEventListener("polaris:refresh", () => this.load());
    },

    async load() {
      try {
        const res = await getJSON("/api/unifi");
        this.routers = res.routers || [];
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

    // Router principal (la consola suele ser una sola).
    get router() {
      return this.routers[0] || null;
    },
    get hasRouter() {
      return !!this.router && !this.router.error;
    },
    get online() {
      return this.hasRouter && this.router.online;
    },

    // Velocidad / throughput formateados.
    downRate() {
      return formatBitrate(this.router?.rx_rate_bps);
    },
    upRate() {
      return formatBitrate(this.router?.tx_rate_bps);
    },
    downloadMbps() {
      return this.router?.download_mbps ? `${this.router.download_mbps.toFixed(0)} Mbps` : "—";
    },
    uploadMbps() {
      return this.router?.upload_mbps ? `${this.router.upload_mbps.toFixed(0)} Mbps` : "—";
    },
    latency() {
      return this.router?.latency_ms ? `${Math.round(this.router.latency_ms)} ms` : "—";
    },

    // Salud del gateway.
    cpuPct() {
      return Math.round((this.router?.cpu || 0) * 100);
    },
    memPct() {
      return Math.round((this.router?.mem || 0) * 100);
    },
    barColor(percent) {
      return `var(--color-${thresholdColor(percent)})`;
    },

    clients() {
      return this.router?.clients_total || 0;
    },
    clientsBreakdown() {
      const r = this.router;
      if (!r) return "";
      return `${r.clients_wired || 0} cable · ${r.clients_wireless || 0} wifi`;
    },
    uptime() {
      return formatUptime(this.router?.uptime);
    },
  };
}
