package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/gameskeleton/internal/config"
	"github.com/comprofix/gameskeleton/internal/gfx"
)

// LineGap is the space between menu lines, in pixels.
const LineGap = 4

// Size returns the menu's width and height in font's pixels, including the
// cursor's column.
func (m *Menu) Size(font *gfx.Font) (w, h int) {
	for i := range m.Items {
		w = max(w, font.Width(m.Text(i)))
	}
	w += 2 * font.Size // the "> " cursor
	return w, len(m.Items)*(font.Size+LineGap) - LineGap
}

// Draw draws the menu with its top-left corner at (x, y): disabled items in
// sc.Dim, the selected item in sc.Highlight after a "> " cursor.
func (m *Menu) Draw(dst *ebiten.Image, font *gfx.Font, x, y int, sc gfx.Scheme) {
	for i := range m.Items {
		c := sc.Text
		switch {
		case m.Items[i].Disabled:
			c = sc.Dim
		case i == m.Cursor:
			c = sc.Highlight
			font.DrawShadowed(dst, ">", x, y, c, sc.Shadow)
		}
		font.DrawShadowed(dst, m.Text(i), x+2*font.Size, y, c, sc.Shadow)
		y += font.Size + LineGap
	}
}

// DrawCentred draws the menu centred across the screen with its top at y.
func (m *Menu) DrawCentred(dst *ebiten.Image, font *gfx.Font, y int, sc gfx.Scheme) {
	w, _ := m.Size(font)
	m.Draw(dst, font, (config.Project.ScreenWidth-w)/2, y, sc)
}

// Panel fills r with palette colour fill inside a one-pixel border.
func Panel(dst *ebiten.Image, r image.Rectangle, fill, border int) {
	x, y := float32(r.Min.X), float32(r.Min.Y)
	w, h := float32(r.Dx()), float32(r.Dy())
	vector.FillRect(dst, x, y, w, h, gfx.RGBA(border), false)
	vector.FillRect(dst, x+1, y+1, w-2, h-2, gfx.RGBA(fill), false)
}

// PanelAround is the rectangle of a panel with pad pixels of margin around
// a w×h block centred across the screen with its top at y.
func PanelAround(w, h, y, pad int) image.Rectangle {
	x := (config.Project.ScreenWidth - w) / 2
	return image.Rect(x-pad, y-pad, x+w+pad, y+h+pad)
}

// TextCentred draws s with a shadow, centred across the screen.
func TextCentred(dst *ebiten.Image, font *gfx.Font, s string, y, c, shadow int) {
	font.DrawShadowed(dst, s, font.CentreX(s), y, c, shadow)
}

// Blink is true for the first half of every period ticks, for blinking
// prompts.
func Blink(tick int64, period int64) bool {
	return tick%period < period/2
}
