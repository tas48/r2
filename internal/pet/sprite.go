package pet

import (
	"time"
	"unicode/utf8"
)

// Size selects a sprite variant based on the available stage height.
type Size uint8

const (
	SizeFull Size = iota
	SizeCompact
	SizeMini
)

// Sprite is a set of animation frames sharing a size.
type Sprite struct {
	Frames   [][]string
	FrameDur time.Duration
	Width    int
	Height   int
}

func newSprite(frameDur time.Duration, frames ...[]string) Sprite {
	width, height := 0, 0
	for _, frame := range frames {
		if len(frame) > height {
			height = len(frame)
		}
		for _, line := range frame {
			if n := utf8.RuneCountInString(line); n > width {
				width = n
			}
		}
	}
	return Sprite{Frames: frames, FrameDur: frameDur, Width: width, Height: height}
}

// SpriteFor returns the sprite for a state at the requested size, falling back
// to smaller variants when the size is unavailable.
func SpriteFor(state State, size Size) Sprite {
	bySize, ok := spriteSet[state]
	if !ok {
		bySize = spriteSet[Idle]
	}
	for s := size; s <= SizeMini; s++ {
		if sprite, ok := bySize[s]; ok {
			return sprite
		}
	}
	return bySize[SizeMini]
}
