// Widget de clima. Consume /api/weather y refresca cada 15 minutos.

import { getJSON } from "../utils/fetch.js";

const REFRESH_MS = 15 * 60 * 1000;

export function weather(units = "metric") {
  return {
    data: null,
    loading: true,
    error: false,
    stale: false,
    units,

    async init() {
      await this.load();
      setInterval(() => this.load(), REFRESH_MS);
      window.addEventListener("polaris:refresh", () => this.load());
    },
    async load() {
      try {
        const res = await getJSON("/api/weather");
        this.data = res.weather;
        this.stale = !!res.stale;
        this.error = false;
        window.dispatchEvent(new CustomEvent("polaris:updated"));
      } catch {
        this.error = true;
      } finally {
        this.loading = false;
      }
    },
    temp(d = this.data) {
      if (!d) return "—";
      return this.units === "imperial" ? `${d.temp_f}°F` : `${d.temp_c}°C`;
    },
    feels() {
      if (!this.data) return "—";
      return this.units === "imperial" ? `${this.data.feels_like_f}°F` : `${this.data.feels_like_c}°C`;
    },
    fMax(day) {
      return this.units === "imperial" ? `${day.max_f}°` : `${day.max_c}°`;
    },
    fMin(day) {
      return this.units === "imperial" ? `${day.min_f}°` : `${day.min_c}°`;
    },
    icon(code) {
      // Mapeo simple de weatherCode wttr.in -> emoji.
      const c = parseInt(code || "0", 10);
      if ([113].includes(c)) return "☀️";
      if ([116, 119, 122].includes(c)) return "⛅";
      if ([143, 248, 260].includes(c)) return "🌫️";
      if (c >= 176 && c <= 377 && c < 300) return "🌧️";
      if (c >= 300 && c < 350) return "🌧️";
      if (c >= 350) return "❄️";
      return "☁️";
    },
  };
}
