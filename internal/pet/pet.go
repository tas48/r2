package pet

import (
	"math/rand"
	"time"
)

// walkSpeed is horizontal movement in cells per second.
const walkSpeed = 4.0

// EventKind is a signal the pet reacts to. The pet never knows about the
// agent; the UI translates agent events into these.
type EventKind uint8

const (
	EventWater EventKind = iota
	EventBreak
	EventReminder
	EventInteract
	EventDayRolled
)

// Event is a reaction trigger.
type Event struct {
	Kind EventKind
}

// Pet is the pure state machine, animation clock and position of the companion.
type Pet struct {
	state State
	since time.Time
	until time.Time
	x     float64
	dir   int
	rng   *rand.Rand
}

// New creates a pet. rng may be nil, in which case a fixed seed is used so the
// behavior stays deterministic.
func New(now time.Time, rng *rand.Rand) *Pet {
	if rng == nil {
		rng = rand.New(rand.NewSource(1))
	}
	p := &Pet{rng: rng, x: -1, dir: 1}
	p.enter(Idle, now)
	return p
}

func (p *Pet) State() State     { return p.state }
func (p *Pet) Since() time.Time { return p.since }
func (p *Pet) Direction() int   { return p.dir }

// X returns the horizontal cell, or -1 when the pet has not been positioned
// yet and should be centered.
func (p *Pet) X() int { return int(p.x) }

// Step advances the animation clock and position, then transitions when the
// current state has lasted long enough. maxX is the highest allowed column.
func (p *Pet) Step(dt time.Duration, now time.Time, maxX int) {
	if p.state == Walk {
		p.walk(dt, maxX)
	}
	if !p.until.IsZero() && !now.Before(p.until) {
		p.transition(now)
	}
}

// React forces a state in response to an external event.
func (p *Pet) React(ev Event, now time.Time) {
	switch ev.Kind {
	case EventWater, EventBreak, EventInteract:
		p.enter(Happy, now)
	case EventReminder:
		p.enter(Curious, now)
	case EventDayRolled:
		p.enter(Wake, now)
	}
}

func (p *Pet) walk(dt time.Duration, maxX int) {
	if p.x < 0 {
		p.x = float64(maxX) / 2
	}
	p.x += float64(p.dir) * walkSpeed * dt.Seconds()
	if p.x <= 0 {
		p.x, p.dir = 0, 1
	}
	if p.x >= float64(maxX) {
		p.x, p.dir = float64(maxX), -1
	}
}

func (p *Pet) transition(now time.Time) {
	rules := transitions[p.state]
	if len(rules) == 0 {
		p.enter(Idle, now)
		return
	}
	total := 0
	for _, r := range rules {
		total += r.weight
	}
	n := p.rng.Intn(total)
	for _, r := range rules {
		n -= r.weight
		if n < 0 {
			p.enter(r.to, now)
			return
		}
	}
	p.enter(rules[0].to, now)
}

func (p *Pet) enter(state State, now time.Time) {
	p.state = state
	p.since = now
	p.until = now.Add(p.stay(state))
	if state == Walk {
		p.dir = 1
		if p.rng.Intn(2) == 0 {
			p.dir = -1
		}
	}
}

func (p *Pet) stay(state State) time.Duration {
	r := stays[state]
	if r[1] <= r[0] {
		return r[0]
	}
	return r[0] + time.Duration(p.rng.Int63n(int64(r[1]-r[0])))
}
