package pet

import "time"

const (
	fullFrameDur    = 600 * time.Millisecond
	compactFrameDur = 500 * time.Millisecond
	miniFrameDur    = 400 * time.Millisecond
)

var spriteSet = map[State]map[Size]Sprite{
	Idle: {
		SizeFull: newSprite(fullFrameDur,
			[]string{` /\_/\ `, `( o.o )`, ` > ^ < `},
			[]string{` /\_/\ `, `( -.- )`, ` > ^ < `},
		),
		SizeCompact: newSprite(compactFrameDur,
			[]string{`/\_/\`, `(o.o)`},
			[]string{`/\_/\`, `(-.-)`},
		),
		SizeMini: newSprite(miniFrameDur, []string{`^._.^`}),
	},
	Walk: {
		SizeFull: newSprite(fullFrameDur,
			[]string{` /\_/\ `, `( o.o )`, ` > ^ < `},
			[]string{` /\_/\ `, `( o.o )`, ` < ^ > `},
		),
		SizeCompact: newSprite(compactFrameDur,
			[]string{`/\_/\`, `(o.o)`},
			[]string{`/\_/\`, `(o.o)`},
		),
		SizeMini: newSprite(miniFrameDur, []string{`^._.^`}),
	},
	Sleep: {
		SizeFull: newSprite(fullFrameDur,
			[]string{` /\_/\ `, `( -.- )`, ` > ^ < `},
			[]string{` /\_/\ `, `( ._. )`, ` > ^ < `},
		),
		SizeCompact: newSprite(compactFrameDur,
			[]string{`/\_/\`, `(-.-)`},
			[]string{`/\_/\`, `(.-.)`},
		),
		SizeMini: newSprite(miniFrameDur, []string{`-.-`}),
	},
	Wake: {
		SizeFull: newSprite(fullFrameDur,
			[]string{` /\_/\ `, `( o.o )`, ` > ^ < `},
			[]string{` /\_/\ `, `( o.o )`, ` > ^ < `},
		),
		SizeCompact: newSprite(compactFrameDur,
			[]string{`/\_/\`, `(o.o)`},
			[]string{`/\_/\`, `(o.o)`},
		),
		SizeMini: newSprite(miniFrameDur, []string{`^._.^`}),
	},
	Happy: {
		SizeFull: newSprite(fullFrameDur,
			[]string{` /\_/\ `, `( ^.^ )`, ` > ^ < `},
			[]string{` /\_/\ `, `( ^o^ )`, ` > ^ < `},
		),
		SizeCompact: newSprite(compactFrameDur,
			[]string{`/\_/\`, `(^.^)`},
			[]string{`/\_/\`, `(^o^)`},
		),
		SizeMini: newSprite(miniFrameDur, []string{`^o^`}),
	},
	Curious: {
		SizeFull: newSprite(fullFrameDur,
			[]string{` /\_/\ `, `( o.O )`, ` > ^ < `},
			[]string{` /\_/\ `, `( O.o )`, ` > ^ < `},
		),
		SizeCompact: newSprite(compactFrameDur,
			[]string{`/\_/\`, `(o.O)`},
			[]string{`/\_/\`, `(O.o)`},
		),
		SizeMini: newSprite(miniFrameDur, []string{`o.O`}),
	},
}
