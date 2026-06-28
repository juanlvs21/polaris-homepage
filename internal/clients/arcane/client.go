// Package arcane es un cliente de solo lectura para la API de Arcane, el gestor
// de contenedores Docker. Se autentica con API key vía header.
//
// Nota: el shape exacto de la API de Arcane puede variar según versión. Los
// structs aquí modelan un contrato razonable; ajusta los tags `json` a la
// respuesta real de tu instancia si difiere.
package arcane

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// Client habla con la API de Arcane.
type Client struct {
	http *resty.Client
	host string
}

// New crea un cliente de Arcane.
func New(host, apiKey string) *Client {
	return &Client{
		http: resty.New().
			SetTimeout(8*time.Second).
			SetBaseURL(host).
			SetHeader("Authorization", "Bearer "+apiKey),
		host: host,
	}
}

// Container es el modelo que consume el frontend.
type Container struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Image   string   `json:"image"`
	Status  string   `json:"status"` // running | exited | paused
	CPU     float64  `json:"cpu"`    // porcentaje 0..100
	MemUsed int64    `json:"mem_used"`
	MemLimit int64   `json:"mem_limit"`
	Ports   []string `json:"ports"`
	Uptime  int64    `json:"uptime"` // segundos
}

type containersResp struct {
	Data []struct {
		ID       string   `json:"id"`
		Names    []string `json:"names"`
		Name     string   `json:"name"`
		Image    string   `json:"image"`
		State    string   `json:"state"`
		Status   string   `json:"status"`
		CPU      float64  `json:"cpu"`
		MemUsage int64    `json:"memUsage"`
		MemLimit int64    `json:"memLimit"`
		Ports    []string `json:"ports"`
		Started  int64    `json:"startedAt"`
	} `json:"data"`
}

// Containers lista los contenedores con sus stats básicos.
func (c *Client) Containers() ([]Container, error) {
	var raw containersResp
	r, err := c.http.R().SetResult(&raw).Get("/api/containers")
	if err != nil {
		return nil, fmt.Errorf("arcane inalcanzable: %w", err)
	}
	if r.IsError() {
		return nil, fmt.Errorf("arcane respondió %d", r.StatusCode())
	}

	out := make([]Container, 0, len(raw.Data))
	for _, d := range raw.Data {
		name := d.Name
		if name == "" && len(d.Names) > 0 {
			name = d.Names[0]
		}
		uptime := int64(0)
		if d.Started > 0 {
			uptime = time.Now().Unix() - d.Started
		}
		out = append(out, Container{
			ID: d.ID, Name: name, Image: d.Image,
			Status: d.State, CPU: d.CPU,
			MemUsed: d.MemUsage, MemLimit: d.MemLimit,
			Ports: d.Ports, Uptime: uptime,
		})
	}
	return out, nil
}

// Logs devuelve las últimas `tail` líneas de logs de un contenedor.
func (c *Client) Logs(id string, tail int) ([]string, error) {
	var raw struct {
		Logs []string `json:"logs"`
	}
	r, err := c.http.R().
		SetResult(&raw).
		SetQueryParam("tail", fmt.Sprintf("%d", tail)).
		Get("/api/containers/" + id + "/logs")
	if err != nil {
		return nil, fmt.Errorf("arcane inalcanzable: %w", err)
	}
	if r.IsError() {
		return nil, fmt.Errorf("arcane respondió %d", r.StatusCode())
	}
	return raw.Logs, nil
}
