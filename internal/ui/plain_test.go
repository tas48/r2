package ui

import "regexp"

var sgrPattern = regexp.MustCompile("\x1b\\[[0-9;]*m")

// plain removes SGR color sequences so golden files capture layout, not color.
func plain(s string) string {
	return sgrPattern.ReplaceAllString(s, "")
}
