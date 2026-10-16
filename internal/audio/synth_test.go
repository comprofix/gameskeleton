package audio

import (
	"encoding/binary"
	"math"
	"testing"
	"time"

	"github.com/comprofix/gameskeleton/internal/config"
)

const sr = config.AudioSampleRate

// render reads n samples from s and returns the left channel, -1 to 1,
// checking the right channel matches.
func render(t *testing.T, s *Synth, n int) []float64 {
	t.Helper()
	buf := make([]byte, 4*n)
	got, err := s.Read(buf)
	if err != nil || got != len(buf) {
		t.Fatalf("Read = %d, %v; want %d, nil", got, err, len(buf))
	}
	out := make([]float64, n)
	for i := range out {
		l := int16(binary.LittleEndian.Uint16(buf[4*i:]))
		r := int16(binary.LittleEndian.Uint16(buf[4*i+2:]))
		if l != r {
			t.Fatalf("sample %d: left %d != right %d", i, l, r)
		}
		out[i] = float64(l) / math.MaxInt16
	}
	return out
}

func peak(xs []float64) float64 {
	p := 0.0
	for _, x := range xs {
		p = max(p, math.Abs(x))
	}
	return p
}

func rms(xs []float64) float64 {
	sum := 0.0
	for _, x := range xs {
		sum += x * x
	}
	return math.Sqrt(sum / float64(len(xs)))
}

// crossings counts sign changes from negative to positive.
func crossings(xs []float64) int {
	n := 0
	for i := 1; i < len(xs); i++ {
		if xs[i-1] < 0 && xs[i] >= 0 {
			n++
		}
	}
	return n
}

// tone is a held note on one channel.
func tone(w Wave, hz float64, d time.Duration) *Effect {
	return &Effect{Name: "tone", Part: Part{Voice: Voice{Wave: w}, Notes: []Note{{Hz: hz, Dur: d}}}}
}

func TestPitch(t *testing.T) {
	for name, want := range map[string]float64{
		"A4": 440, "A5": 880, "A3": 220, "C4": 261.626, "C#4": 277.183, "Db4": 277.183, "B3": 246.942, "C0": 16.352,
	} {
		got, err := Pitch(name)
		if err != nil || math.Abs(got-want) > 0.01 {
			t.Errorf("Pitch(%q) = %v, %v; want %v", name, got, err, want)
		}
	}
	for _, bad := range []string{"", "A", "H4", "A#", "Ax4", "A10", "a4", "C-1"} {
		if _, err := Pitch(bad); err == nil {
			t.Errorf("Pitch(%q) gave no error", bad)
		}
	}
}

func TestNotes(t *testing.T) {
	got := Notes("C4:12 -:6", "A4:60")
	want := []Note{{Hz: 261.6255653005986, Dur: Ticks(12)}, {Dur: Ticks(6)}, {Hz: 440, Dur: time.Second}}
	if len(got) != len(want) {
		t.Fatalf("Notes = %v, want %v", got, want)
	}
	for i := range want {
		if math.Abs(got[i].Hz-want[i].Hz) > 1e-9 || got[i].Dur != want[i].Dur {
			t.Errorf("note %d = %v, want %v", i, got[i], want[i])
		}
	}
	if Length(got) != Ticks(78) {
		t.Errorf("Length = %v, want %v", Length(got), Ticks(78))
	}
	for _, bad := range []string{"C4", "C4:", "C4:0", "C4:x", "Q4:3", "-:-1"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Notes(%q) did not panic", bad)
				}
			}()
			Notes(bad)
		}()
	}
}

func TestNilSynthIsSafe(t *testing.T) {
	var s *Synth
	s.Now(10)
	s.PlayTune(&Tune{})
	s.PlayEffect(tone(Pulse50, 440, time.Second))
	s.StopTune()
	s.Stop()
	s.SetEnabled(false)
	s.SetVolumes(1, 1, 1)
	if s.Tune() != nil || s.Channels() != 0 {
		t.Error("nil synth reports a tune or channels")
	}
}

func TestChannels(t *testing.T) {
	if n := NewSynth(config.Chip4).Channels(); n != 4 {
		t.Errorf("Chip4 has %d channels, want 4", n)
	}
	if n := NewSynth(config.Speaker).Channels(); n != 1 {
		t.Errorf("Speaker has %d channels, want 1", n)
	}
}

func TestSilentWithNothingPlaying(t *testing.T) {
	if p := peak(render(t, NewSynth(config.Chip4), sr/10)); p != 0 {
		t.Errorf("peak %v with nothing playing, want 0", p)
	}
}

