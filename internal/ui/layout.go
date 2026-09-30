package ui

// Minimum widths at which an element's shortest variant fits.
const (
	headerMinWidth = 9
	statusMinWidth = 10
	hintMinWidth   = 6
)

// stageMinRows is the smallest pet area worth reserving.
const stageMinRows = 3

// bannerRows is the banner text plus a separating blank line.
const bannerRows = 2

// Rect is a rectangular area of the screen.
type Rect struct {
	X, Y, W, H int
}

// Layout describes which elements fit and where the pet stage is.
type Layout struct {
	Width  int
	Height int

	Header bool
	Banner bool
	Status bool
	Hint   bool

	Stage Rect
}

// compute fits elements by priority: the banner is always kept, the pet stage
// reserves stageMinRows, and status, clock and hint are dropped as space runs
// out (hint first, then clock, then status).
func compute(width, height int, hasBanner bool) Layout {
	if width <= 0 || height <= 0 {
		return Layout{}
	}
	l := Layout{Width: width, Height: height}

	rows := 0
	if hasBanner {
		l.Banner = true
		rows += bannerRows
	}
	if width >= statusMinWidth && height-rows-1 >= stageMinRows {
		l.Status = true
		rows++
	}
	if width >= headerMinWidth && height-rows-1 >= stageMinRows {
		l.Header = true
		rows++
	}
	if width >= hintMinWidth && height-rows-1 >= stageMinRows {
		l.Hint = true
		rows++
	}

	top := 0
	if l.Header {
		top++
	}
	if l.Banner {
		top += bannerRows
	}
	stageRows := height - rows
	if stageRows < 0 {
		stageRows = 0
	}
	l.Stage = Rect{X: 0, Y: top, W: width, H: stageRows}
	return l
}
