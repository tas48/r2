package ui

import "testing"

func TestComputeFitsByPriority(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		height    int
		banner    bool
		wantHdr   bool
		wantBan   bool
		wantStat  bool
		wantHint  bool
		wantStage int
		wantTop   int
	}{
		{"tiny width only stage", 10, 3, false, false, false, false, false, 3, 0},
		{"tiny with banner keeps banner", 10, 3, true, false, true, false, false, 1, 2},
		{"banner taller than screen", 10, 2, true, false, true, false, false, 0, 2},
		{"narrow drops text", 8, 20, false, false, false, false, true, 19, 0},
		{"small all fit", 20, 8, false, true, false, true, true, 5, 1},
		{"small header and status only", 20, 5, false, true, false, true, false, 3, 1},
		{"very short keeps status and stage", 20, 4, false, false, false, true, false, 3, 0},
		{"too short for status", 20, 2, false, false, false, false, false, 2, 0},
		{"compact all fit", 30, 10, false, true, false, true, true, 7, 1},
		{"wide short keeps all", 100, 8, false, true, false, true, true, 5, 1},
		{"medium", 40, 15, false, true, false, true, true, 12, 1},
		{"medium with banner", 40, 15, true, true, true, true, true, 10, 3},
		{"tall", 30, 30, false, true, false, true, true, 27, 1},
		{"large", 80, 24, false, true, false, true, true, 21, 1},
		{"huge", 120, 40, false, true, false, true, true, 37, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compute(tt.width, tt.height, tt.banner)
			if got.Header != tt.wantHdr || got.Banner != tt.wantBan ||
				got.Status != tt.wantStat || got.Hint != tt.wantHint {
				t.Fatalf("flags = header:%v banner:%v status:%v hint:%v, want header:%v banner:%v status:%v hint:%v",
					got.Header, got.Banner, got.Status, got.Hint,
					tt.wantHdr, tt.wantBan, tt.wantStat, tt.wantHint)
			}
			if got.Stage.H != tt.wantStage {
				t.Fatalf("stage height = %d, want %d", got.Stage.H, tt.wantStage)
			}
			if got.Stage.Y != tt.wantTop {
				t.Fatalf("stage y = %d, want %d", got.Stage.Y, tt.wantTop)
			}
		})
	}
}

func TestComputeIgnoresEmptySize(t *testing.T) {
	if got := compute(0, 0, false); got.Width != 0 {
		t.Fatalf("expected zero layout, got %#v", got)
	}
}