func TestPitchOfEachWave(t *testing.T) {
	for _, w := range []Wave{Pulse50, Pulse25, Pulse12, Triangle} {
		s := NewSynth(config.Chip4)
		s.PlayEffect(tone(w, 441, time.Second))
		out := render(t, s, sr)
		if n := crossings(out[sr/10:]); n < 395 || n > 399 { // 441 Hz over 0.9 s
			t.Errorf("wave %d: %d upward crossings in 0.9 s, want about 397", w, n)
		}
	}
}

func TestPulseDuty(t *testing.T) {
	const dt = 100.0 / sr
	for duty, w := range map[float64]Wave{0.5: Pulse50, 0.25: Pulse25, 0.125: Pulse12} {
		c := channel{voice: Voice{Wave: w}, hz: 100}
		high := 0
		for range sr / 100 { // one cycle
			if c.wave() > 0 {
				high++
			}
		}
		if got := float64(high) * dt; math.Abs(got-duty) > 0.01 {
			t.Errorf("wave %d is high %.3f of the time, want %v", w, got, duty)
		}
	}
}

func TestNoiseIsNotATone(t *testing.T) {
	s := NewSynth(config.Chip4)
	s.PlayEffect(tone(Noise, 4000, time.Second))
	out := render(t, s, sr/2)
	if rms(out) < 0.05 {
		t.Fatalf("noise is too quiet: rms %v", rms(out))
	}
	// Noise keeps changing; a stuck shift register would go quiet once the
	// DC filter settles.
	if r := rms(out[len(out)-sr/20:]); r < 0.05 {
		t.Errorf("noise died away: rms %v at the end", r)
	}
}

func TestEnvelope(t *testing.T) {
	s := NewSynth(config.Chip4)
	e := &Effect{Part: Part{
		Voice: Voice{Wave: Pulse50, Attack: 100 * time.Millisecond, Release: 100 * time.Millisecond},
		Notes: []Note{{Hz: 441, Dur: 300 * time.Millisecond}},
	}}
	s.PlayEffect(e)
	out := render(t, s, sr/2)
	ms := func(a, b int) []float64 { return out[a*sr/1000 : b*sr/1000] }
	early, full, fading, after := peak(ms(0, 20)), peak(ms(150, 250)), peak(ms(350, 380)), peak(ms(420, 500))
	if !(early < full/2) {
		t.Errorf("attack: peak %v in the first 20 ms, want under half of %v", early, full)
	}
	if !(fading > 0.1*full && fading < full) {
		t.Errorf("release: peak %v 50 ms after the note, want between 0.1 and 1 × %v", fading, full)
	}
	if after > 0.02*full {
		t.Errorf("peak %v after the release, want silence", after)
	}
}

func TestDecayToSustain(t *testing.T) {
	s := NewSynth(config.Chip4)
	e := &Effect{Part: Part{
		Voice: Voice{Wave: Pulse50, Decay: 50 * time.Millisecond, Sustain: 0.25},
		Notes: []Note{{Hz: 441, Dur: time.Second}},
	}}
	s.PlayEffect(e)
	out := render(t, s, sr/2)
	start, held := rms(out[0:sr/100]), rms(out[sr/5:sr*3/10])
	if r := held / start; r < 0.2 || r > 0.35 {
		t.Errorf("held level is %.2f of the start, want about 0.25", r)
	}
}

func TestVolumes(t *testing.T) {
	music := &Tune{Parts: []Part{{Voice: Voice{Wave: Pulse50}, Notes: []Note{{Hz: 441, Dur: time.Second}}}}}
	effect := tone(Pulse50, 441, time.Second)
	effect.Channel = 1
	level := func(master, mus, sfx float64, enabled bool, tune *Tune, e *Effect) float64 {
		s := NewSynth(config.Chip4)
		s.SetVolumes(master, mus, sfx)
		s.SetEnabled(enabled)
		s.PlayTune(tune)
		s.PlayEffect(e)
		return rms(render(t, s, sr/2)[sr/4:])
	}
	full := level(1, 1, 1, true, music, nil)
	if full == 0 {
		t.Fatal("no sound at full volume")
	}
	for _, c := range []struct {
		name string
		got  float64
		want float64 // as a fraction of full
	}{
		{"half master", level(0.5, 1, 1, true, music, nil), 0.5},
		{"half music", level(1, 0.5, 1, true, music, nil), 0.5},
		{"music off, effect plays", level(1, 0, 1, true, music, effect), 1},
		{"effects off, music plays", level(1, 1, 0, true, music, effect), 1},
		{"master off", level(0, 1, 1, true, music, effect), 0},
		{"disabled", level(1, 1, 1, false, music, effect), 0},
	} {
		if r := c.got / full; math.Abs(r-c.want) > 0.02 {
			t.Errorf("%s: level %.3f of full, want %v", c.name, r, c.want)
		}
	}
}

