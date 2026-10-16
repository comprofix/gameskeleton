// Package game implements ebiten.Game: it runs the scenes, handles the
// window keys, applies and saves the settings, and presents the internal
// screen scaled to the window (DESIGN §5, §7).
package game

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand/v2"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/comprofix/gameskeleton/assets"
	"github.com/comprofix/gameskeleton/internal/config"
	"github.com/comprofix/gameskeleton/internal/demo"
	"github.com/comprofix/gameskeleton/internal/gfx"
	"github.com/comprofix/gameskeleton/internal/input"
	"github.com/comprofix/gameskeleton/internal/scene"
)

// Saved is what was loaded from the saved files, and where they live
// (DESIGN §10). An empty Dir saves nothing.
type Saved struct {
	Settings config.Settings
	Dir      string
}

// Game is the top-level ebiten.Game.
type Game struct {
	preset   config.Preset
	screen   *ebiten.Image // the internal screen
	prescale *ebiten.Image // whole-number upscale used for CRT presentation
	outW     int
	outH     int

	ctx    *scene.Context
	scenes *scene.Manager

	// Tab shows the display test pattern instead of the scenes.
	showPattern bool
	pixels      *image.RGBA // CPU-side screen for the test pattern

	dir      string
	settings config.Settings
}

// New loads the assets and starts at the title screen with the given
// random seed and saved files.
func New(seed uint64, saved Saved) (*Game, error) {
	p := config.Project
	fonts, err := gfx.LoadFonts(assets.FS, "font/font8x8.txt")
	if err != nil {
		return nil, fmt.Errorf("fonts: %w", err)
	}
	demoSheet, err := gfx.LoadSheet(assets.FS, "gen/demo")
	if err != nil {
		return nil, fmt.Errorf("demo sheet: %w", err)
	}
	g := &Game{
		preset:   p,
		screen:   ebiten.NewImage(p.ScreenWidth, p.ScreenHeight),
		pixels:   image.NewRGBA(image.Rect(0, 0, p.ScreenWidth, p.ScreenHeight)),
		dir:      saved.Dir,
		settings: saved.Settings,
	}
	g.ctx = &scene.Context{
		Input:   input.New(&input.Ebiten{}),
		RNG:     rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)),
		Fonts:   fonts,
		Scheme:  gfx.SchemeFor(p.Palette),
		Sheets:  map[string]*gfx.Sheet{"demo": demoSheet},
		NewPlay: demo.NewPlay,
	}
	g.ctx.Settings = &g.settings
	g.scenes = scene.NewManager(scene.NewTitle())
	return g, nil
}

// Update handles the window keys and runs the scenes for one tick.
func (g *Game) Update() error {
	in := g.ctx.Input
	in.Update()
	s := &g.settings

	if in.JustPressed(input.Quit) || ebiten.IsWindowBeingClosed() {
		g.saveSettings() // keeps the window size
		return ebiten.Termination
	}
	if in.JustPressed(input.Fullscreen) {
		s.Fullscreen = !s.Fullscreen
		g.ctx.SettingsChanged()
	}
	if in.JustPressed(input.ScaleMode) {
		s.ScaleMode = g.preset.NextMode(s.ScaleMode)
		g.ctx.SettingsChanged()
	}
	if in.JustPressed(input.Sound) {
		s.SoundOn = !s.SoundOn
		g.ctx.SettingsChanged()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.showPattern = !g.showPattern
	}

	if err := g.scenes.Update(g.ctx); err != nil {
		return err
	}
	g.ctx.Tick++
	if g.ctx.Quitting() {
		g.saveSettings()
		return ebiten.Termination
	}
	if g.ctx.TakeSettingsChanged() {
		g.applySettings()
		g.saveSettings()
	}
	return nil
}

// applySettings makes the window match the settings.
func (g *Game) applySettings() {
	if ebiten.IsFullscreen() != g.settings.Fullscreen {
		ebiten.SetFullscreen(g.settings.Fullscreen)
	}
}

// saveSettings writes the current settings (DESIGN §10). A failure is
// logged, never fatal.
func (g *Game) saveSettings() {
	s := &g.settings
	if !ebiten.IsFullscreen() {
		if w, h := ebiten.WindowSize(); w > 0 && h > 0 {
			s.WindowW, s.WindowH = w, h
		}
	}
	if g.dir == "" {
		return
	}
	if err := config.SaveSettings(filepath.Join(g.dir, config.SettingsFile), *s); err != nil {
		log.Printf("saving settings: %v", err)
	}
}

// Draw renders the internal screen and presents it.
func (g *Game) Draw(out *ebiten.Image) {
	if g.showPattern {
		drawTestPattern(g.pixels, g.preset, g.ctx.Tick)
		g.screen.WritePixels(g.pixels.Pix)
	} else {
		g.scenes.Draw(g.ctx, g.screen)
	}
	g.present(out)
}

// present scales the internal screen onto the output surface in the
// current scale mode (DESIGN §5).
func (g *Game) present(out *ebiten.Image) {
	sw, sh := g.preset.ScreenWidth, g.preset.ScreenHeight
	out.Fill(color.Black)
	r := gfx.Fit(g.settings.ScaleMode, sw, sh, g.outW, g.outH)
	if g.settings.ScaleMode != config.ScaleCRT {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(r.W/float64(sw), r.H/float64(sh))
		op.GeoM.Translate(r.X, r.Y)
		op.Filter = ebiten.FilterNearest
		out.DrawImage(g.screen, op)
		return
	}

	// Upscale by a whole number with nearest-neighbour, then resample
	// smoothly to the 4:3 size so rows stay even and sharp.
	k := gfx.PrescaleFactor(r, sw, sh)
	w, h := sw*k, sh*k
	if g.prescale == nil || g.prescale.Bounds().Dx() != w || g.prescale.Bounds().Dy() != h {
		if g.prescale != nil {
			g.prescale.Deallocate()
		}
		g.prescale = ebiten.NewImage(w, h)
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(k), float64(k))
	op.Filter = ebiten.FilterNearest
	g.prescale.DrawImage(g.screen, op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(r.W/float64(w), r.H/float64(h))
	op.GeoM.Translate(r.X, r.Y)
	op.Filter = ebiten.FilterLinear
	out.DrawImage(g.prescale, op)
}

// LayoutF makes the output surface match the window in device pixels, so
// scaling is done by present rather than by Ebitengine.
func (g *Game) LayoutF(outsideWidth, outsideHeight float64) (float64, float64) {
	s := ebiten.Monitor().DeviceScaleFactor()
	w, h := outsideWidth*s, outsideHeight*s
	g.outW, g.outH = int(w), int(h)
	return w, h
}

// Layout is required by ebiten.Game; LayoutF takes precedence.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	panic("unreachable: LayoutF is used")
}
