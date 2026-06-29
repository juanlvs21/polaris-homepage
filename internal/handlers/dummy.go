package handlers

import (
	"time"

	"polaris-dashboard/internal/clients/arcane"
	"polaris-dashboard/internal/clients/proxmox"
	"polaris-dashboard/internal/clients/unifi"
	"polaris-dashboard/internal/clients/weather"
)

// useDummyData mantiene la API en modo demo mientras se desarrolla la UI.
// Cuando conectemos las integraciones reales, este flag se puede retirar junto
// con los helpers de este archivo.
const useDummyData = true

// showUnifiDummy fuerza datos de prueba en /api/unifi cuando no hay ninguna
// consola UniFi configurada, para poder maquetar el panel.
const showUnifiDummy = true

const gib = 1024 * 1024 * 1024

func dummyWeather(location, units string) *weather.Weather {
	if location == "" {
		location = "Caracas, Venezuela"
	}
	if units == "" {
		units = "metric"
	}
	now := time.Now()
	return &weather.Weather{
		Location:    location,
		TempC:       28,
		TempF:       82,
		FeelsLikeC:  30,
		FeelsLikeF:  86,
		Humidity:    68,
		WindKph:     12,
		WindDir:     "NE",
		Description: "Parcialmente nublado",
		Code:        "116",
		Units:       units,
		Forecast: []weather.ForecastDay{
			{Date: now.Format("2006-01-02"), MaxC: 31, MinC: 23, MaxF: 88, MinF: 73, Code: "116", Hourly: "Parcialmente nublado"},
			{Date: now.AddDate(0, 0, 1).Format("2006-01-02"), MaxC: 30, MinC: 22, MaxF: 86, MinF: 72, Code: "176", Hourly: "Lluvias ligeras"},
			{Date: now.AddDate(0, 0, 2).Format("2006-01-02"), MaxC: 32, MinC: 24, MaxF: 90, MinF: 75, Code: "113", Hourly: "Soleado"},
		},
	}
}

func dummyProxmoxNodes() []proxmox.Node {
	return []proxmox.Node{
		{
			Name:      "Polaris-1",
			Online:    true,
			CPU:       0.34,
			MemUsed:   18 * gib,
			MemTotal:  64 * gib,
			DiskUsed:  720 * gib,
			DiskTotal: 2 * 1024 * gib,
			Uptime:    14*86400 + 3*3600 + 22*60,
			Version:   "pve-manager/8.2.4",
			Guests: []proxmox.Guest{
				{VMID: 101, Name: "docker-vm", Type: "qemu", Status: "running", CPU: 0.21, MemUsed: 9 * gib, MemMax: 16 * gib},
				{VMID: 104, Name: "home-assistant", Type: "qemu", Status: "running", CPU: 0.08, MemUsed: 3 * gib, MemMax: 6 * gib},
				{VMID: 121, Name: "pihole", Type: "lxc", Status: "running", CPU: 0.02, MemUsed: 420 * 1024 * 1024, MemMax: 1 * gib},
				{VMID: 140, Name: "lab-ubuntu", Type: "qemu", Status: "stopped", CPU: 0, MemUsed: 0, MemMax: 8 * gib},
			},
		},
		{
			Name:      "Polaris-2",
			Online:    true,
			CPU:       0.57,
			MemUsed:   42 * gib,
			MemTotal:  96 * gib,
			DiskUsed:  1_420 * gib,
			DiskTotal: 4 * 1024 * gib,
			Uptime:    31*86400 + 7*3600 + 8*60,
			Version:   "pve-manager/8.2.4",
			Guests: []proxmox.Guest{
				{VMID: 201, Name: "nas-services", Type: "qemu", Status: "running", CPU: 0.28, MemUsed: 12 * gib, MemMax: 24 * gib},
				{VMID: 212, Name: "media-stack", Type: "lxc", Status: "running", CPU: 0.43, MemUsed: 7 * gib, MemMax: 12 * gib},
				{VMID: 219, Name: "monitoring", Type: "lxc", Status: "running", CPU: 0.11, MemUsed: 2 * gib, MemMax: 4 * gib},
				{VMID: 230, Name: "backup-runner", Type: "lxc", Status: "stopped", CPU: 0, MemUsed: 0, MemMax: 2 * gib},
			},
		},
	}
}

