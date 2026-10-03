//go:build windows

package geo

import "runtime"

func windowsAvailable() bool {
	return runtime.GOOS == "windows"
}
