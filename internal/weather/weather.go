package weather

import (
	"context"
	"time"
)

// Alert is a severe-weather warning derived from the forecast (heat, cold,
// wind, heavy rain). These are computed locally, not from Open-Meteo.
type Alert struct {
	Kind AlertKind
	Text string
}

// AlertKind identifies a computed alert.
type AlertKind uint8

const (
	AlertHeat AlertKind = iota
	AlertCold
	AlertWind
	AlertRain
)

// HourSample is one hour of the forecast.
type HourSample struct {
	Time     time.Time
	TempC    float64
	RainProb int
	WindKph  float64
}

// DaySample is one day of the forecast.
type DaySample struct {
	Date     time.Time
	TempMinC float64
	TempMaxC float64
	RainProb int
	WindKph  float64
	UVMax    float64
}

// Provider fetches a forecast for a city or for explicit coordinates.
type Provider interface {
	Forecast(ctx context.Context, city string) (Forecast, error)
	ForecastAt(ctx context.Context, name string, lat, lon float64) (Forecast, error)
}

// Forecast is a rich weather summary for a location.
type Forecast struct {
	Location string

	RainProb   int
	TempC      float64
	FeelsLikeC float64
	WindKph    float64
	Humidity   int
	UVIndex    float64
	RainLikely bool
	IsDay      bool

	NextHours []HourSample
	Today     []HourSample
	Daily     []DaySample
	Alerts    []Alert

	FetchedAt time.Time
}
