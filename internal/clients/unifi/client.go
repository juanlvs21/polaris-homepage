// Package unifi es un cliente de solo lectura para la API oficial de UniFi
// Network Integration (la que se habilita en consolas UniFi OS bajo
// Settings → Control Plane → Integrations).
//
// Se autentica con una API key vía el header `X-API-KEY` (sin login/cookies ni
// CSRF), igual de simple que el token de Proxmox. Soporta el certificado
// self-signed de la consola.
//
// Base URL: https://{console}/proxy/network/integration/v1
//
// Nota: el shape exacto de algunos campos (sobre todo las estadísticas del
// gateway) varía según la versión de UniFi Network. Los structs aquí modelan un
// contrato razonable basado en la API v9.3+; si tu instancia difiere, ajusta
// los tags `json`. El cliente es defensivo: si un campo no llega, ese dato sale
// en cero y el resto del panel sigue funcionando.
package unifi

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// Client habla con una única consola UniFi OS.
type Client struct {
	http   *resty.Client
	name   string
	host   string
	siteID string // opcional: si está vacío se resuelve al primer site
}

// New crea un cliente de UniFi.
//
//	name      Etiqueta para mostrar (ej. "Casa").
//	host      URL base de la consola (ej. "https://192.168.1.1").
//	apiKey    API key generada en la consola.
//	siteID    ID del site a monitorear; vacío = primer site disponible.
//	verifyTLS Validar el certificado (false para el self-signed por defecto).
func New(name, host, apiKey, siteID string, verifyTLS bool) *Client {
	host = strings.TrimRight(host, "/")
	http := resty.New().
		SetTimeout(8*time.Second).
		SetBaseURL(host+"/proxy/network/integration/v1").
		SetHeader("X-API-KEY", apiKey).
		SetHeader("Accept", "application/json")
	if !verifyTLS {
		http.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}
	return &Client{http: http, name: name, host: host, siteID: siteID}
}

// Name devuelve el nombre configurado de la consola.
func (c *Client) Name() string { return c.name }

// --- Modelo de salida (lo que consume el frontend) ---

// Router agrega el estado de la red: gateway + WAN + clientes.
type Router struct {
	Name    string `json:"name"`    // etiqueta de la consola
	Online  bool   `json:"online"`  // gateway reportando estado
	Gateway string `json:"gateway"` // modelo del gateway (ej. "UDM-Pro")
	Version string `json:"version"` // versión de UniFi Network
	Uptime  int64  `json:"uptime"`  // segundos

	// WAN / "velocidad del router"
	WanIP        string  `json:"wan_ip"`        // IP pública WAN
	ISP          string  `json:"isp"`           // proveedor (si lo expone)
	DownloadMbps float64 `json:"download_mbps"` // último speedtest (bajada)
	UploadMbps   float64 `json:"upload_mbps"`   // último speedtest (subida)
	LatencyMs    float64 `json:"latency_ms"`    // latencia WAN
	RxRateBps    int64   `json:"rx_rate_bps"`   // throughput WAN en vivo (descarga)
	TxRateBps    int64   `json:"tx_rate_bps"`   // throughput WAN en vivo (subida)

	// Salud del gateway
	CPU float64 `json:"cpu"` // 0..1
	Mem float64 `json:"mem"` // 0..1

	// Clientes
	ClientsTotal    int `json:"clients_total"`
	ClientsWired    int `json:"clients_wired"`
	ClientsWireless int `json:"clients_wireless"`

	Error string `json:"error,omitempty"`
}

// --- Modelos de respuesta de la API de UniFi ---

// envelope es el sobre paginado estándar de la Integration API.
type envelope[T any] struct {
	Offset     int `json:"offset"`
	Limit      int `json:"limit"`
	Count      int `json:"count"`
	TotalCount int `json:"totalCount"`
	Data       []T `json:"data"`
}

type siteResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type deviceResp struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Model    string `json:"model"`
	State    string `json:"state"` // ONLINE | OFFLINE | ...
	Features []string `json:"features"`
	// Algunos firmwares devuelven el rol del dispositivo aquí:
	Type    string `json:"type"`
	Version string `json:"version"`
	Uptime  int64  `json:"uptimeSec"`
	IPv4    string `json:"ipAddress"`
}

// deviceStats modela /devices/{id}/statistics/latest para el gateway.
type deviceStats struct {
	UptimeSec         int64   `json:"uptimeSec"`
	CPUUtilizationPct float64 `json:"cpuUtilizationPct"`
	MemoryUtilizationPct float64 `json:"memoryUtilizationPct"`
	Uplink            struct {
		TxRateBps int64 `json:"txRateBps"`
		RxRateBps int64 `json:"rxRateBps"`
	} `json:"uplink"`
	// Resultado del último speedtest del gateway, cuando está disponible.
	Internet struct {
		DownlinkKbps float64 `json:"downlinkKbps"`
		UplinkKbps   float64 `json:"uplinkKbps"`
		LatencyMs    float64 `json:"latencyAvgMs"`
		ISP          string  `json:"ispName"`
		WanIP        string  `json:"wanIp"`
	} `json:"internet"`
}

type clientResp struct {
	ID         string `json:"id"`
	Type       string `json:"type"`       // WIRED | WIRELESS | VPN
	Connection string `json:"connection"` // algunos firmwares usan este campo
}