func dummyContainers() []arcane.Container {
	return flattenContainers(dummyDockerInstances())
}

func dummyDockerInstances() []dockerInstance {
	return []dockerInstance{
		{
			Name: "Arcane · Core",
			Host: "http://arcane-core.webapp.casa",
			Containers: []arcane.Container{
				{ID: "caddy", Name: "caddy", Image: "caddy:2.8", Status: "running", CPU: 1.8, MemUsed: 180 * 1024 * 1024, MemLimit: 1 * gib, Ports: []string{"80:80", "443:443"}, Uptime: 9*86400 + 2*3600},
				{ID: "gitea", Name: "gitea", Image: "gitea/gitea:1.22", Status: "running", CPU: 4.6, MemUsed: 940 * 1024 * 1024, MemLimit: 2 * gib, Ports: []string{"3000:3000", "222:22"}, Uptime: 6*86400 + 18*3600},
				{ID: "postgres", Name: "postgres", Image: "postgres:16-alpine", Status: "running", CPU: 2.2, MemUsed: 740 * 1024 * 1024, MemLimit: 4 * gib, Ports: []string{"5432"}, Uptime: 16*86400 + 5*3600},
				{ID: "redis", Name: "redis", Image: "redis:7-alpine", Status: "running", CPU: 0.7, MemUsed: 96 * 1024 * 1024, MemLimit: 512 * 1024 * 1024, Ports: []string{"6379"}, Uptime: 16*86400 + 5*3600},
			},
		},
		{
			Name: "Arcane · Media",
			Host: "http://arcane-media.webapp.casa",
			Containers: []arcane.Container{
				{ID: "jellyfin", Name: "jellyfin", Image: "jellyfin/jellyfin:10.9", Status: "running", CPU: 12.3, MemUsed: 3 * gib, MemLimit: 8 * gib, Ports: []string{"8096:8096"}, Uptime: 4*86400 + 11*3600},
				{ID: "immich-worker", Name: "immich-worker", Image: "ghcr.io/immich-app/immich-server:release", Status: "paused", CPU: 0, MemUsed: 1 * gib, MemLimit: 6 * gib, Ports: []string{}, Uptime: 2*86400 + 3*3600},
				{ID: "old-wiki", Name: "old-wiki", Image: "requarks/wiki:2", Status: "exited", CPU: 0, MemUsed: 0, MemLimit: 1 * gib, Ports: []string{"8080:3000"}, Uptime: 0},
			},
		},
	}
}

func dummyLogs(id string, tail int) []string {
	lines := []string{
		"2026-06-28T08:14:03Z boot sequence complete",
		"2026-06-28T08:14:04Z loaded configuration from /config",
		"2026-06-28T08:14:05Z healthcheck endpoint registered",
		"2026-06-28T08:15:10Z request GET / responded 200 in 4ms",
		"2026-06-28T08:16:42Z background worker tick completed",
		"2026-06-28T08:17:11Z cache refresh completed",
		"2026-06-28T08:18:28Z request GET /metrics responded 200 in 7ms",
	}
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, id+" | "+line)
	}
	if tail > 0 && tail < len(out) {
		return out[len(out)-tail:]
	}
	return out
}

// dummyUnifiRouters devuelve un router UniFi simulado para maquetar el panel
// sin una consola real conectada.
func dummyUnifiRouters() []unifi.Router {
	return []unifi.Router{
		{
			Name:            "Casa",
			Online:          true,
			Gateway:         "UDM-Pro",
			Version:         "9.0.114",
			Uptime:          21*86400 + 7*3600 + 12*60,
			WanIP:           "190.202.10.42",
			ISP:             "CANTV",
			DownloadMbps:    312.4,
			UploadMbps:      48.7,
			LatencyMs:       9.3,
			RxRateBps:       86 * 1024 * 1024,
			TxRateBps:       12 * 1024 * 1024,
			CPU:             0.18,
			Mem:             0.42,
			ClientsTotal:    37,
			ClientsWired:    11,
			ClientsWireless: 26,
		},
	}
}
