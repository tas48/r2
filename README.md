# r2

A terminal-based companion that keeps you informed, focused, and taking care of
yourself throughout the day.

r2 is a small, rule-based agent that lives in a dedicated terminal window. It
reminds you to drink water, nudges you to take breaks, and shows a live weather
dashboard for your location — with rain chances for the coming hours, the next
days, wind, humidity, UV and heat/cold/rain alerts. No AI, no cloud account, no
magic: just explicit rules, a few JSON files, and a terminal-native interface
that stays readable from a 10x3 window to a 120x40 one.

> **Status: work in progress.** The agent core, persistence, precise location
> resolution, the weather provider and the responsive dashboard are in place.
> See the [Roadmap](#roadmap).

## Why

Most reminder tools assume a full GUI or send you away from your work. r2 does
the opposite: it stays where you already are — the terminal — showing the
information you actually want (weather, water, breaks) as a compact, icon-led
dashboard that adapts to the window you have.

## Features

- **Water reminders** on a configurable interval, with a daily goal and bottle
  size. Log a bottle and the timer resets.
- **Break reminders** so you actually get up from the chair.
- **Daily counters** for water and breaks, resetting automatically at midnight.
- **Persistent state** — progress survives restarts.
- **Weather dashboard**: current conditions, next hours, the following days,
  wind, humidity and UV, with icons for sun, clouds, rain, wind, cold and night.
- **Local alerts**: heat, cold, strong wind and likely rain.
- **Automatic location**: resolved from the Windows Location API (most precise),
  falling back to IP geolocation, then refined with reverse geocoding to the
  city or neighbourhood.
- **Responsive layout** that drops secondary elements before the essentials.

## Keys

| Key          | Action                    |
| ------------ | ------------------------- |
| `w`          | Log a bottle of water     |
| `b`          | Log a break               |
| `q`/`ctrl+c` | Quit                      |

## Install

Requires **Go 1.26+**.

```bash
go build -o r2 ./cmd/r2
```

Or run it directly:

```bash
go run ./cmd/r2
```

## Configuration

On first run r2 writes a default config to your user config directory:

- Linux/macOS: `$XDG_CONFIG_HOME/r2/config.json` (usually `~/.config/r2`)
- Windows: `%AppData%\r2\config.json`

```json
{
  "city": "Sao Paulo",
  "water_goal_ml": 2000,
  "bottle_ml": 500,
  "water_every": "30m",
  "break_every": "45m",
  "weather_every": "15m",
  "rain_threshold": 50,
  "rain_horizon_hours": 6
}
```

Durations use Go syntax (`"30m"`, `"1h30m"`). Missing fields fall back to the
defaults, so a partial file is fine. `city` is only used when automatic
location resolution fails; `rain_threshold` is the chance of rain (percent)
that triggers the "close the window" alert.

## State

Daily progress is stored next to the config in `state.json`. When the stored
date does not match today, counters reset on load. Writes are atomic.

## Architecture

r2 keeps its domains independent and its IO at the edges:

```text
cmd/r2/            entrypoint
internal/agent/    scheduler, trackers, reminders, events (pure, no IO)
internal/config/   JSON configuration and defaults
internal/store/    daily state persistence and reset
internal/geo/      location resolution: Windows API, IP geolocation, resolver
internal/weather/  weather provider, Open-Meteo and reverse geocoding
internal/ui/       Bubble Tea model, layout, rendering, theme
internal/app/      wiring: config -> store -> agent -> geo -> weather -> ui
```

The **agent never performs IO**. When it needs something from the outside (like
a weather forecast), it emits an event; the UI turns that event into a Bubble Tea
command. Network calls never run inside the update loop.

## Locations and weather

On startup r2 resolves your location, most precise first:

1. **Windows Location API** (`Windows.Devices.Geolocation`), which uses
   GPS/Wi-Fi/cell — the most accurate option, if location access is allowed.
2. **IP geolocation** (keyless, city-level): `ipwho.is`, then `ipinfo.io`, then
   `freeipapi.com`.
3. **Configured city** as the final fallback.

The coordinates are then reverse-geocoded (BigDataCloud, keyless) to a readable
city or neighbourhood name. Weather comes from Open-Meteo, also keyless, and
includes current conditions, the next hours, the following days, wind, humidity,
UV and locally-derived heat/cold/wind/rain alerts.

## Roadmap

- [x] **M0** — Bubble Tea v2 alt-screen skeleton, tick loop, quit.
- [x] **M1** — rule-based agent: scheduler, water/break trackers, events.
- [x] **M2** — JSON config, daily state persistence, app wiring.
- [x] **M3** — responsive layout by priority, theme.
- [x] **M4** — precise location (Windows API + IP) and reverse geocoding.
- [x] **M5** — weather dashboard: current, hourly, daily, wind, humidity, UV,
      alerts.
- [ ] **M6** — polish, keyboard help overlay, CPU checks.

## Development

```bash
gofmt -l .
go vet ./...
go test ./...
go build ./...
```

For live reload, install [air](https://github.com/air-verse/air) and run it:

```bash
go install github.com/air-verse/air@latest
air
```

Air is intentionally not a module dependency. Configuration lives in
`.air.toml`.

See [AGENTS.md](AGENTS.md) for architecture rules, code style and commit
conventions.

## License

See [LICENSE](LICENSE).
