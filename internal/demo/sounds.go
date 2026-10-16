package demo

import "github.com/comprofix/gameskeleton/internal/audio"

// Dot Dash's sounds. Channels follow the shared tune layout (see
// scene.TitleTune): 0 lead, 1 harmony, 2 bass, 3 noise.
var (
	// Step is the player moving one cell.
	Step = &audio.Effect{Name: "step", Channel: 3, Priority: 1, Part: audio.Part{
		Voice: audio.Voice{Wave: audio.Noise, Decay: audio.Ticks(3), Volume: 0.5},
		Notes: []audio.Note{{Hz: 6000, Dur: audio.Ticks(3)}},
	}}
	// Pickup is a dot collected.
	Pickup = &audio.Effect{Name: "pickup", Channel: 1, Priority: 2, Part: audio.Part{
		Voice: audio.Voice{Wave: audio.Pulse25, Decay: audio.Ticks(8), Sustain: 0.4, Release: audio.Ticks(3), Volume: 0.8},
		Notes: audio.Notes("G6:3 C7:8"),
	}}
	// Bump is the player hitting a hazard.
	Bump = &audio.Effect{Name: "bump", Channel: 3, Priority: 3, Part: audio.Part{
		Voice: audio.Voice{Wave: audio.Noise, Decay: audio.Ticks(14), Release: audio.Ticks(2)},
		Notes: []audio.Note{{Hz: 900, Dur: audio.Ticks(14)}},
	}}
	// GameOver is the end of a game.
	GameOver = &audio.Effect{Name: "game over", Channel: 0, Priority: 5, Part: audio.Part{
		Voice: audio.Voice{Wave: audio.Pulse50, Decay: audio.Ticks(20), Sustain: 0.5, Release: audio.Ticks(12), Volume: 0.8},
		Notes: audio.Notes("G5:10 E5:10 C5:10 -:4 G4:36"),
	}}
)

// Sounds are the demo's sounds, for the sound test.
var Sounds = audio.Bank{Effects: []*audio.Effect{Step, Pickup, Bump, GameOver}}
