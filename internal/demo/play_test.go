package demo

import (
	"math/rand/v2"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/gameskeleton/internal/audio"
	"github.com/comprofix/gameskeleton/internal/config"
	"github.com/comprofix/gameskeleton/internal/demo/world"
	"github.com/comprofix/gameskeleton/internal/input"
	"github.com/comprofix/gameskeleton/internal/scene"
	"github.com/comprofix/gameskeleton/internal/score"
)

func TestBoardFitsBothPresets(t *testing.T) {
	for _, p := range []config.Preset{config.Modern, config.Classic} {
		cols, rows := Board(p)
		if cols < 2*world.ForSkill("hard").SafeRadius+1 || rows < 8 {
			t.Errorf("%s: board %d×%d is too small", p.Name, cols, rows)
		}
	}
}

// TestAWholeGame plays a game headless, standing still until the clock runs
// out, then goes through game over, the tally and initials to the table.
func TestAWholeGame(t *testing.T) {
	set := config.DefaultSettings(config.Project)
	script := &input.Script{}
	ctx := &scene.Context{
		Input:    input.New(script),
		RNG:      rand.New(rand.NewPCG(3, 4)),
		Settings: &set,
		Scores:   score.New(),
		NewPlay:  NewPlay,
	}
	skill := config.Skills[0]
	play := NewPlay(ctx, skill).(*Play)
	mgr := scene.NewManager(play)
	ticks := func(n int) {
		for range n {
			ctx.Input.Update()
			if err := mgr.Update(ctx); err != nil {
				t.Fatal(err)
			}
		}
	}
	press := func(k ebiten.Key) {
		script.Steps = append(script.Steps, input.Press(k), input.Idle(1))
		ticks(2 + 2*scene.FadeTicks)
	}

	ticks(scene.FadeTicks + world.ForSkill(skill.ID).Ticks)
	if _, ok := mgr.Current().(*Ended); !ok {
		t.Fatalf("after the clock ran out the scene is %T", mgr.Current())
	}
	ticks(gameOverTicks + 2*scene.FadeTicks)
	tally, ok := mgr.Current().(*scene.Tally)
	if !ok {
		t.Fatalf("after game over the scene is %T", mgr.Current())
	}
	w := play.World()
	want := w.DotsTaken*w.Tuning.DotPoints + w.HeartBonus()
	if ctx.Session.Score != want {
		t.Errorf("final score %d, want dots plus hearts %d", ctx.Session.Score, want)
	}
	press(ebiten.KeyEnter) // finish counting
	if tally.Total() != ctx.Session.Score {
		t.Errorf("tally total %d, score %d", tally.Total(), ctx.Session.Score)
	}
	press(ebiten.KeyEnter)
	if want == 0 {
		t.Skip("bumped out with no score: nothing to enter") // not with this seed
	}
	if _, ok := mgr.Current().(*scene.HighScoreEntry); !ok {
		t.Fatalf("a first score went to %T, want initials entry", mgr.Current())
	}
	for range config.InitialsLen {
		press(ebiten.KeyEnter)
	}
	if _, ok := mgr.Current().(*scene.HighScoreTable); !ok {
		t.Fatalf("after the initials the scene is %T", mgr.Current())
	}
	if es := ctx.Scores.Table(skill.ID); len(es) != 1 || es[0].Score != want || es[0].Initials != "AAA" {
		t.Errorf("table %v", es)
	}
	if !ctx.TakeScoresChanged() {
		t.Error("the score was not reported for saving")
	}
}

func TestMusicDuringPlay(t *testing.T) {
	set := config.DefaultSettings(config.Project)
	ctx := &scene.Context{
		Input:    input.New(&input.Script{}),
		RNG:      rand.New(rand.NewPCG(3, 4)),
		Settings: &set,
		Scores:   score.New(),
		Audio:    audio.NewSynth(config.Chip4),
	}
	ctx.Audio.PlayTune(scene.TitleTune)
	play := NewPlay(ctx, config.Skills[0]).(*Play)
	if ctx.Audio.Tune() != DashTune {
		t.Fatal("starting a game did not start its tune")
	}
	for range world.ForSkill(config.Skills[0].ID).Ticks {
		ctx.Input.Update()
		play.Update(ctx)
	}
	if ctx.Audio.Tune() != nil {
		t.Error("the tune is still playing after the game ended")
	}
}
