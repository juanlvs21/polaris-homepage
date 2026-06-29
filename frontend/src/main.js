// Entry point del frontend. Carga el branding/tema antes de arrancar Alpine,
// registra los componentes y expone el branding al markup vía un store global.

import Alpine from "alpinejs";
import "./style.css";

import { loadBranding } from "./theme.js";
import { search } from "./components/search.js";
import { weather } from "./components/weather.js";
import { calendar } from "./components/calendar.js";
import { unifi } from "./components/unifi.js";
import { services } from "./components/services.js";
import { proxmox } from "./components/proxmox.js";
import { docker } from "./components/docker.js";
import { status } from "./components/status.js";

async function bootstrap() {
  const config = await loadBranding();

  // Store global con la identidad, accesible desde cualquier componente.
  Alpine.store("app", {
    branding: config.branding,
    calendar: config.calendar,
    weather: config.weather,
  });

  // Registro de componentes Alpine.
  Alpine.data("search", () => search(config.branding));
  Alpine.data("weather", () => weather(config.weather?.units));
  Alpine.data("calendar", () => calendar(config.calendar?.first_day_of_week));
  Alpine.data("unifi", unifi);
  Alpine.data("services", services);
  Alpine.data("proxmox", proxmox);
  Alpine.data("docker", docker);
  Alpine.data("status", status);

  window.Alpine = Alpine;
  Alpine.start();
}

bootstrap();