func TestEffectPriority(t *testing.T) {
	s := NewSynth(config.Chip4)
	hi := &Effect{Name: "hi", Priority: 2, Part: Part{Notes: []Note{{Hz: 440, Dur: time.Second}}}}
	lo := &Effect{Name: "lo", Priority: 1, Part: Part{Notes: []Note{{Hz: 880, Dur: time.Second}}}}
	same := &Effect{Name: "same", Priority: 2, Part: Part{Notes: []Note{{Hz: 660, Dur: time.Second}}}}
	s.PlayEffect(hi)
	s.PlayEffect(lo)
	if p := s.ch[0].sfx.part; p != &hi.Part {
		t.Fatal("a lower-priority effect replaced a higher one")
	}
	s.PlayEffect(same)
	if p := s.ch[0].sfx.part; p != &same.Part {
		t.Fatal("an equal-priority effect did not replace the one playing")
	}
	render(t, s, sr+10) // let it finish
	s.PlayEffect(lo)
	if p := s.ch[0].sfx.part; p != &lo.Part {
		t.Error("a low-priority effect did not play once the channel was free")
	}
}

func TestEffectChannelIsClamped(t *testing.T) {
	s := NewSynth(config.Speaker)
	e := tone(Pulse50, 440, time.Second)
	e.Channel = 3
	s.PlayEffect(e)
	if !s.ch[0].sfx.active() {
		t.Error("an effect for channel 3 did not play on Speaker's one channel")
	}
}

// longTune is one note per second for a minute, so the test can see how far
// the music got.
func longTune() *Tune {
	var notes []Note
	for i := range 60 {
		notes = append(notes, Note{Hz: 200 + float64(i), Dur: time.Second})
	}
	return &Tune{Parts: []Part{{Voice: Voice{Wave: Pulse50}, Notes: notes}}}
}

func TestChip4MusicKeepsTimeUnderAnEffect(t *testing.T) {
	s := NewSynth(config.Chip4)
	s.PlayTune(longTune())
	render(t, s, sr/2)
	s.PlayEffect(tone(Pulse50, 880, time.Second))
	render(t, s, sr)
	if i := s.ch[0].music.i; i != 1 {
		t.Errorf("music is on note %d after 1.5 s, want 1", i)
	}
}

func TestSpeakerMusicWaitsForAnEffect(t *testing.T) {
	s := NewSynth(config.Speaker)
	s.PlayTune(longTune())
	render(t, s, sr/2)
	s.PlayEffect(tone(Pulse50, 880, time.Second))
	render(t, s, sr)
	m := &s.ch[0].music
	if m.i != 0 || m.left != sr/2 {
		t.Errorf("music at note %d with %d samples left, want note 0 with %d", m.i, m.left, sr/2)
	}
	// After the effect, the rest of note 0 plays, then note 1.
	render(t, s, sr/2+10)
	if m.i != 1 {
		t.Errorf("music on note %d after resuming, want 1", m.i)
	}
}

func TestSpeakerPlaysSquareWaves(t *testing.T) {
	s := NewSynth(config.Speaker)
	s.PlayEffect(tone(Triangle, 440, time.Second))
	render(t, s, 10)
	if w := s.ch[0].voice.Wave; w != Pulse50 {
		t.Errorf("Speaker plays a triangle as wave %d, want Pulse50", w)
	}
}

func TestSpeakerPlaysOnlyTheFirstPart(t *testing.T) {
	s := NewSynth(config.Speaker)
	tune := &Tune{Parts: []Part{
		{Notes: []Note{{Hz: 440, Dur: time.Second}}},
		{Notes: []Note{{Hz: 220, Dur: time.Second}}},
	}}
	s.PlayTune(tune)
	if s.ch[0].music.part != &tune.Parts[0] {
		t.Error("Speaker is not playing part 0")
	}
}

