package game

import (
	"image/color"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	ScreenW = 320
	ScreenH = 180
	groundY = 156.0
)

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StateReflection
)

type Game struct {
	beat     *Beat
	beeper   *Beeper
	synth    *Synth
	score    int
	combo    int
	state    GameState
	level    *Level
	circles  []*HitCircle
	sliders  []*HitSlider
	player   *Player
	platform *Platform
	started  time.Time
	finished bool
}

func NewGame() *Game {
	return &Game{
		beat:   NewBeat(Levels[0].BPM),
		beeper: NewBeeper(),
		synth:  NewSynth(),
		state:  StateMenu,
		level:  Levels[0],
	}
}

func (g *Game) Init() error { return nil }

func (g *Game) StartLevel(idx int) {
	g.level = Levels[idx]
	g.state = StatePlaying
	g.beat = NewBeat(g.level.BPM)
	g.player = NewPlayer()
	g.platform = NewPlatform(120, groundY, 200, 20)
	g.circles = g.buildCircles()
	g.sliders = g.buildSliders()
	g.score = 0
	g.combo = 0
	g.finished = false
	g.started = time.Now()
}

func (g *Game) buildCircles() []*HitCircle {
	now := g.started
	var out []*HitCircle
	for i := 0; i < 4; i++ {
		spawn := now.Add(time.Duration(i) * time.Second)
		pitch := 0
		if len(g.level.Scale) > 0 {
			pitch = i % len(g.level.Scale)
		}
		out = append(out, &HitCircle{
			X:             float64(60 + i*50),
			Y:             60,
			Radius:       12,
			SpawnTime:    spawn,
			HitTime:      spawn.Add(150 * time.Millisecond),
			ApproachScale: 3.0,
			PitchNote:    pitch,
			Hit:           false,
			Missed:        false,
		})
	}
	return out
}

func (g *Game) buildSliders() []*HitSlider {
	return []*HitSlider{
		NewHitSlider(
			Vec2{X: 40, Y: 80},
			Vec2{X: 280, Y: 100},
			Vec2{X: 160, Y: 70},
			2 * time.Second,
			g.started,
		),
	}
}

func (g *Game) startFromMenuTap(x, y float64) {
	idx := 0
	for i := 0; i < len(Levels); i++ {
		yy := 30 + i*38
		if y > float64(yy) && y < float64(yy+30) {
			idx = i
			break
		}
	}
	g.StartLevel(idx)
}

func (g *Game) Update() error {
	switch g.state {
	case StateMenu:
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			mx, my := ebiten.CursorPosition()
			g.handleTap(float64(mx), float64(my))
		}
		return nil
	case StatePlaying:
		g.beat.Update(time.Now())
		if g.beat.JustTicked() {
			g.beeper.Tick()
		}
		g.synth.Update()
		now := time.Now()
		for _, c := range g.circles {
			c.Update(now)
		}
		for _, s := range g.sliders {
			s.Update(now)
		}
		g.player.Update(g.beat, g.platform, groundY)
		g.handleInput(now)
		if g.checkFinished() {
			g.state = StateReflection
		}
		return nil
	case StateReflection:
		return nil
	}
	return nil
}

func (g *Game) handleInput(now time.Time) {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	mx, my := ebiten.CursorPosition()
	g.handleTap(float64(mx), float64(my))
}

func (g *Game) handleTap(x, y float64) {
	if g.state == StateMenu {
		g.startFromMenuTap(x, y)
		return
	}
	if g.state != StatePlaying {
		return
	}
	for _, c := range g.circles {
		if c.Hit || c.Missed {
			continue
		}
		if x > c.X-c.Radius && x < c.X+c.Radius && y > c.Y-c.Radius && y < c.Y+c.Radius {
			c.Hit = true
			g.score += 10
			g.combo++
			if len(g.level.Scale) > 0 {
				g.synth.PlayScaleIndex(g.level.Scale, c.PitchNote)
			}
		}
	}
	for _, s := range g.sliders {
		if s.Hit || s.Missed {
			continue
		}
		if time.Now().After(s.StartTime.Add(s.Duration)) {
			continue
		}
		if math.Hypot(x-s.CurrentPos.X, y-s.CurrentPos.Y) < 10 {
			s.Hit = true
			g.score += 25
			g.combo++
		}
	}
}

func (g *Game) checkFinished() bool {
	if len(g.circles) == 0 {
		g.state = StateReflection
		return true
	}
	for _, c := range g.circles {
		if !c.Hit && !c.Missed {
			return false
		}
	}
	for _, s := range g.sliders {
		if !s.Hit && !s.Missed {
			return false
		}
	}
	g.state = StateReflection
	return true
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{13, 11, 22, 255})
	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying:
		for _, c := range g.circles {
			c.Draw(screen, time.Now())
		}
		for _, s := range g.sliders {
			s.Draw(screen, time.Now())
		}
		g.platform.Draw(screen, g.beat.phase)
		g.player.Draw(screen)
		g.drawHUD(screen)
	case StateReflection:
		g.drawReflection(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Fondo + titulo
	vector.DrawFilledRect(screen, 0, 0, ScreenW, 20, color.RGBA{74, 205, 123, 255}, false)
	ebitenutil.DebugPrint(screen, "RITMO DEL BARRO")
	for i, l := range Levels {
		y := 30 + i*38
		w := float32(ScreenW - 16)
		h := float32(30)
		vector.DrawFilledRect(screen, 8, float32(y), w, h, color.RGBA{35, 30, 54, 255}, false)
		ebitenutil.DebugPrint(screen, l.Name)
		ebitenutil.DebugPrint(screen, fmt.Sprintf("%v bpm", g.level.BPM))
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	mx, my := ebiten.CursorPosition()
	ebitenutil.DebugPrint(screen, "ritmo-del-barro score")
	_ = mx
	_ = my
}

func (g *Game) drawReflection(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "REFLEXION COMPLETADA")
	ebitenutil.DebugPrint(screen, "Puntuación: "+strconv.Itoa(g.score))
	ebitenutil.DebugPrint(screen, "Combo: "+strconv.Itoa(g.combo))
}

func (g *Game) Layout(w, h int) (int, int) { return ScreenW, ScreenH }
