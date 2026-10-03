package weather

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestServer serves geocoding, forecast and reverse-geocoding responses.
func newTestServer(t *testing.T, forecastBody string) *httptest.Server {
	t.Helper()
	var geocodeCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/search"):
			geocodeCalls++
			w.Write([]byte(`{"results":[{"name":"Recife","latitude":-8.05,"longitude":-34.9}]}`))
		case strings.HasPrefix(r.URL.Path, "/reverse"):
			w.Write([]byte(`{"city":"Boa Viagem","locality":"Boa Viagem"}`))
		case strings.HasPrefix(r.URL.Path, "/forecast"):
			w.Write([]byte(forecastBody))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	_ = geocodeCalls
	return server
}

func pastHourTimes() string {
	return `["2020-01-01T00:00","2020-01-01T01:00","2020-01-01T02:00"]`
}

func TestForecastDerivesRainAndAlerts(t *testing.T) {
	body := `{
		"current":{"temperature_2m":39,"apparent_temperature":41,"relative_humidity_2m":55,
			"wind_speed_10m":60,"uv_index":9,"is_day":1,"precipitation_probability":10},
		"hourly":{"time":` + pastHourTimes() + `,
			"temperature_2m":[30,31,32],"precipitation_probability":[10,70,20],
			"wind_speed_10m":[10,12,14],"uv_index":[5,6,7]},
		"daily":{"time":["2020-01-01"],"temperature_2m_max":[35],"temperature_2m_min":[25],
			"precipitation_probability_max":[70],"wind_speed_10m_max":[40],"uv_index_max":[9]}
	}`
	server := newTestServer(t, body)
	provider := NewOpenMeteo(nil, 50, 6)
	provider.geocodeBase = server.URL + "/search"
	provider.forecastBase = server.URL + "/forecast"

	forecast, err := provider.Forecast(context.Background(), "Recife")
	if err != nil {
		t.Fatalf("forecast failed: %v", err)
	}
	if forecast.Location != "Recife" {
		t.Fatalf("expected location Recife, got %q", forecast.Location)
	}
	if forecast.TempC != 39 || forecast.FeelsLikeC != 41 {
		t.Fatalf("unexpected temps: %#v", forecast)
	}
	if !forecast.RainLikely {
		t.Fatal("expected rain likely from the 70% hour within the horizon")
	}
	if len(forecast.Daily) != 1 || forecast.Daily[0].UVMax != 9 {
		t.Fatalf("expected one daily sample with uv 9, got %#v", forecast.Daily)
	}
	// 39 C, 60 km/h wind and rain must all produce alerts.
	kinds := map[AlertKind]bool{}
	for _, a := range forecast.Alerts {
		kinds[a.Kind] = true
	}
	if !kinds[AlertHeat] || !kinds[AlertWind] || !kinds[AlertRain] {
		t.Fatalf("expected heat, wind and rain alerts, got %#v", forecast.Alerts)
	}
}

func TestForecastNoFalseColdAlert(t *testing.T) {
	body := `{
		"current":{"temperature_2m":20,"apparent_temperature":19,"relative_humidity_2m":60,
			"wind_speed_10m":8,"uv_index":3,"is_day":1,"precipitation_probability":5},
		"hourly":{"time":` + pastHourTimes() + `,
			"temperature_2m":[19,20,21],"precipitation_probability":[5,10,15],
			"wind_speed_10m":[8,8,8],"uv_index":[3,3,3]},
		"daily":{"time":["2020-01-01"],"temperature_2m_max":[22],"temperature_2m_min":[15],
			"precipitation_probability_max":[15],"wind_speed_10m_max":[8],"uv_index_max":[3]}
	}`
	server := newTestServer(t, body)
	provider := NewOpenMeteo(nil, 50, 6)
	provider.geocodeBase = server.URL + "/search"
	provider.forecastBase = server.URL + "/forecast"

	forecast, err := provider.Forecast(context.Background(), "Recife")
	if err != nil {
		t.Fatalf("forecast failed: %v", err)
	}
	if len(forecast.Alerts) != 0 {
		t.Fatalf("expected no alerts at 20 C, got %#v", forecast.Alerts)
	}
}

func TestReverseGeocoderName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"city":"Boa Viagem","locality":"Pina"}`))
	}))
	defer server.Close()

	geocoder := NewReverseGeocoder(server.Client())
	geocoder.base = server.URL
	name, err := geocoder.Name(context.Background(), -8.05, -34.9)
	if err != nil {
		t.Fatalf("reverse geocoding failed: %v", err)
	}
	if name != "Pina" {
		t.Fatalf("expected locality Pina, got %q", name)
	}
}

func TestForecastAtSkipsGeocoding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/search") {
			t.Fatal("ForecastAt must not geocode")
		}
		w.Write([]byte(`{"current":{"temperature_2m":25,"apparent_temperature":25,"relative_humidity_2m":50,"wind_speed_10m":5,"uv_index":2,"is_day":1,"precipitation_probability":0},"hourly":{"time":` + pastHourTimes() + `,"temperature_2m":[24,25,26],"precipitation_probability":[0,0,0],"wind_speed_10m":[5,5,5],"uv_index":[2,2,2]},"daily":{"time":["2020-01-01"],"temperature_2m_max":[27],"temperature_2m_min":[18],"precipitation_probability_max":[0],"wind_speed_10m_max":[5],"uv_index_max":[2]}}`))
	}))
	defer server.Close()

	provider := NewOpenMeteo(nil, 50, 6)
	provider.forecastBase = server.URL
	forecast, err := provider.ForecastAt(context.Background(), "Somewhere", -8, -34)
	if err != nil {
		t.Fatalf("ForecastAt failed: %v", err)
	}
	if forecast.Location != "Somewhere" {
		t.Fatalf("expected location name Somewhere, got %q", forecast.Location)
	}
}

func TestGeocodeNoResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"results":[]}`))
	}))
	defer server.Close()

	provider := NewOpenMeteo(nil, 50, 6)
	provider.geocodeBase = server.URL
	if _, err := provider.Forecast(context.Background(), "Atlantis"); err == nil {
		t.Fatal("expected an error when geocoding returns no results")
	}
}
