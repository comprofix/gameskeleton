package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/gameskeleton/internal/config"
	"github.com/comprofix/gameskeleton/internal/gfx"
	"github.com/comprofix/gameskeleton/internal/gfx/palettes"
)

// showcase draws the demo sprites and both fonts, animated, to check the
// sprite and font systems (DESIGN §6). Scenes replace it.
type showcase struct {
	sheet  *gfx.Sheet
	fonts  gfx.Fonts
	bob    gfx.Animator
	glint  gfx.Animator
	pulse  gfx.Animator
	colour struct{ back, ink, title, shadow int }
}

func newShowcase(sheet *gfx.Sheet, fonts gfx.Fonts, pal palettes.Palette) *showcase {
	s := &showcase{
		sheet: sheet,
		fonts: fonts,
		bob:   gfx.Animator{TicksPerFrame: 20, Frames: 2},
		glint: gfx.Animator{TicksPerFrame: 8, Frames: 4},
		pulse: gfx.Animator{TicksPerFrame: 15, Frames: 2},
	}
	// Named colours where the palette has them, so the showcase works with
	// any preset.
	s.colour.back = pick(pal, pal.Darkest(), "Navy")
	s.colour.ink = pal.Brightest()
	s.colour.title = pick(pal, pal.Brightest(), "Gold", "Yellow")
	s.colour.shadow = pal.Darkest()
	return s
}

// pick returns the first of names the palette has, or fallback.
func pick(pal palettes.Palette, fallback int, names ...string) int {
	for _, n := range names {
		if i, ok := pal.Lookup(n); ok {
			return i
		}
	}
	return fallback
}

func (s *showcase) update() {
	s.bob.Advance(1)
	s.glint.Advance(1)
	s.pulse.Advance(1)
}

func (s *showcase) draw(dst *ebiten.Image, tick int64) {
	w := config.Project.ScreenWidth
	dst.Fill(gfx.RGBA(s.colour.back))

	title := "GAME SKELETON"
	s.fonts.Large.DrawShadowed(dst, title, s.fonts.Large.CentreX(title), 12, s.colour.title, s.colour.shadow)
	lines := []string{
		"Large font 16x16 (Scale2x)",
		"Small font 8x8: 0123456789 !?",
		"Tab: test pattern   F10: quit",
	}
	for i, l := range lines {
		s.fonts.Small.DrawShadowed(dst, l, s.fonts.Small.CentreX(l), 40+i*12, s.colour.ink, s.colour.shadow)
	}

	// A row of sprites: the player bobbing, hurt, the hazard, coins, hearts.
	y := 90.0
	player := s.sheet.Sprite("player")
	player.Draw(dst, s.bob.Frame(), 16, y, false)
	player.Draw(dst, s.bob.Frame(), 40, y, true)
	s.sheet.Sprite("player_hurt").Draw(dst, s.bob.Frame(), 64, y, false)
	s.sheet.Sprite("hazard").Draw(dst, s.pulse.Frame(), 88, y, false)
	dot := s.sheet.Sprite("dot")
	heart := s.sheet.Sprite("heart")
	for i := range 3 {
		dot.Draw(dst, s.glint.Frame()+i, float64(116+i*12), y+4, false)
		heart.Draw(dst, 0, float64(160+i*10), y+4, false)
	}

	// The player walks back and forth along the bottom, mirrored when it
	// heads left.
	span := int64(w - 32 - player.W)
	pos := tick % (2 * span)
	flip := pos >= span
	if flip {
		pos = 2*span - pos
	}
	player.Draw(dst, int(tick/10), float64(16+pos), float64(config.Project.ScreenHeight-32), flip)
}
