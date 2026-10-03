package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultReverseBase = "https://api.bigdatacloud.net/data/reverse-geocode-client"

// ReverseGeocoder turns coordinates into a human-readable place name, refining
// an IP-only result towards the neighbourhood/localidade when possible.
type ReverseGeocoder struct {
	client *http.Client
	base   string
}

func NewReverseGeocoder(client *http.Client) *ReverseGeocoder {
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	return &ReverseGeocoder{client: client, base: defaultReverseBase}
}

// Name returns the most specific name available (locality, then city), or an
// empty string when nothing useful is found.
func (g *ReverseGeocoder) Name(ctx context.Context, lat, lon float64) (string, error) {
	query := url.Values{
		"latitude":         {fmt.Sprintf("%.4f", lat)},
		"longitude":        {fmt.Sprintf("%.4f", lon)},
		"localityLanguage": {"pt"},
	}
	var raw struct {
		City     string `json:"city"`
		Locality string `json:"locality"`
	}
	endpoint := g.base + "?" + query.Encode()
	if err := g.getJSON(ctx, endpoint, &raw); err != nil {
		return "", fmt.Errorf("reverse geocoding %.4f,%.4f: %w", lat, lon, err)
	}
	return firstNonEmptyString(raw.Locality, raw.City), nil
}

func (g *ReverseGeocoder) getJSON(ctx context.Context, endpoint string, dst any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	response, err := g.client.Do(request)
	if err != nil {
		return fmt.Errorf("requesting %s: %w", endpoint, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from %s", response.StatusCode, endpoint)
	}
	return decodeJSON(response.Body, dst)
}

func decodeJSON(reader io.Reader, dst any) error {
	return json.NewDecoder(reader).Decode(dst)
}

func firstNonEmptyString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