func TestTuneLoopsAndStops(t *testing.T) {
	s := NewSynth(config.Chip4)
	tune := &Tune{Loop: true, Parts: []Part{{Notes: []Note{{Hz: 440, Dur: 100 * time.Millisecond}}}}}
	s.PlayTune(tune)
	if s.Tune() != tune {
		t.Error("Tune() is not the tune just started")
	}
	render(t, s, sr/2)
	if !s.ch[0].music.active() {
		t.Error("a looping tune stopped")
	}
	s.StopTune()
	if s.Tune() != nil || s.ch[0].music.active() {
		t.Error("StopTune did not stop the music")
	}

	once := &Tune{Parts: []Part{{Notes: []Note{{Hz: 440, Dur: 100 * time.Millisecond}}}}}
	s.PlayTune(once)
	render(t, s, sr/5)
	if s.ch[0].music.active() {
		t.Error("a tune that does not loop is still playing")
	}
}

func TestEmptyLoopingPartStops(t *testing.T) {
	s := NewSynth(config.Chip4)
	s.PlayTune(&Tune{Loop: true, Parts: []Part{{Notes: []Note{{Hz: 440}}}}})
	render(t, s, 10) // must not hang
	if s.ch[0].music.active() {
		t.Error("a looping part with no length is still playing")
	}
}

// onset is the index of the first sample louder than a whisper.
func onset(xs []float64) int {
	for i, x := range xs {
		if math.Abs(x) > 1e-3 {
			return i
		}
	}
	return -1
}

func TestScheduledOnTheTick(t *testing.T) {
	const perTick = sr / config.TPS
	base := config.AudioScheduleMs * sr / 1000
	for _, tick := range []int64{1, 5, 9} {
		s := NewSynth(config.Chip4)
		s.Now(0)
		s.Now(tick)
		s.PlayEffect(tone(Pulse50, 441, time.Second))
		out := render(t, s, base+12*perTick)
		want := base + int(tick)*perTick
		// The note starts on its sample; the attack ramp means the first
		// audible sample is at most a couple later.
		if got := onset(out); got < want || got > want+2 {
			t.Errorf("tick %d: sound starts at sample %d, want %d", tick, got, want)
		}
	}
}

func TestScheduleWidensWhenLate(t *testing.T) {
	const perTick = sr / config.TPS
	base := int64(config.AudioScheduleMs * sr / 1000)
	s := NewSynth(config.Chip4)
	s.Now(0)
	// The device reads 100 ms ahead before the next tick: commands for tick
	// 1 would already be late, so the delay widens.
	render(t, s, sr/10+perTick)
	s.Now(1)
	if s.latency <= base {
		t.Errorf("latency %d did not widen past %d", s.latency, base)
	}
	if ahead := 1*perTick + s.offset - s.pos; ahead < 0 {
		t.Errorf("tick 1 is still %d samples late", -ahead)
	}
}

func TestCommandsKeepTheirOrder(t *testing.T) {
	s := NewSynth(config.Chip4)
	s.Now(0)
	s.PlayTune(longTune())
	s.StopTune()
	render(t, s, sr/5)
	if s.ch[0].music.active() {
		t.Error("a stop sent after a play did not win")
	}
}

func TestBankCheck(t *testing.T) {
	part := func(ticks ...int) Part {
		var p Part
		for _, n := range ticks {
			p.Notes = append(p.Notes, Note{Hz: 440, Dur: Ticks(n)})
		}
		return p
	}
	good := Bank{
		Tunes:   []*Tune{{Name: "t", Loop: true, Parts: []Part{part(10, 20), part(30)}}},
		Effects: []*Effect{{Name: "e", Channel: 3, Part: part(5)}},
	}
	if err := good.Check(); err != nil {
		t.Errorf("good bank: %v", err)
	}
	for name, b := range map[string]Bank{
		"no name":      {Effects: []*Effect{{Part: part(5)}}},
		"same name":    {Tunes: []*Tune{{Name: "x", Parts: []Part{part(1)}}}, Effects: []*Effect{{Name: "x", Part: part(5)}}},
		"empty part":   {Tunes: []*Tune{{Name: "t", Parts: []Part{{}}}}},
		"empty effect": {Effects: []*Effect{{Name: "e"}}},
		"channel 4":    {Effects: []*Effect{{Name: "e", Channel: 4, Part: part(5)}}},
		"drifting":     {Tunes: []*Tune{{Name: "t", Loop: true, Parts: []Part{part(10), part(11)}}}},
	} {
		if b.Check() == nil {
			t.Errorf("%s: no error", name)
		}
	}
	// Parts of a tune that doesn't loop may differ.
	once := Bank{Tunes: []*Tune{{Name: "t", Parts: []Part{part(10), part(11)}}}}
	if err := once.Check(); err != nil {
		t.Errorf("tune without loop: %v", err)
	}
}
