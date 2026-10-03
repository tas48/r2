package ui

import (
	"strings"
	"testing"

	"github.com/tas48/r2/internal/weather"
)

func TestComputeFitsByPriority(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		banner   bool
		alert    bool
		wxLines  int
		wantHdr  bool
		wantBan  bool
		wantAl   bool
		wantWx   bool
		wantStat bool
		wantHint bool
	}{
		// width 20, no banner/alert. Weather greedy consumes the rest.
		{"header only", 20, 1, false, false, 3, true, false, false, false, false, false},
		{"header + 1 weather line", 20, 2, false, false, 3, true, false, false, true, false, false},
		{"header + 2 weather lines", 20, 3, false, false, 3, true, false, false, true, false, false},
		{"weather fills all", 20, 3, false, false, 9, true, false, false, true, false, false},
		{"no weather leaves room for status", 20, 3, false, false, 0, true, false, false, false, true, true},
		{"banner then weather", 20, 4, true, false, 1, true, true, false, true, false, false},
		{"alert then weather and status", 20, 4, false, true, 1, true, false, true, true, true, false},
		{"narrow drops weather", 6, 20, false, false, 3, false, false, false, false, false, true},
		{"zero size", 0, 0, false, false, 0, false, false, false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compute(tt.width, tt.height, tt.banner, tt.alert, tt.wxLines)
			if got.Header != tt.wantHdr || got.Banner != tt.wantBan || got.Alert != tt.wantAl ||
				got.Weather != tt.wantWx || got.Status != tt.wantStat || got.Hint != tt.wantHint {
				t.Fatalf("flags = header:%v banner:%v alert:%v weather:%v status:%v hint:%v, want %v/%v/%v/%v/%v/%v",
					got.Header, got.Banner, got.Alert, got.Weather, got.Status, got.Hint,
					tt.wantHdr, tt.wantBan, tt.wantAl, tt.wantWx, tt.wantStat, tt.wantHint)
			}
		})
	}
}

func TestRenderNeverExceedsHeight(t *testing.T) {
	for _, size := range [][2]int{{10, 3}, {100, 8}, {40, 15}, {30, 30}, {20, 2}, {6, 20}} {
		for _, banner := range []bool{false, true} {
			for _, alert := range []bool{false, true} {
				m := sizedModel(size[0], size[1])
				if banner {
					m.banner = "Time to drink water"
				}
				if alert {
					m.forecast.Alerts = append(m.forecast.Alerts, weather.Alert{Kind: weather.AlertRain, Text: "Rain likely — close the window"})
				}
				rendered := m.render()
				if rendered == "" {
					continue
				}
				if got := len(strings.Split(rendered, "\n")); got > size[1] {
					t.Fatalf("size %dx%d banner=%v alert=%v: rendered %d lines", size[0], size[1], banner, alert, got)
				}
			}
		}
	}
}

func TestComputeWeatherClampedToHeight(t *testing.T) {
	l := compute(80, 4, false, false, 4)
	if !l.Weather {
		t.Fatal("expected weather to be shown")
	}
	if l.Status || l.Hint {
		t.Fatal("expected weather lines to consume the remaining height")
	}
}
