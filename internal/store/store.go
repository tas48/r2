package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const dayLayout = "2006-01-02"

// State is the persisted snapshot of the agent's counters.
type State struct {
	Day         string    `json:"day"`
	WaterML     int       `json:"water_ml"`
	LastDrink   time.Time `json:"last_drink"`
	LastBreak   time.Time `json:"last_break"`
	BreaksToday int       `json:"breaks_today"`
}

// Store persists State as a small JSON file.
type Store struct {
	path string
}

func New(path string) *Store {
	return &Store{path: path}
}

// Path returns the default state file location.
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolving user config directory: %w", err)
	}
	return filepath.Join(dir, "r2", "state.json"), nil
}

// Load reads the persisted state, resetting daily counters when the stored day
// differs from today. A missing file is not an error.
func (s *Store) Load(today time.Time) (State, error) {
	fresh := State{Day: today.Format(dayLayout)}
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return fresh, nil
	}
	if err != nil {
		return State{}, fmt.Errorf("reading state %s: %w", s.path, err)
	}
	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return State{}, fmt.Errorf("parsing state %s: %w", s.path, err)
	}
	if st.Day != fresh.Day {
		return fresh, nil
	}
	return st, nil
}

// Save writes the state atomically, creating the parent directory if needed.
func (s *Store) Save(st State) error {
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding state: %w", err)
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, "state-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp state file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temp state %s: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp state %s: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return fmt.Errorf("replacing state %s: %w", s.path, err)
	}
	return nil
}
