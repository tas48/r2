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
	Weather bool
	Status  bool
	Hint    bool
}

// compute fits elements top-down with a fixed priority: header, banner,
// weather, status, hint. Elements are dropped from the bottom (hint first)
// when the terminal is too short. The banner is reserved before the weather.
func compute(width, height int, hasBanner bool) Layout {
	if width <= 0 || height <= 0 {
		return Layout{}
	}
	l := Layout{Width: width, Height: height}

	rows := 0
	if width >= headerMinWidth && rows+1 <= height {
		l.Header = true
		rows++
	}
	if hasBanner && rows+bannerRows <= height {
		l.Banner = true
		rows += bannerRows
	}
	if width >= weatherMinWidth && rows+1 <= height {
		l.Weather = true
		rows++
	}
	if width >= statusMinWidth && rows+1 <= height {
		l.Status = true
		rows++
	}
	if width >= hintMinWidth && rows+1 <= height {
		l.Hint = true
		rows++
	}
	return l
}
