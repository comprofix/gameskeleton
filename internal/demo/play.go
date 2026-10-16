// Package demo is Dot Dash, the skeleton's demo game (DESIGN §3). Nothing
// outside it imports it except the game package, which hands NewPlay to the
// scenes; tools/newgame removes it.
package demo

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/gameskeleton/internal/config"
	"github.com/comprofix/gameskeleton/internal/gfx"
	"github.com/comprofix/gameskeleton/internal/input"
	"github.com/comprofix/gameskeleton/internal/scene"
	"github.com/comprofix/gameskeleton/internal/ui"
)

// Shake on Confirm, to show the screen shake.
const (
	shakeTicks = 12
	shakePx    = 3
)

// Play is the demo's play scene. For now it shows the sprites and fonts
// animating; Dot Dash's rules replace it.
type Play struct {
	skill config.Skill
	ticks int64
	bob   gfx.Animator
	glint gfx.Animator
	pulse gfx.Animator
}

// NewPlay starts the demo at a skill.
func NewPlay(_ *scene.Context, skill config.Skill) scene.Scene {
	return &Play{
		skill: skill,
		bob:   gfx.Animator{TicksPerFrame: 20, Frames: 2},
		glint: gfx.Animator{TicksPerFrame: 8, Frames: 4},
		pulse: gfx.Animator{TicksPerFrame: 15, Frames: 2},
	}
}

// Update implements scene.Scene.
func (p *Play) Update(ctx *scene.Context) error {
	in := ctx.Input
	switch {
	case in.JustPressed(input.Pause):
		ctx.Cut(scene.NewPause(p))
		return nil
	case in.JustPressed(input.Menu):
		ctx.Switch(scene.NewMainMenu())
		return nil
	case in.JustPressed(input.Restart):
		ctx.Switch(NewPlay(ctx, p.skill))
		return nil
	case in.JustPressed(input.Confirm):
		ctx.Shake.Start(shakeTicks, shakePx)
	}
	p.ticks++
	p.bob.Advance(1)
	p.glint.Advance(1)
	p.pulse.Advance(1)
	return nil
}

// Draw implements scene.Scene.
func (p *Play) Draw(ctx *scene.Context, dst *ebiten.Image) {
	sc, f := ctx.Scheme, ctx.Fonts
	w, h := config.Project.ScreenWidth, config.Project.ScreenHeight
	dst.Fill(gfx.RGBA(sc.Backdrop))

	ui.TextCentred(dst, f.Large, "DOT DASH", 12, sc.Title, sc.Shadow)
	ui.TextCentred(dst, f.Small, "Skill: "+p.skill.Name, 36, sc.Text, sc.Shadow)
	ui.TextCentred(dst, f.Small, "Esc pause  Space shake  Ctrl+M menu", h-2*f.Small.Size, sc.Dim, sc.Shadow)

	// A row of sprites: the player bobbing, mirrored, hurt, the hazard,
	// coins, hearts.
	sheet := ctx.Sheet("demo")
	y := 64.0
	player := sheet.Sprite("player")
	player.Draw(dst, p.bob.Frame(), 16, y, false)
	player.Draw(dst, p.bob.Frame(), 40, y, true)
	sheet.Sprite("player_hurt").Draw(dst, p.bob.Frame(), 64, y, false)
	sheet.Sprite("hazard").Draw(dst, p.pulse.Frame(), 88, y, false)
	dot := sheet.Sprite("dot")
	heart := sheet.Sprite("heart")
	for i := range 3 {
		dot.Draw(dst, p.glint.Frame()+i, float64(116+i*12), y+4, false)
		heart.Draw(dst, 0, float64(160+i*10), y+4, false)
	}

	// The player walks back and forth, mirrored when it heads left.
	span := int64(w - 32 - player.W)
	pos := p.ticks % (2 * span)
	flip := pos >= span
	if flip {
		pos = 2*span - pos
	}
	player.Draw(dst, int(p.ticks/10), float64(16+pos), float64(h-48), flip)
}
