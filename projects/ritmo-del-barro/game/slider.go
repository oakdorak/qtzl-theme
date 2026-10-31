package game

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Vec2 es un punto 2D en coordenadas logicas de pantalla.
type Vec2 struct {
	X float64
	Y float64
}

// HitSlider es un objetivo de tipo slider: el jugador debe deslizar el dedo
// (o cursor) siguiendo una curva Bezier cuadratica hasta el final.
type HitSlider struct {
	Start       Vec2
	End         Vec2
	ControlPoint Vec2
	StartTime  time.Time
	Duration   time.Duration
	CurrentPos Vec2
	Tracking   bool
	Hit        bool
	Missed     bool
}

func NewHitSlider(start, end Vec2, control Vec2, duration time.Duration, now time.Time) *HitSlider {
	s := &HitSlider{Start: start, End: end, ControlPoint: control, StartTime: now, Duration: duration, Hit: false, Missed: false}
	s.CurrentPos = start
	return s
}

func (s *HitSlider) Update(now time.Time) {
	if s.Hit || s.Missed {
		return
	}
	hitAt := s.StartTime.Add(s.Duration)
	if now.Before(s.StartTime) || now.After(hitAt) {
		if !now.Before(hitAt) {
			s.Missed = true
		}
		return
	}
	t := float64(now.Sub(s.StartTime)) / float64(s.Duration)
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	s.CurrentPos = quadraticBezier(s.Start, s.ControlPoint, s.End, t)
}

// Draw dibuja el slider en su CurrentPos actual de traccion.
func (s *HitSlider) Draw(screen *ebiten.Image, now time.Time) {
	if s.Hit || s.Missed {
		return
	}
	vector.DrawFilledCircle(screen, float32(s.CurrentPos.X), float32(s.CurrentPos.Y), 5, color.RGBA{110, 186, 222, 255}, false)
}

func quadraticBezier(p0, p1, p2 Vec2, t float64) Vec2 {
	u := 1 - t
	return Vec2{
		X: u*u*p0.X + 2*u*t*p1.X + t*t*p2.X,
		Y: u*u*p0.Y + 2*u*t*p1.Y + t*t*p2.Y,
	}
}
