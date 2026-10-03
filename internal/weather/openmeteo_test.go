package weather

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestForecastFindsRainWithinHorizon(t *testing.T) {
	var geocodeCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/search"):
			geocodeCalls++
			w.Write([]byte(`{"results":[{"latitude":-8.05,"longitude":-34.9}]}`))
		case strings.HasPrefix(r.URL.Path, "/forecast"):
			w.Write([]byte(`{"current":{"temperature_2m":28.5},"hourly":{"precipitation_probability":[10,70,20]}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider := NewOpenMeteo(nil, 50, 6)
	provider.geocodeBase = server.URL + "/search"
	provider.forecastBase = server.URL + "/forecast"

	forecast, err := provider.Forecast(context.Background(), "Recife")
	if err != nil {
		t.Fatalf("forecast failed: %v", err)
	}
	if forecast.RainProb != 70 {
		t.Fatalf("expected rain probability 70, got %d", forecast.RainProb)
	}
	if !forecast.RainLikely {
		t.Fatal("expected rain to be likely at 70%% over a 50%% threshold")
	}
	if forecast.TempC != 28.5 {
		t.Fatalf("expected 28.5 C, got %v", forecast.TempC)
	}

	if _, err := provider.Forecast(context.Background(), "Recife"); err != nil {
		t.Fatalf("second forecast failed: %v", err)
	}
	if geocodeCalls != 1 {
		t.Fatalf("expected geocoding to be cached, got %d calls", geocodeCalls)
	}
}

func TestForecastBelowThreshold(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/search") {
			w.Write([]byte(`{"results":[{"latitude":0,"longitude":0}]}`))
			return
		}
		w.Write([]byte(`{"current":{"temperature_2m":20},"hourly":{"precipitation_probability":[5,10,15]}}`))
	}))
	defer server.Close()

	provider := NewOpenMeteo(nil, 50, 6)
	provider.geocodeBase = server.URL + "/search"
	provider.forecastBase = server.URL + "/forecast"

	forecast, err := provider.Forecast(context.Background(), "Nowhere")
	if err != nil {
		t.Fatalf("forecast failed: %v", err)
	}
	if forecast.RainLikely || forecast.RainProb != 15 {
		t.Fatalf("expected no rain, got %#v", forecast)
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
