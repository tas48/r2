package geo

import (
	"fmt"
	"strconv"
	"strings"
)

// parseLoc parses the "lat,lon" form used by ipinfo.io.
func parseLoc(loc string) (float64, float64, error) {
	parts := strings.Split(loc, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected lat,lon, got %q", loc)
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parsing latitude %q: %w", parts[0], err)
	}
	lon, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parsing longitude %q: %w", parts[1], err)
	}
	return lat, lon, nil
}
