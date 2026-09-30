package agent

import "time"

// Kind identifies a recurring task of the agent.
type Kind uint8

const (
	KindWater Kind = iota
	KindBreak
)

func (k Kind) String() string {
	switch k {
	case KindWater:
		return "water"
	case KindBreak:
		return "break"
	default:
		return "unknown"
	}
}

// Event is produced by the agent on each Step. The agent never performs IO;
// the UI interprets events and turns the ones that need IO into tea.Cmd.
type Event interface{ isEvent() }

// ReminderFired asks the user to act (drink water, take a break).
type ReminderFired struct {
	Kind Kind
	Text string
	At   time.Time
}

func (ReminderFired) isEvent() {}

// WaterLogged reports that the user registered a bottle of water.
type WaterLogged struct {
	TotalML int
	GoalML  int
	At      time.Time
}

func (WaterLogged) isEvent() {}

// BreakLogged reports that the user registered a pause.
type BreakLogged struct {
	SinceLastBreak time.Duration
	At             time.Time
}

func (BreakLogged) isEvent() {}

// DayRolled reports a date change; daily counters have been reset.
type DayRolled struct {
	Date string
	At   time.Time
}

func (DayRolled) isEvent() {}
