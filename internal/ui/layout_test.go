package ui

import "testing"

func TestComputeFitsByPriority(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		banner   bool
		wantHdr  bool
		wantBan  bool
		wantWx   bool
		wantStat bool
		wantHint bool
	}{
		{"one row only header", 20, 1, false, true, false, false, false, false},
		{"two rows header weather", 20, 2, false, true, false, true, false, false},
		{"four rows all but hint", 20, 3, false, true, false, true, true, false},
		{"four rows all fit", 20, 4, false, true, false, true, true, true},
		{"banner pushes hint out", 20, 4, true, true, true, true, false, false},
		{"banner plus all", 40, 15, true, true, true, true, true, true},
		{"narrow width drops weather", 6, 20, false, false, false, false, false, true},
		{"zero size", 0, 0, false, false, false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compute(tt.width, tt.height, tt.banner)
			if got.Header != tt.wantHdr || got.Banner != tt.wantBan ||
				got.Weather != tt.wantWx || got.Status != tt.wantStat || got.Hint != tt.wantHint {
				t.Fatalf("flags = header:%v banner:%v weather:%v status:%v hint:%v, want %v/%v/%v/%v/%v",
					got.Header, got.Banner, got.Weather, got.Status, got.Hint,
					tt.wantHdr, tt.wantBan, tt.wantWx, tt.wantStat, tt.wantHint)
			}
			if tt.width == 0 && got.Width != 0 {
				t.Fatalf("expected zero layout, got %#v", got)
			}
		})
	}
}

func TestComputeRowsNeverExceedHeight(t *testing.T) {
	for _, size := range [][2]int{{10, 3}, {100, 8}, {40, 15}, {30, 30}} {
		for _, banner := range []bool{false, true} {
			l := compute(size[0], size[1], banner)
			rows := 0
			if l.Header {
				rows++
			}
			if l.Banner {
				rows += bannerRows
			}
			for _, on := range []bool{l.Weather, l.Status, l.Hint} {
				if on {
					rows++
				}
			}
			if rows > size[1] {
				t.Fatalf("size %dx%d banner=%v: %d rows exceeds height", size[0], size[1], banner, rows)
			}
		}
	}
}
