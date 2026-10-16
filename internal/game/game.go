// Package game implements ebiten.Game: it handles the window keys, runs the
// game at the project preset's internal resolution, and presents it scaled
// to the window (DESIGN §5).
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
	"github.com/comprofix/gameskeleton/internal/gfx"
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
	pixels   *image.RGBA   // CPU-side copy of the screen for the test pattern
	prescale *ebiten.Image // whole-number upscale used for CRT presentation
	outW     int
	outH     int

	rng  *rand.Rand
	tick int64

	showcase    *showcase
	showPattern bool // Tab switches between the showcase and the test pattern

	dir      string
	settings config.Settings
}

// New loads the assets and starts the game with the given random seed and
// saved files.
func New(seed uint64, saved Saved) (*Game, error) {
	p := config.Project
	fonts, err := gfx.LoadFonts(assets.FS, "font/font8x8.txt")
	if err != nil {
		return nil, fmt.Errorf("fonts: %w", err)
	}
	demo, err := gfx.LoadSheet(assets.FS, "gen/demo")
	if err != nil {
		return nil, fmt.Errorf("demo sheet: %w", err)
	}
	return &Game{
		preset:   p,
		screen:   ebiten.NewImage(p.ScreenWidth, p.ScreenHeight),
		pixels:   image.NewRGBA(image.Rect(0, 0, p.ScreenWidth, p.ScreenHeight)),
		rng:      rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)),
		showcase: newShowcase(demo, fonts, p.Palette),
		dir:      saved.Dir,
		settings: saved.Settings,
	}, nil
}

// Update handles the window keys and advances the game by one tick.
func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF10) || ebiten.IsWindowBeingClosed() {
		g.saveSettings() // keeps the window size
		return ebiten.Termination
	}
	altEnter := ebiten.IsKeyPressed(ebiten.KeyAlt) && inpututil.IsKeyJustPressed(ebiten.KeyEnter)
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) || altEnter {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
		g.saveSettings()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF9) {
		g.settings.ScaleMode = g.preset.NextMode(g.settings.ScaleMode)
		g.saveSettings()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.showPattern = !g.showPattern
	}
	g.showcase.update()
	g.tick++
	return nil
}

// saveSettings writes the current settings (DESIGN §10). A failure is
// logged, never fatal.
func (g *Game) saveSettings() {
	s := &g.settings
	s.Fullscreen = ebiten.IsFullscreen()
	if !s.Fullscreen {
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
		drawTestPattern(g.pixels, g.preset, g.tick)
		g.screen.WritePixels(g.pixels.Pix)
	} else {
		g.showcase.draw(g.screen, g.tick)
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
