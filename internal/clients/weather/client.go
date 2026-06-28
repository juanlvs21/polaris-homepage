// Package weather consulta el clima actual y el pronóstico desde wttr.in.
// wttr.in expone un JSON (formato j1) sin necesidad de API key.
package weather

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

// Client consulta wttr.in para una ubicación dada.
type Client struct {
	http     *resty.Client
	location string
	units    string
}

// New crea un cliente de clima para la ubicación y unidades indicadas.
func New(location, units string) *Client {
	return &Client{
		http: resty.New().
			SetTimeout(8 * time.Second).
			SetHeader("User-Agent", "curl/8"), // wttr.in responde JSON a clientes "curl-like"
		location: location,
		units:    units,
	}
}

// Weather es la respuesta tipada que consume el frontend.
type Weather struct {
	Location    string         `json:"location"`
	TempC       int            `json:"temp_c"`
	TempF       int            `json:"temp_f"`
	FeelsLikeC  int            `json:"feels_like_c"`
	FeelsLikeF  int            `json:"feels_like_f"`
	Humidity    int            `json:"humidity"`
	WindKph     int            `json:"wind_kph"`
	WindDir     string         `json:"wind_dir"`
	Description string         `json:"description"`
	Code        string         `json:"code"`
	Forecast    []ForecastDay  `json:"forecast"`
	Units       string         `json:"units"`
}

// ForecastDay es el pronóstico de un día.
type ForecastDay struct {
	Date   string `json:"date"`
	MaxC   int    `json:"max_c"`
	MinC   int    `json:"min_c"`
	MaxF   int    `json:"max_f"`
	MinF   int    `json:"min_f"`
	Code   string `json:"code"`
	Hourly string `json:"description"`
}

// wttr.in j1 — solo modelamos los campos que usamos.
type wttrResponse struct {
	CurrentCondition []struct {
		TempC          string                `json:"temp_C"`
		TempF          string                `json:"temp_F"`
		FeelsLikeC     string                `json:"FeelsLikeC"`
		FeelsLikeF     string                `json:"FeelsLikeF"`
		Humidity       string                `json:"humidity"`
		WindspeedKmph  string                `json:"windspeedKmph"`
		WindDir16Point string                `json:"winddir16Point"`
		WeatherCode    string                `json:"weatherCode"`
		WeatherDesc    []wttrValue           `json:"weatherDesc"`
	} `json:"current_condition"`
	Weather []struct {
		Date    string `json:"date"`
		MaxtempC string `json:"maxtempC"`
		MintempC string `json:"mintempC"`
		MaxtempF string `json:"maxtempF"`
		MintempF string `json:"mintempF"`
		Hourly  []struct {
			WeatherCode string      `json:"weatherCode"`
			WeatherDesc []wttrValue `json:"weatherDesc"`
		} `json:"hourly"`
	} `json:"weather"`
}

type wttrValue struct {
	Value string `json:"value"`
}

// Fetch obtiene el clima actual. Devuelve error si la API falla o la respuesta
// es inesperada; el handler decide si usar un valor cacheado como fallback.
func (c *Client) Fetch() (*Weather, error) {
	var raw wttrResponse
	resp, err := c.http.R().
		SetResult(&raw).
		SetQueryParam("format", "j1").
		Get(fmt.Sprintf("https://wttr.in/%s", c.location))
	if err != nil {
		return nil, fmt.Errorf("wttr.in inalcanzable: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("wttr.in respondió %d", resp.StatusCode())
	}
	if len(raw.CurrentCondition) == 0 {
		return nil, fmt.Errorf("wttr.in devolvió una respuesta vacía")
	}

	cur := raw.CurrentCondition[0]
	w := &Weather{
		Location:    c.location,
		TempC:       atoi(cur.TempC),
		TempF:       atoi(cur.TempF),
		FeelsLikeC:  atoi(cur.FeelsLikeC),
		FeelsLikeF:  atoi(cur.FeelsLikeF),
		Humidity:    atoi(cur.Humidity),
		WindKph:     atoi(cur.WindspeedKmph),
		WindDir:     cur.WindDir16Point,
		Code:        cur.WeatherCode,
		Units:       c.units,
		Description: firstDesc(cur.WeatherDesc),
	}

	for i, day := range raw.Weather {
		if i >= 3 {
			break
		}
		code := ""
		desc := ""
		if len(day.Hourly) > 4 { // mediodía aprox
			code = day.Hourly[4].WeatherCode
			desc = firstDesc(day.Hourly[4].WeatherDesc)
		}
		w.Forecast = append(w.Forecast, ForecastDay{
			Date: day.Date,
			MaxC: atoi(day.MaxtempC), MinC: atoi(day.MintempC),
			MaxF: atoi(day.MaxtempF), MinF: atoi(day.MintempF),
			Code: code, Hourly: desc,
		})
	}

	return w, nil
}

func firstDesc(v []wttrValue) string {
	if len(v) == 0 {
		return ""
	}
	return v[0].Value
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
