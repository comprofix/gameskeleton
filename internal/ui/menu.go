// Package ui has the shared menu widget and drawing helpers for panels and
// text (DESIGN §7).
package ui

import (
	"github.com/comprofix/gameskeleton/internal/input"
)

// Item is one line of a Menu.
type Item struct {
	Label    string
	Value    func() string // optional: shown after the label, e.g. "On"
	Disabled bool
	Confirm  func()        // Confirm on this item
	Adjust   func(dir int) // Left (-1) or Right (+1) on this item
}

// Menu is a vertical list of items with a cursor. Up and Down move (with
// repeat, wrapping and skipping disabled items), Confirm activates, and
// Left/Right adjust values.
type Menu struct {
	Items  []Item
	Cursor int
	// OnMove and OnConfirm are called after the cursor moves or an item is
	// activated, for sounds.
	OnMove    func()
	OnConfirm func()
}

// Update handles one tick of input.
func (m *Menu) Update(in *input.Input) {
	if len(m.Items) == 0 {
		return
	}
	m.skipDisabled(1)
	switch {
	case in.Repeat(input.Up):
		m.move(-1)
	case in.Repeat(input.Down):
		m.move(1)
	case in.JustPressed(input.Confirm):
		it := m.Items[m.Cursor]
		if it.Disabled {
			return
		}
		if it.Confirm != nil {
			it.Confirm()
		} else if it.Adjust != nil {
			it.Adjust(1)
		}
		if m.OnConfirm != nil {
			m.OnConfirm()
		}
	case in.Repeat(input.Left):
		m.adjust(-1)
	case in.Repeat(input.Right):
		m.adjust(1)
	}
}

func (m *Menu) move(dir int) {
	start := m.Cursor
	for range m.Items {
		m.Cursor = (m.Cursor + dir + len(m.Items)) % len(m.Items)
		if !m.Items[m.Cursor].Disabled {
			break
		}
	}
	if m.Cursor != start && m.OnMove != nil {
		m.OnMove()
	}
}

func (m *Menu) adjust(dir int) {
	it := m.Items[m.Cursor]
	if it.Disabled || it.Adjust == nil {
		return
	}
	it.Adjust(dir)
	if m.OnMove != nil {
		m.OnMove()
	}
}

// skipDisabled moves the cursor off a disabled item, if it is on one.
func (m *Menu) skipDisabled(dir int) {
	m.Cursor = min(max(m.Cursor, 0), len(m.Items)-1)
	if m.Items[m.Cursor].Disabled {
		m.move(dir)
	}
}

// Text returns item i's text as drawn: the label, then its value if any.
func (m *Menu) Text(i int) string {
	it := m.Items[i]
	if it.Value == nil {
		return it.Label
	}
	return it.Label + ": " + it.Value()
}

// OnOff is a Value for a boolean setting.
func OnOff(v *bool) func() string {
	return func() string {
		if *v {
			return "On"
		}
		return "Off"
	}
}
