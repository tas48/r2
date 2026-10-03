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

// Location is a resolved place with coordinates.
type Location struct {
	Name      string
	Latitude  float64
	Longitude float64
}

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
	if horizonHours <= 0 {
		horizonHours = 6
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

// Forecast resolves the city, fetches everything Open-Meteo offers and derives
// local alerts. It embeds the resolved location name in the result.
func (o *OpenMeteo) Forecast(ctx context.Context, city string) (Forecast, error) {
	location, err := o.geocode(ctx, city)
	if err != nil {
		return Forecast{}, err
	}
	return o.forecastAt(ctx, location.Name, location.Latitude, location.Longitude)
}

// ForecastAt fetches a forecast directly for coordinates, skipping geocoding.
func (o *OpenMeteo) ForecastAt(ctx context.Context, name string, lat, lon float64) (Forecast, error) {
	return o.forecastAt(ctx, name, lat, lon)
}

func (o *OpenMeteo) forecastAt(ctx context.Context, name string, lat, lon float64) (Forecast, error) {
	query := url.Values{
		"latitude":  {strconv.FormatFloat(lat, 'f', 4, 64)},
		"longitude": {strconv.FormatFloat(lon, 'f', 4, 64)},
		"current": {
			"temperature_2m,apparent_temperature,relative_humidity_2m," +
				"wind_speed_10m,uv_index,is_day,precipitation_probability",
		},
		"hourly": {
			"temperature_2m,precipitation_probability,wind_speed_10m,uv_index",
		},
		"daily": {
			"temperature_2m_max,temperature_2m_min,precipitation_probability_max," +
				"wind_speed_10m_max,uv_index_max",
		},
		"forecast_days": {"4"},
		"timezone":      {"auto"},
	}
	var payload struct {
		Current struct {
			Temperature float64 `json:"temperature_2m"`
			Apparent    float64 `json:"apparent_temperature"`
			Humidity    int     `json:"relative_humidity_2m"`
			WindSpeed   float64 `json:"wind_speed_10m"`
			UVIndex     float64 `json:"uv_index"`
			IsDay       int     `json:"is_day"`
			PrecipProb  int     `json:"precipitation_probability"`
		} `json:"current"`
		Hourly struct {
			Time        []string  `json:"time"`
			Temperature []float64 `json:"temperature_2m"`
			PrecipProb  []int     `json:"precipitation_probability"`
			WindSpeed   []float64 `json:"wind_speed_10m"`
			UVIndex     []float64 `json:"uv_index"`
		} `json:"hourly"`
		Daily struct {
			Time       []string  `json:"time"`
			TempMax    []float64 `json:"temperature_2m_max"`
			TempMin    []float64 `json:"temperature_2m_min"`
			PrecipProb []int     `json:"precipitation_probability_max"`
			WindSpeed  []float64 `json:"wind_speed_10m_max"`
			UVIndexMax []float64 `json:"uv_index_max"`
		} `json:"daily"`
	}
	endpoint := o.forecastBase + "?" + query.Encode()
	if err := o.getJSON(ctx, endpoint, &payload); err != nil {
		return Forecast{}, fmt.Errorf("fetching forecast for %s: %w", name, err)
	}

	hours := buildHours(payload.Hourly.Time, payload.Hourly.Temperature, payload.Hourly.PrecipProb, payload.Hourly.WindSpeed, payload.Hourly.UVIndex)
	forecast := Forecast{
		Location:   name,
		RainProb:   payload.Current.PrecipProb,
		TempC:      payload.Current.Temperature,
		FeelsLikeC: payload.Current.Apparent,
		WindKph:    payload.Current.WindSpeed,
		Humidity:   payload.Current.Humidity,
		UVIndex:    payload.Current.UVIndex,
		IsDay:      payload.Current.IsDay == 1,
		NextHours:  nextHours(hours, o.horizonHours),
		Today:      hours,
		Daily:      buildDays(payload.Daily.Time, payload.Daily.TempMax, payload.Daily.TempMin, payload.Daily.PrecipProb, payload.Daily.WindSpeed, payload.Daily.UVIndexMax),
		FetchedAt:  time.Now(),
	}
	forecast.RainProb = maxRainWindow(hours, o.horizonHours, payload.Current.PrecipProb)
	forecast.RainLikely = forecast.RainProb >= o.rainThreshold
	forecast.Alerts = deriveAlerts(forecast)
	return forecast, nil
}

func (o *OpenMeteo) geocode(ctx context.Context, city string) (Location, error) {
	if cached, ok := o.coords[city]; ok {
		return Location{Name: city, Latitude: cached.lat, Longitude: cached.lon}, nil
	}
	query := url.Values{
		"name":     {city},
		"count":    {"1"},
		"language": {"pt"},
		"format":   {"json"},
	}
	var payload struct {
		Results []struct {
			Name      string  `json:"name"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"results"`
	}
	if err := o.getJSON(ctx, o.geocodeBase+"?"+query.Encode(), &payload); err != nil {
		return Location{}, fmt.Errorf("geocoding %s: %w", city, err)
	}
	if len(payload.Results) == 0 {
		return Location{}, fmt.Errorf("geocoding %s: no results", city)
	}
	result := payload.Results[0]
	o.coords[city] = coords{lat: result.Latitude, lon: result.Longitude}
	name := result.Name
	if name == "" {
		name = city
	}
	return Location{Name: name, Latitude: result.Latitude, Longitude: result.Longitude}, nil
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

func buildHours(times []string, temps []float64, rain []int, wind []float64, uv []float64) []HourSample {
	n := len(times)
	hours := make([]HourSample, 0, n)
	for i := 0; i < n; i++ {
		t, err := time.ParseInLocation("2006-01-02T15:04", times[i], time.Local)
		if err != nil {
			continue
		}
		hours = append(hours, HourSample{
			Time:     t,
			TempC:    at(temps, i),
			RainProb: atInt(rain, i),
			WindKph:  at(wind, i),
		})
	}
	_ = uv
	return hours
}

func buildDays(times []string, max []float64, min []float64, rain []int, wind []float64, uv []float64) []DaySample {
	n := len(times)
	days := make([]DaySample, 0, n)
	for i := 0; i < n; i++ {
		t, err := time.ParseInLocation("2006-01-02", times[i], time.Local)
		if err != nil {
			continue
		}
		days = append(days, DaySample{
			Date:     t,
			TempMaxC: at(max, i),
			TempMinC: at(min, i),
			RainProb: atInt(rain, i),
			WindKph:  at(wind, i),
			UVMax:    at(uv, i),
		})
	}
	return days
}

func nextHours(hours []HourSample, limit int) []HourSample {
	now := time.Now()
	start := 0
	for i, h := range hours {
		if !h.Time.Before(now.Truncate(time.Hour)) {
			start = i
			break
		}
	}
	end := start + limit
	if end > len(hours) {
		end = len(hours)
	}
	return hours[start:end]
}

func maxRainWindow(hours []HourSample, limit, fallback int) int {
	window := nextHours(hours, limit)
	best := 0
	for _, h := range window {
		if h.RainProb > best {
			best = h.RainProb
		}
	}
	if best == 0 {
		return fallback
	}
	return best
}

// deriveAlerts computes heat, cold, wind and rain warnings locally.
func deriveAlerts(f Forecast) []Alert {
	var alerts []Alert
	if f.TempC >= 38 || (f.FeelsLikeC > 0 && f.FeelsLikeC >= 40) {
		feels := f.TempC
		if f.FeelsLikeC != 0 {
			feels = f.FeelsLikeC
		}
		alerts = append(alerts, Alert{Kind: AlertHeat, Text: fmt.Sprintf("Heat alert: feels like %.0f°C", feels)})
	}
	if f.TempC <= 3 {
		alerts = append(alerts, Alert{Kind: AlertCold, Text: fmt.Sprintf("Cold alert: %.0f°C — bundle up", f.TempC)})
	}
	if f.WindKph >= 50 {
		alerts = append(alerts, Alert{Kind: AlertWind, Text: fmt.Sprintf("Strong wind: %.0f km/h", f.WindKph)})
	}
	if f.RainLikely {
		alerts = append(alerts, Alert{Kind: AlertRain, Text: fmt.Sprintf("Rain likely (%d%%) — close the window", f.RainProb)})
	}
	return alerts
}

func at(values []float64, i int) float64 {
	if i < len(values) {
		return values[i]
	}
	return 0
}

func atInt(values []int, i int) int {
	if i < len(values) {
		return values[i]
	}
	return 0
}
