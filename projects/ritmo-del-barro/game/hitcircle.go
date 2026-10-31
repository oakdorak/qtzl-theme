package game

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// HitCircle es un objetivo de ritmo tipo "osu!" con anillo de aproximacion.
// El enfoque crece desde 3.0x hasta colapsar a 1.0 en HitTime.
type HitCircle struct {
	X            float64
	Y            float64
	Radius       float64
	SpawnTime    time.Time
	HitTime      time.Time
	ApproachScale float64
	PitchNote    int
	Hit          bool
	Missed       bool
}

func NewHitCircle(x, y float64, radius float64, pitchNote int, now time.Time) *HitCircle {
	h := &HitCircle{X: x, Y: y, Radius: radius, PitchNote: pitchNote, Hit: false, Missed: false}
	h.SpawnTime = now
	h.HitTime = now.Add(time.Duration(BeatWindow) * 4)
	h.ApproachScale = 3.0
	return h
}

func (h *HitCircle) Update(now time.Time) {
	if h.Hit || h.Missed {
		return
	}
	if now.Before(h.SpawnTime) || now.After(h.HitTime) {
		if !now.Before(h.HitTime) {
			h.Missed = true
		}
		return
	}
	elapsed := now.Sub(h.SpawnTime)
	total := h.HitTime.Sub(h.SpawnTime)
	if total > 0 {
		h.ApproachScale = 3.0 * (1 - float64(elapsed)/float64(total))
	}
}

func (h *HitCircle) Draw(screen *ebiten.Image, now time.Time) {
	if h.Hit || h.Missed {
		return
	}
	r := float32(h.Radius * h.ApproachScale)
	if r < 8 {
		r = 8
	}
	vector.DrawFilledCircle(screen, float32(h.X), float32(h.Y), r, color.RGBA{235, 230, 166, 255}, false)
}
