// Package proxmox es un cliente de solo lectura para la REST API de Proxmox VE.
// Se autentica con API Token (no usuario/contraseña) y soporta certificados
// self-signed (común en instalaciones caseras).
package proxmox

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// Client habla con un único nodo Proxmox.
type Client struct {
	http *resty.Client
	name string
	host string
}

// New crea un cliente para un nodo Proxmox.
func New(name, host, tokenID, tokenSecret string, verifyTLS bool) *Client {
	http := resty.New().
		SetTimeout(8*time.Second).
		SetBaseURL(host).
		SetHeader("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", tokenID, tokenSecret))
	if !verifyTLS {
		http.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}
	return &Client{http: http, name: name, host: host}
}

// Name devuelve el nombre configurado del nodo.
func (c *Client) Name() string { return c.name }

// --- Modelos de salida (lo que consume el frontend) ---

// Node agrega las métricas de un nodo y sus guests (VMs + LXC).
type Node struct {
	Name    string  `json:"name"`
	Online  bool    `json:"online"`
	CPU     float64 `json:"cpu"`      // 0..1
	MemUsed int64   `json:"mem_used"` // bytes
	MemTotal int64  `json:"mem_total"`
	DiskUsed int64  `json:"disk_used"`
	DiskTotal int64 `json:"disk_total"`
	Uptime  int64   `json:"uptime"` // segundos
	Version string  `json:"version"`
	Guests  []Guest `json:"guests"`
	Error   string  `json:"error,omitempty"`
}

// Guest es una VM o contenedor LXC.
type Guest struct {
	VMID    int     `json:"vmid"`
	Name    string  `json:"name"`
	Type    string  `json:"type"` // qemu | lxc
	Status  string  `json:"status"`
	CPU     float64 `json:"cpu"`
	MemUsed int64   `json:"mem_used"`
	MemMax  int64   `json:"mem_max"`
}

// --- Modelos de respuesta de la API de Proxmox ---

type statusResp struct {
	Data struct {
		Uptime  int64   `json:"uptime"`
		CPU     float64 `json:"cpu"`
		Memory  struct{ Used, Total int64 } `json:"memory"`
		RootFS  struct{ Used, Total int64 } `json:"rootfs"`
		PVEVersion string `json:"pveversion"`
	} `json:"data"`
}

type guestListResp struct {
	Data []struct {
		VMID   int     `json:"vmid"`
		Name   string  `json:"name"`
		Status string  `json:"status"`
		CPU    float64 `json:"cpu"`
		Mem    int64   `json:"mem"`
		MaxMem int64   `json:"maxmem"`
	} `json:"data"`
}

// nodeName es el nombre interno del nodo en Proxmox (puede diferir del display).
// Para simplificar asumimos que el primer nodo del cluster es el local; en una
// instalación típica casera hay un nodo por host.
func (c *Client) resolveNode() (string, error) {
	var resp struct {
		Data []struct {
			Node string `json:"node"`
		} `json:"data"`
	}
	r, err := c.http.R().SetResult(&resp).Get("/api2/json/nodes")
	if err != nil {
		return "", fmt.Errorf("listando nodos: %w", err)
	}
	if r.IsError() {
		return "", fmt.Errorf("listando nodos: HTTP %d", r.StatusCode())
	}
	if len(resp.Data) == 0 {
		return "", fmt.Errorf("el cluster no devolvió nodos")
	}
	return resp.Data[0].Node, nil
}

// Fetch obtiene el estado completo del nodo: métricas + lista de guests.
func (c *Client) Fetch() Node {
	out := Node{Name: c.name}

	node, err := c.resolveNode()
	if err != nil {
		out.Error = err.Error()
		return out
	}

	var st statusResp
	r, err := c.http.R().SetResult(&st).Get("/api2/json/nodes/" + node + "/status")
	if err != nil || r.IsError() {
		out.Error = fmt.Sprintf("estado del nodo no disponible: %v", err)
		return out
	}
	out.Online = true
	out.CPU = st.Data.CPU
	out.MemUsed, out.MemTotal = st.Data.Memory.Used, st.Data.Memory.Total
	out.DiskUsed, out.DiskTotal = st.Data.RootFS.Used, st.Data.RootFS.Total
	out.Uptime = st.Data.Uptime
	out.Version = st.Data.PVEVersion

	out.Guests = append(out.Guests, c.fetchGuests(node, "qemu")...)
	out.Guests = append(out.Guests, c.fetchGuests(node, "lxc")...)
	return out
}

func (c *Client) fetchGuests(node, kind string) []Guest {
	var resp guestListResp
	r, err := c.http.R().SetResult(&resp).Get("/api2/json/nodes/" + node + "/" + kind)
	if err != nil || r.IsError() {
		return nil
	}
	guests := make([]Guest, 0, len(resp.Data))
	for _, g := range resp.Data {
		guests = append(guests, Guest{
			VMID: g.VMID, Name: g.Name, Type: kind,
			Status: g.Status, CPU: g.CPU,
			MemUsed: g.Mem, MemMax: g.MaxMem,
		})
	}
	return guests
}
