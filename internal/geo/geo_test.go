package geo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseLoc(t *testing.T) {
	lat, lon, err := parseLoc("-19.3111,-46.0489")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if lat != -19.3111 || lon != -46.0489 {
		t.Fatalf("unexpected coords: %v, %v", lat, lon)
	}
}

func TestParseLocInvalid(t *testing.T) {
	if _, _, err := parseLoc("nope"); err == nil {
		t.Fatal("expected an error")
	}
}

func TestIPLocatorParsesSuperset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ipwho.is style: nested latitude/longitude + city.
		w.Write([]byte(`{"city":"Rio Paranaiba","region":"Minas Gerais","country":"Brazil","latitude":-19.19,"longitude":-46.24}`))
	}))
	defer server.Close()

	locator := &IPLocator{client: server.Client()}
	location, err := locator.locateFrom(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("locate failed: %v", err)
	}
	if location.City != "Rio Paranaiba" {
		t.Fatalf("unexpected city %q", location.City)
	}
	if location.Latitude != -19.19 || location.Longitude != -46.24 {
		t.Fatalf("unexpected coords: %#v", location)
	}
}

func TestIPLocatorParsesIpinfoLoc(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"city":"Sao Gotardo","region":"Minas Gerais","country":"BR","loc":"-19.3111,-46.0489"}`))
	}))
	defer server.Close()

	locator := &IPLocator{client: server.Client()}
	location, err := locator.locateFrom(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("locate failed: %v", err)
	}
	if location.Latitude != -19.3111 || location.Longitude != -46.0489 {
		t.Fatalf("unexpected coords: %#v", location)
	}
}

func TestResolverFallsBackToSecondLocator(t *testing.T) {
	failing := fakeLocator{err: context.DeadlineExceeded}
	working := fakeLocator{location: Location{City: "Fallback", Source: "fake"}}
	resolver := NewResolver(nil, failing, working)
	location, err := resolver.Locate(context.Background())
	if err != nil {
		t.Fatalf("expected fallback to succeed: %v", err)
	}
	if location.City != "Fallback" {
		t.Fatalf("unexpected location %#v", location)
	}
}

type fakeLocator struct {
	location Location
	err      error
}

func (f fakeLocator) Locate(context.Context) (Location, error) {
	return f.location, f.err
}
