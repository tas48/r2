package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	defaultGeocodeBase  = "https://geocoding-api.open-meteo.com/v1/search"
	defaultForecastBase = "https://api.open-meteo.com/v1/forecast"
)

type coords struct {
	lat float64
	lon float64
}

// OpenMeteo is a Provider backed by Open-Meteo. It needs no API key and caches
// geocoding results per city.
type OpenMeteo struct {
	client        *http.Client
	geocodeBase   string
	forecastBase  string
	rainThreshold int
	horizonHours  int
	coords        map[string]coords
}

func NewOpenMeteo(client *http.Client, rainThreshold, horizonHours int) *OpenMeteo {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &OpenMeteo{
		client:        client,
		geocodeBase:   defaultGeocodeBase,
		forecastBase:  defaultForecastBase,
		rainThreshold: rainThreshold,
		horizonHours:  horizonHours,
		coords:        make(map[string]coords),
	}
}

func (o *OpenMeteo) Forecast(ctx context.Context, city string) (Forecast, error) {
	location, err := o.geocode(ctx, city)
	if err != nil {
		return Forecast{}, err
	}

	query := url.Values{
		"latitude":      {strconv.FormatFloat(location.lat, 'f', 4, 64)},
		"longitude":     {strconv.FormatFloat(location.lon, 'f', 4, 64)},
		"current":       {"temperature_2m"},
		"hourly":        {"precipitation_probability"},
		"forecast_days": {"1"},
		"timezone":      {"auto"},
	}
	var payload struct {
		Current struct {
			Temperature float64 `json:"temperature_2m"`
		} `json:"current"`
		Hourly struct {
			Precipitation []int `json:"precipitation_probability"`
		} `json:"hourly"`
	}
	if err := o.getJSON(ctx, o.forecastBase+"?"+query.Encode(), &payload); err != nil {
		return Forecast{}, fmt.Errorf("fetching forecast for %s: %w", city, err)
	}

	prob := maxWithin(payload.Hourly.Precipitation, o.horizonHours)
	return Forecast{
		RainProb:   prob,
		TempC:      payload.Current.Temperature,
		RainLikely: prob >= o.rainThreshold,
		FetchedAt:  time.Now(),
	}, nil
}

func (o *OpenMeteo) geocode(ctx context.Context, city string) (coords, error) {
	if cached, ok := o.coords[city]; ok {
		return cached, nil
	}
	query := url.Values{
		"name":     {city},
		"count":    {"1"},
		"language": {"pt"},
		"format":   {"json"},
	}
	var payload struct {
		Results []struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"results"`
	}
	if err := o.getJSON(ctx, o.geocodeBase+"?"+query.Encode(), &payload); err != nil {
		return coords{}, fmt.Errorf("geocoding %s: %w", city, err)
	}
	if len(payload.Results) == 0 {
		return coords{}, fmt.Errorf("geocoding %s: no results", city)
	}
	location := coords{lat: payload.Results[0].Latitude, lon: payload.Results[0].Longitude}
	o.coords[city] = location
	return location, nil
}

func (o *OpenMeteo) getJSON(ctx context.Context, endpoint string, dst any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	response, err := o.client.Do(request)
	if err != nil {
		return fmt.Errorf("requesting %s: %w", endpoint, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from %s", response.StatusCode, endpoint)
	}
	if err := json.NewDecoder(response.Body).Decode(dst); err != nil {
		return fmt.Errorf("decoding response from %s: %w", endpoint, err)
	}
	return nil
}

func maxWithin(values []int, limit int) int {
	if limit <= 0 || limit > len(values) {
		limit = len(values)
	}
	best := 0
	for _, v := range values[:limit] {
		if v > best {
			best = v
		}
	}
	return best
}
