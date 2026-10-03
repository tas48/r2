package ui

// Minimum widths at which an element's shortest variant fits.
const (
	headerMinWidth  = 9
	weatherMinWidth = 8
	statusMinWidth  = 10
	hintMinWidth    = 6
)

// bannerRows is the banner text plus a separating blank line.
const bannerRows = 2

// Layout describes which dashboard elements fit vertically and horizontally.
type Layout struct {
	Width  int
	Height int

	Header  bool
	Banner  bool
	Alert   bool
	Weather bool
	Status  bool
	Hint    bool
}

// compute fits elements top-down in priority order: header, banner, alert,
// weather, status, hint. Each element is added only if its minimum width fits
// and at least one row remains. Weather consumes as many of the remaining rows
// as it has lines. The renderer clips the final output to Height.
func compute(width, height int, hasBanner, hasAlert bool, weatherLines int) Layout {
	if width <= 0 || height <= 0 {
		return Layout{}
	}
	l := Layout{Width: width, Height: height}
	rows := 0

	add := func(minWidth int, rowsNeeded int, target *bool) {
		if rowsNeeded <= 0 {
			return
		}
		if width >= minWidth && rows+rowsNeeded <= height {
			*target = true
			rows += rowsNeeded
		}
	}

	add(headerMinWidth, 1, &l.Header)
	if hasBanner {
		add(headerMinWidth, bannerRows, &l.Banner)
	}
	if hasAlert {
		add(statusMinWidth, 1, &l.Alert)
	}
	add(weatherMinWidth, bounded(weatherLines, height-rows), &l.Weather)
	add(statusMinWidth, 1, &l.Status)
	add(hintMinWidth, 1, &l.Hint)
	return l
}

func bounded(want, available int) int {
	if want > available {
		return available
	}
	return want
}
