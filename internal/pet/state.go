package pet

// State is the pet's current behavior.
type State uint8

const (
	Idle State = iota
	Walk
	Sleep
	Wake
	Happy
	Curious
)

func (s State) String() string {
	switch s {
	case Idle:
		return "idle"
	case Walk:
		return "walk"
	case Sleep:
		return "sleep"
	case Wake:
		return "wake"
	case Happy:
		return "happy"
	case Curious:
		return "curious"
	default:
		return "unknown"
	}
}
