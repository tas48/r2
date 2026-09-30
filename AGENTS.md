# AGENTS.md

## Project

r2 is a terminal companion: a rule-based (non-AI) terminal agent that reminds
you to drink water and take breaks, checks the weather, and tracks your
progress, presented through a small ASCII/Unicode pet.

Written in Go, built on Bubble Tea v2. Designed as independent domains
(agent, pet) behind a thin TUI.

## Core principles

- Keep the agent domain independent from any UI framework and from IO.
- Keep the pet domain independent from the agent and the UI; it only reacts to events.
- No AI, and no network inside the UI update loop.
- Prefer simple, explicit, maintainable Go code.
- Do not add unnecessary abstractions or dependencies.
- Persist state as small JSON files; the agent stays pure.
- Never expose credentials or sensitive information in logs, errors, tests, or docs.

## Architecture

```text
cmd/r2/
    Application entrypoint

internal/agent/
    Scheduler, reminders, trackers, domain events (pure, no IO)

internal/weather/
    Weather provider interface and Open-Meteo implementation

internal/config/
    Configuration loading and defaults

internal/store/
    State persistence and daily reset

internal/pet/
    Pet state machine, animation, movement, sprites (pure)

internal/ui/
    Bubble Tea model, layout, rendering, keymap, theme

internal/app/
    Wiring: config -> store -> agent -> pet -> ui
```

## Dependency direction

```text
main -> app -> {config, store, agent, weather, pet, ui}
ui    -> {agent (events), pet, lipgloss}
agent -> stdlib only
pet   -> stdlib only
weather -> net/http, encoding/json
```

- Domain packages (`agent`, `pet`) must not import `ui`, `bubbletea`, or `lipgloss`.
- The agent never performs IO. When a task needs IO (e.g. weather), it emits an
  event; the UI converts that event into a `tea.Cmd`.
- No package cycles. `agent` and `pet` never import `ui`.

## Development

Before considering a change complete:

```bash
gofmt -l .
go vet ./...
go test ./...
go build ./...
```

Prefer adding tests for new behavior and bug fixes.

## Code style

Follow standard Go conventions and run `gofmt`.

- Keep functions small and focused; prefer guard clauses and early returns over
  deep nesting (max ~2 levels of indentation).
- Keep files under ~500 lines; split by responsibility.
- Use specific, grep-friendly names. Avoid vague ones like `data`, `handler`,
  `manager`, `util`.
- Prefer explicit types. Avoid `any`/`interface{}` except at IO boundaries
  (JSON, HTTP).
- Wrap errors with context: `fmt.Errorf("fetching forecast for %s: %w", city, err)`.
  Do not swallow errors; return them.
- Abstract only at IO/trust boundaries (HTTP, filesystem, Bubble Tea). Do not add
  premature abstractions around pure logic.
- Time and randomness must be injected (a clock and a `*rand.Rand`); never read
  them from package-level state.
- Comments carry intent and provenance. Explain why, not what. Do not restate the code.

## Tests

- Every new behavior gets a test; every bug fix gets a regression test.
- Tests must run headless with `go test ./...`: no manual setup, no external
  services. Use fakes and `httptest` for IO; inject a fake clock.
- Prefer table-driven tests and named fakes over inline stubs.
- Include golden tests for the layout at the supported size matrix.
- Inject dependencies; do not rely on package-level state.

## Logging and output

- In alt-screen mode `stdout` is owned by Bubble Tea. Never write to `stdout`
  directly; send diagnostics to `stderr` or a log file.
- Use structured logging (`log/slog`) outside the TUI. Messages must be
  parseable, not prose.
- Never log credentials, tokens, or user data.

## Commits

Use Conventional Commits.

Format:

```text
<type>(<scope>): <description>
```

Common types: `feat`, `fix`, `refactor`, `perf`, `test`, `docs`, `chore`,
`build`, `ci`.

Examples:

```text
feat(agent): add water reminder scheduler
feat(weather): add open-meteo geocoding
fix(layout): keep banner visible in tiny terminals
test(agent): cover daily reset with fake clock
refactor(pet): extract behavior transition table
docs: document config fields
```

Keep commits focused on a single logical change. Do not combine unrelated
changes in the same commit.

## Pull requests

- Explain what changed and why.
- Include tests when appropriate.
- Keep changes focused.
- Avoid unrelated formatting or refactoring.
- Document breaking changes.

## Agent behavior

Before making significant changes:

1. Inspect the existing architecture.
2. Identify the appropriate package.
3. Follow existing patterns.
4. Avoid modifying unrelated code.
5. Run relevant tests after the change.

Do not create new files, abstractions, dependencies, or architecture without a
reason. When requirements are ambiguous, prefer the smallest implementation
consistent with the project's architecture.
