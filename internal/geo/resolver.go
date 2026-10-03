package geo

import (
	"context"
	"fmt"
	"log/slog"
)

// Resolver tries locators in order of precision (Windows location, then IP) and
// returns the first success. Failures are logged and skipped so a denied
// permission never blocks startup.
type Resolver struct {
	locators []Locator
	logger   *slog.Logger
}

func NewResolver(logger *slog.Logger, locators ...Locator) *Resolver {
	if logger == nil {
		logger = slog.Default()
	}
	return &Resolver{locators: locators, logger: logger}
}

// DefaultResolver tries Windows location first, then IP geolocation.
func DefaultResolver(logger *slog.Logger) *Resolver {
	return NewResolver(logger, NewWindowsLocator(0), NewIPLocator(nil))
}

func (r *Resolver) Locate(ctx context.Context) (Location, error) {
	var lastErr error
	for _, locator := range r.locators {
		location, err := locator.Locate(ctx)
		if err != nil {
			lastErr = err
			r.logger.Debug("locator failed", "error", err)
			continue
		}
		r.logger.Info("location resolved", "source", location.Source, "city", location.City)
		return location, nil
	}
	return Location{}, fmt.Errorf("no locator succeeded: %w", lastErr)
}
