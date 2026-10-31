package game

import (
	"encoding/binary"
	"math"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	SampleRate = 44100
)

var (
	globalAudioCtx  *audio.Context
	globalAudioOnce sync.Once
)

func getAudioContext() *audio.Context {
	globalAudioOnce.Do(func() {
		globalAudioCtx = audio.NewContext(SampleRate)
	})
	return globalAudioCtx
}

// Synth maneja la síntesis procedural pura (senoidales armónicas y pad ambiental).
type Synth struct {
	ctx          *audio.Context
	notePlayers  map[float64][]*audio.Player
	noteIdx      map[float64]int
	padPlayer    *audio.Player
	mu           sync.Mutex
	isMuted      bool
}

// Escalas musicales terapéuticas precalculadas (frecuencias en Hz)
var (
	// Escala Pentatónica Mayor de Do (Calmante, sin disonancias): Do, Re, Mi, Sol, La, Do5, Re5, Mi5
	ScaleCPentatonic = []float64{
		261.63, 293.66, 329.63, 392.00, 440.00, 523.25, 587.33, 659.25, 783.99, 880.00,
	}

	// Modo Lidio de Fa (Brillo suave, apertura cognitiva): Fa, Sol, La, Si, Do, Re, Mi, Fa5
	ScaleFLydian = []float64{
		174.61, 196.00, 220.00, 246.94, 261.63, 293.66, 329.63, 349.23, 392.00, 440.00, 493.88, 523.25,
	}

	// Acorde Cmaj9 / Am7 (Flow state y enfoque): Do, Mi, Sol, Si, Re, Mi5
	ScaleCmaj9 = []float64{
		261.63, 329.63, 392.00, 493.88, 587.33, 659.25, 783.99, 987.77, 1174.66,
	}

	// Modo Lidio de Sol (Travesía Mística): Sol, La, Si, Do#, Re, Mi, Fa#, Sol5
	ScaleGLydian = []float64{
		196.00, 220.00, 246.94, 277.18, 293.66, 329.63, 369.99, 392.00, 440.00, 493.88, 554.37, 587.33,
	}
)

func NewSynth() *Synth {
	ctx := getAudioContext()
	s := &Synth{
		ctx:         ctx,
		notePlayers: make(map[float64][]*audio.Player),
		noteIdx:     make(map[float64]int),
	}

	// Pre-generar un pool polifónico de 4 reproductores por frecuencia para permitir notas superpuestas
	allFreqs := append([]float64{}, ScaleCPentatonic...)
	allFreqs = append(allFreqs, ScaleFLydian...)
	allFreqs = append(allFreqs, ScaleCmaj9...)
	allFreqs = append(allFreqs, ScaleGLydian...)

	uniqueFreqs := make(map[float64]bool)
	for _, f := range allFreqs {
		if !uniqueFreqs[f] {
			uniqueFreqs[f] = true
			pcm := generateChimePCM(f, 350*time.Millisecond, 0.22)
			players := make([]*audio.Player, 4)
			for i := 0; i < 4; i++ {
				players[i] = ctx.NewPlayerFromBytes(pcm)
			}
			s.notePlayers[f] = players
			s.noteIdx[f] = 0
		}
	}

	// Generar pad ambiental continuo en loop (acorde Cmaj9 suspendido, 4 segundos en loop suave)
	padPCM := generatePadPCM(4 * time.Second)
	s.padPlayer = ctx.NewPlayerFromBytes(padPCM)

	return s
}

// StartPad inicia el tapiz armónico ambiental de fondo.
func (s *Synth) StartPad() {
	if s.padPlayer != nil && !s.padPlayer.IsPlaying() {
		_ = s.padPlayer.Rewind()
		s.padPlayer.Play()
	}
}

// Update mantiene el loop del pad ambiental si terminó.
func (s *Synth) Update() {
	if s.padPlayer != nil && !s.padPlayer.IsPlaying() {
		_ = s.padPlayer.Rewind()
		s.padPlayer.Play()
	}
}

// PlayNote reproduce una campanilla armónica a la frecuencia indicada con envolvente suave.
func (s *Synth) PlayNote(freq float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	players, ok := s.notePlayers[freq]
	if !ok || len(players) == 0 {
		return
	}

	idx := s.noteIdx[freq]
	p := players[idx]
	s.noteIdx[freq] = (idx + 1) % len(players)

	_ = p.Rewind()
	p.Play()
}

// PlayScaleIndex reproduce la n-ésima nota de una escala dada.
func (s *Synth) PlayScaleIndex(scale []float64, idx int) {
	if len(scale) == 0 {
		return
	}
	note := scale[idx%len(scale)]
	s.PlayNote(note)
}

// generateChimePCM genera una onda compuesta (fundamental + segundo armónico cálido)
// con envolvente ADSR suave sin transitorios desagradables.
func generateChimePCM(freq float64, dur time.Duration, amp float64) []byte {
	numSamples := int(float64(SampleRate) * dur.Seconds())
	if numSamples < 1 {
		numSamples = 1
	}

	buf := make([]int16, numSamples)
	attackSamples := int(math.Round(0.015 * float64(SampleRate))) // 15ms de ataque suave (evita clicks)

	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		decay := math.Exp(-t * 8.0) // decaimiento exponencial natural

		// Envolvente de entrada suave
		env := decay
		if i < attackSamples {
			env = float64(i) / float64(attackSamples) * decay
		}

		// Fundamental senoidal + 2do armónico sutil (-12dB)
		sampleF := math.Sin(2*math.Pi*freq*t) + 0.25*math.Sin(2*math.Pi*freq*2*t)
		val := int16(amp * env * sampleF * 32767.0)
		buf[i] = val
	}

	out := make([]byte, numSamples*2)
	for i, s := range buf {
		binary.LittleEndian.PutUint16(out[i*2:], uint16(s))
	}
	return out
}

// generatePadPCM genera un pad armónico continuo (C - G - E - B) con modulación sutil LFO.
func generatePadPCM(dur time.Duration) []byte {
	numSamples := int(float64(SampleRate) * dur.Seconds())
	buf := make([]int16, numSamples)

	// Frecuencias base del acorde cálido Cmaj7 (Do3, Sol3, Mi4, Si4)
	freqs := []float64{130.81, 196.00, 329.63, 493.88}
	weights := []float64{0.08, 0.06, 0.04, 0.03}

	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		var sum float64

		// LFO sutil a 0.25 Hz para sensación orgánica de respiración
		lfo := 0.85 + 0.15*math.Sin(2*math.Pi*0.25*t)

		for j, f := range freqs {
			sum += weights[j] * math.Sin(2*math.Pi*f*t)
		}

		// Fade in / fade out en los extremos del buffer para loop imperceptible
		edgeSamples := int(0.2 * SampleRate)
		window := 1.0
		if i < edgeSamples {
			window = float64(i) / float64(edgeSamples)
		} else if i > numSamples-edgeSamples {
			window = float64(numSamples-i) / float64(edgeSamples)
		}

		val := int16(sum * lfo * window * 32767.0)
		buf[i] = val
	}

	out := make([]byte, numSamples*2)
	for i, s := range buf {
		binary.LittleEndian.PutUint16(out[i*2:], uint16(s))
	}
	return out
}
