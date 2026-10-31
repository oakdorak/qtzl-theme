package game

// Level es una experiencia terapeutica con ritmo fijo pre-diseñado.
type Level struct {
	Name    string
	BPM     float64
	Scale   []float64
	Circles []*HitCircle
	Sliders []*HitSlider
}

func NewLevel(name string, bpm float64, scale []float64, circles []*HitCircle, sliders []*HitSlider) *Level {
	return &Level{Name: name, BPM: bpm, Scale: scale, Circles: circles, Sliders: sliders}
}

var Levels []*Level

func init() {
	// Nivel 1: Respiro Serena — 64 BPM, Pentatónica Mayor de Do, respiracion 4-4
	l1 := NewLevel("Respiracion Serena", 64, ScaleCPentatonic, []*HitCircle{}, []*HitSlider{})
	Levels = append(Levels, l1)

	// Nivel 2: Alternancia Bilateral — 82 BPM, modo Lydio de Fa
	l2 := NewLevel("Alternancia Bilateral", 82, ScaleFLydian, []*HitCircle{}, []*HitSlider{})
	Levels = append(Levels, l2)

	// Nivel 3: Enfoque Profundo / Flow — 104 BPM, acordes Cmaj9 / Am7
	l3 := NewLevel("Enfoque Profundo", 104, ScaleCmaj9, []*HitCircle{}, []*HitSlider{})
	Levels = append(Levels, l3)

	// Nivel 4: Traviesa de Odiseo — 90 BPM, modo Lydio de Sol
	l4 := NewLevel("Traviesa de Odiseo", 90, ScaleGLydian, []*HitCircle{}, []*HitSlider{})
	Levels = append(Levels, l4)
}
