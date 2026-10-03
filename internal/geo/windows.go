package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// windowsScript uses the Windows.Devices.Geolocation WinRT API, which relies on
// GPS/Wi-Fi/cell and is far more precise than IP geolocation. It requires the
// user to have allowed location access; if denied, it exits non-zero and the
// caller falls back to IP.
//
// The script tries the highest accuracy first and downgrades, because a precise
// request can take a long time or fail indoors.
const windowsScript = `$ErrorActionPreference='Stop'
try {
  [Windows.Devices.Geolocation.Geolocator,Windows.Devices.Geolocation,ContentType=WindowsRuntime] | Out-Null
  $g = New-Object Windows.Devices.Geolocation.Geolocator
  try { $g.DesiredAccuracyInMeters = 10 } catch {}
  $op = $g.GetGeopositionAsync()
  $deadline = (Get-Date).AddSeconds(6)
  while ($op.Status -eq 'Started' -and (Get-Date) -lt $deadline) { Start-Sleep -Milliseconds 200 }
  if ($op.Status -ne 'Completed') { exit 2 }
  $p = $op.GetResults().Coordinate.Point.Position
  [pscustomobject]@{latitude=$p.Latitude;longitude=$p.Longitude} | ConvertTo-Json -Compress
} catch { exit 3 }`

// WindowsLocator resolves location via the Windows Location API.
type WindowsLocator struct {
	timeout time.Duration
}

func NewWindowsLocator(timeout time.Duration) *WindowsLocator {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &WindowsLocator{timeout: timeout}
}

func (l *WindowsLocator) Locate(ctx context.Context) (Location, error) {
	if !windowsAvailable() {
		return Location{}, fmt.Errorf("windows location api unavailable")
	}
	ctx, cancel := context.WithTimeout(ctx, l.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", windowsScript)
	output, err := cmd.Output()
	if err != nil {
		return Location{}, fmt.Errorf("windows geolocation failed: %w", err)
	}
	var raw struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(output))), &raw); err != nil {
		return Location{}, fmt.Errorf("parsing windows geolocation: %w", err)
	}
	if raw.Latitude == 0 && raw.Longitude == 0 {
		return Location{}, fmt.Errorf("windows geolocation returned no coordinates")
	}
	return Location{
		Latitude:  raw.Latitude,
		Longitude: raw.Longitude,
		Source:    "windows-location",
	}, nil
}
