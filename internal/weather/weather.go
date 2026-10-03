package weather

import (
	"context"
	"time"
)

// Forecast is a weather summary for a location.
type Forecast struct {
	RainProb   int
	TempC      float64
	RainLikely bool
	FetchedAt  time.Time
}

// Provider fetches a forecast for a city.
type Provider interface {
	Forecast(ctx context.Context, city string) (Forecast, error)
}
