# r2

A terminal-based companion that keeps you informed, focused, and taking care of
yourself throughout the day.

r2 is a small, rule-based agent that lives in a dedicated terminal window. It
reminds you to drink water, nudges you to take breaks, checks the weather for
your city, and tracks your progress — all presented through a small
ASCII/Unicode pet with a bit of personality.

No AI, no cloud account, no magic. Just explicit rules, a few JSON files, and a
terminal-native interface that stays responsive from a 10x3 window to a 120x40
one.

> **Status: work in progress.** The agent core, configuration, persistence and
> the responsive layout are in place. The pet and the weather integration are
> next; see the [Roadmap](#roadmap).

## Why

Most reminder tools assume a full GUI or send you away from your work. r2 does
the opposite: it stays where you already are — the terminal — and gives you a
small creature to take care of that quietly takes care of you.

## Features

- **Water reminders** on a configurable interval, with a daily goal and bottle
  size. Log a bottle and the timer resets.
- **Break reminders** so you actually get up from the chair.
- **Daily counters** for water and breaks, resetting automatically at midnight.
- **Persistent state** — progress survives restarts.
- **Configurable** through a small JSON file; everything else is hardcoded.
- **Responsive layout** that adapts to tiny and large terminals, dropping
  secondary elements before the essentials.
- **A pet** that reacts to what you do (in progress).

## Keys

| Key         | Action                    |
| ----------- | ------------------------- |
| `w`         | Log a bottle of water     |
| `b`         | Log a break               |
| `space`     | Interact with the pet     |
| `?`         | Toggle help               |
| `q`/`ctrl+c`| Quit                      |

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
  "break_every": "45m"
}
```

Durations use Go syntax (`"30m"`, `"1h30m"`). Missing fields fall back to the
defaults, so a partial file is fine.

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
internal/pet/      pet state machine, animation, movement (pure)
internal/weather/  weather provider interface and Open-Meteo implementation
internal/ui/       Bubble Tea model, layout, rendering, keymap, theme
internal/app/      wiring: config -> store -> agent -> pet -> ui
```

The **agent never performs IO**. When it needs something from the outside
(like a weather forecast), it emits an event; the UI turns that event into a
Bubble Tea command. The **pet** never knows about the agent or the UI — it only
reacts to events.

## Roadmap

- [x] **M0** — Bubble Tea v2 alt-screen skeleton, tick loop, quit.
- [x] **M1** — rule-based agent: scheduler, water/break trackers, events.
- [x] **M2** — JSON config, daily state persistence, app wiring.
- [x] **M3** — responsive layout by priority, theme, help overlay.
- [ ] **M4** — pet: states, transition table, animation, movement, reactions.
- [ ] **M5** — weather: Open-Meteo geocoding and forecast, async fetch, rain
      alert.
- [ ] **M6** — polish, CPU checks, layout golden tests.

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