// Fetch obtiene el estado completo de la red. Nunca retorna error: cualquier
// fallo se expresa en Router.Error para que el handler lo cachee/degrade igual
// que hace con Proxmox.
func (c *Client) Fetch() Router {
	out := Router{Name: c.name}

	site, err := c.resolveSite()
	if err != nil {
		out.Error = err.Error()
		return out
	}

	gw, err := c.findGateway(site)
	if err != nil {
		out.Error = err.Error()
		return out
	}
	out.Online = strings.EqualFold(gw.State, "ONLINE")
	out.Gateway = firstNonEmpty(gw.Model, gw.Name)
	out.Version = gw.Version
	out.Uptime = gw.Uptime
	out.WanIP = gw.IPv4

	// Estadísticas en vivo del gateway (throughput, cpu/mem, speedtest).
	if st, ok := c.gatewayStats(site, gw.ID); ok {
		if st.UptimeSec > 0 {
			out.Uptime = st.UptimeSec
		}
		out.CPU = st.CPUUtilizationPct / 100
		out.Mem = st.MemoryUtilizationPct / 100
		out.RxRateBps = st.Uplink.RxRateBps
		out.TxRateBps = st.Uplink.TxRateBps
		out.DownloadMbps = st.Internet.DownlinkKbps / 1000
		out.UploadMbps = st.Internet.UplinkKbps / 1000
		out.LatencyMs = st.Internet.LatencyMs
		if st.Internet.ISP != "" {
			out.ISP = st.Internet.ISP
		}
		if st.Internet.WanIP != "" {
			out.WanIP = st.Internet.WanIP
		}
	}

	// Conteo de clientes conectados.
	c.countClients(site, &out)

	return out
}

// resolveSite devuelve el site configurado o, en su defecto, el primero.
func (c *Client) resolveSite() (string, error) {
	if c.siteID != "" {
		return c.siteID, nil
	}
	var resp envelope[siteResp]
	r, err := c.http.R().SetResult(&resp).Get("/sites")
	if err != nil {
		return "", fmt.Errorf("listando sites: %w", err)
	}
	if r.IsError() {
		return "", fmt.Errorf("listando sites: HTTP %d", r.StatusCode())
	}
	if len(resp.Data) == 0 {
		return "", fmt.Errorf("la consola no devolvió ningún site")
	}
	return resp.Data[0].ID, nil
}

// findGateway localiza el dispositivo gateway/router del site. Se identifica por
// el feature/role "gateway"; si la API no lo etiqueta, cae al primer dispositivo
// cuyo modelo parezca un gateway (UDM/UXG/UCG/USG).
func (c *Client) findGateway(site string) (deviceResp, error) {
	var resp envelope[deviceResp]
	r, err := c.http.R().SetResult(&resp).Get("/sites/" + site + "/devices")
	if err != nil {
		return deviceResp{}, fmt.Errorf("listando dispositivos: %w", err)
	}
	if r.IsError() {
		return deviceResp{}, fmt.Errorf("listando dispositivos: HTTP %d", r.StatusCode())
	}
	for _, d := range resp.Data {
		if hasFeature(d.Features, "gateway") || strings.EqualFold(d.Type, "gateway") {
			return d, nil
		}
	}
	for _, d := range resp.Data {
		if looksLikeGateway(d.Model) {
			return d, nil
		}
	}
	if len(resp.Data) == 0 {
		return deviceResp{}, fmt.Errorf("el site no tiene dispositivos")
	}
	return deviceResp{}, fmt.Errorf("no se encontró un gateway en el site")
}

// gatewayStats consulta las métricas en vivo del gateway. Devuelve ok=false si
// no están disponibles (firmware sin ese endpoint), sin tratarlo como error.
func (c *Client) gatewayStats(site, deviceID string) (deviceStats, bool) {
	var st deviceStats
	r, err := c.http.R().SetResult(&st).Get("/sites/" + site + "/devices/" + deviceID + "/statistics/latest")
	if err != nil || r.IsError() {
		return deviceStats{}, false
	}
	return st, true
}

// countClients suma los clientes conectados, separando cableados de inalámbricos.
func (c *Client) countClients(site string, out *Router) {
	var resp envelope[clientResp]
	r, err := c.http.R().
		SetQueryParam("limit", "200").
		SetResult(&resp).
		Get("/sites/" + site + "/clients")
	if err != nil || r.IsError() {
		return
	}
	total := resp.TotalCount
	if total == 0 {
		total = len(resp.Data)
	}
	out.ClientsTotal = total
	for _, cl := range resp.Data {
		kind := strings.ToUpper(firstNonEmpty(cl.Type, cl.Connection))
		switch {
		case strings.Contains(kind, "WIRED"):
			out.ClientsWired++
		case strings.Contains(kind, "WIRELESS"), strings.Contains(kind, "WIFI"):
			out.ClientsWireless++
		}
	}
}

// --- helpers ---

func hasFeature(features []string, want string) bool {
	for _, f := range features {
		if strings.EqualFold(f, want) {
			return true
		}
	}
	return false
}

func looksLikeGateway(model string) bool {
	m := strings.ToUpper(model)
	for _, p := range []string{"UDM", "UXG", "UCG", "USG", "UDW", "UDR"} {
		if strings.Contains(m, p) {
			return true
		}
	}
	return false
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
