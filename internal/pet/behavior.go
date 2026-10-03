package pet

import "time"

type behavior struct {
	to     State
	weight int
}

// transitions is the data-driven automaton: from each state, weighted choices.
var transitions = map[State][]behavior{
	Idle:    {{to: Walk, weight: 3}, {to: Sleep, weight: 1}},
	Walk:    {{to: Idle, weight: 1}},
	Sleep:   {{to: Wake, weight: 1}},
	Wake:    {{to: Idle, weight: 1}},
	Happy:   {{to: Idle, weight: 1}},
	Curious: {{to: Idle, weight: 1}},
}

// stays is the min and max time a state persists before transitioning.
var stays = map[State][2]time.Duration{
	Idle:    {3 * time.Second, 8 * time.Second},
	Walk:    {3 * time.Second, 8 * time.Second},
	Sleep:   {8 * time.Second, 20 * time.Second},
	Wake:    {1 * time.Second, 2 * time.Second},
	Happy:   {1 * time.Second, 3 * time.Second},
	Curious: {2 * time.Second, 5 * time.Second},
}
