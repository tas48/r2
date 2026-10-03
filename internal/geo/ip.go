package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Location is a resolved place, optionally with a street-level detail.
type Location struct {
	City      string
	Region    string
	Country   string
	Latitude  float64
	Longitude float64
	Source    string
}

// Locator resolves the user's approximate location. Implementations must not
// require an API key and must be safe to call once per session.
type Locator interface {
	Locate(ctx context.Context) (Location, error)
}

// IP locator endpoints, tried in order. All are keyless and return JSON with
// city + coordinates.
var ipEndpoints = []string{
	"https://ipwho.is/",
	"https://ipinfo.io/json",
	"https://freeipapi.com/api/json",
}

// IPLocator resolves the location from the public IP address. Accuracy is
// city-level (sometimes better), which is the best a keyless service offers.
type IPLocator struct {
	client *http.Client
}

func NewIPLocator(client *http.Client) *IPLocator {
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	return &IPLocator{client: client}
}

func (l *IPLocator) Locate(ctx context.Context) (Location, error) {
	var lastErr error
	for _, endpoint := range ipEndpoints {
		location, err := l.locateFrom(ctx, endpoint)
		if err != nil {
			lastErr = err
			continue
		}
		return location, nil
	}
	return Location{}, fmt.Errorf("all ip locators failed: %w", lastErr)
}

func (l *IPLocator) locateFrom(ctx context.Context, endpoint string) (Location, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Location{}, fmt.Errorf("building request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	response, err := l.client.Do(request)
	if err != nil {
		return Location{}, fmt.Errorf("requesting %s: %w", endpoint, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Location{}, fmt.Errorf("unexpected status %d from %s", response.StatusCode, endpoint)
	}

	// Decode into a superset: ipwho.is nests coordinates and city; ipinfo
	// provides "loc":"lat,lon"; freeipapi uses latitude/longitude.
	var raw struct {
		City        string  `json:"city"`
		CityName    string  `json:"cityName"`
		Region      string  `json:"region"`
		RegionName  string  `json:"regionName"`
		Country     string  `json:"country"`
		CountryName string  `json:"countryName"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		Loc         string  `json:"loc"`
	}
	if err := json.NewDecoder(response.Body).Decode(&raw); err != nil {
		return Location{}, fmt.Errorf("decoding response from %s: %w", endpoint, err)
	}

	location := Location{
		City:    firstNonEmpty(raw.City, raw.CityName),
		Region:  firstNonEmpty(raw.Region, raw.RegionName),
		Country: firstNonEmpty(raw.Country, raw.CountryName),
		Source:  endpoint,
	}
	switch {
	case raw.Latitude != 0 && raw.Longitude != 0:
		location.Latitude, location.Longitude = raw.Latitude, raw.Longitude
	case raw.Loc != "":
		lat, lon, err := parseLoc(raw.Loc)
		if err != nil {
			return Location{}, fmt.Errorf("parsing loc from %s: %w", endpoint, err)
		}
		location.Latitude, location.Longitude = lat, lon
	}
	if location.City == "" || (location.Latitude == 0 && location.Longitude == 0) {
		return Location{}, fmt.Errorf("incomplete location from %s", endpoint)
	}
	return location, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
