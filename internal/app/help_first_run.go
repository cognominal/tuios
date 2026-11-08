package app

import (
	"os"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/adrg/xdg"
)

// The dock's help control is the one piece of chrome nobody is looking for: it
// is a single glyph in a bar of values. A first run blinks it for a few
// seconds, then writes a marker so it never does again. An introduction, not a
// permanent animation, and the marker is written when the blink ends rather
// than when it starts, so a session cut short has still taught the control on
// the next start.

const (
	helpBlinkFor  = 6 * time.Second
	helpBlinkStep = 400 * time.Millisecond
)

const helpBlinkSeenFile = "help-control-seen"

// helpBlinkSeenPath is where the marker lives. A variable so a test can point
// it at a scratch directory.
var helpBlinkSeenPath = func() string {
	return filepath.Join(xdg.StateHome, "tuios", helpBlinkSeenFile)
}

// helpBlinkMsg advances the first-run blink by one half.
type helpBlinkMsg struct{}

// loadHelpBlink starts the blink on a machine that has not seen it yet.
func (m *OS) loadHelpBlink() {
	if _, err := os.Stat(helpBlinkSeenPath()); err == nil {
		return
	}
	m.helpBlinkUntil = time.Now().Add(helpBlinkFor)
	m.helpBlinkOn = true
	m.helpInvalidate()
}

// helpBlinkTick arms the next half of the blink, when one is running.
func (m *OS) helpBlinkTick() tea.Cmd {
	if m.helpBlinkUntil.IsZero() || m.helpBlinkTicking {
		return nil
	}
	m.helpBlinkTicking = true
	return tea.Tick(helpBlinkStep, func(time.Time) tea.Msg { return helpBlinkMsg{} })
}

// advanceHelpBlink toggles the lit half, or ends the blink and writes the
// marker that keeps it from coming back. A marker that cannot be written is
// no reason to fail a frame: the control blinks again next start.
func (m *OS) advanceHelpBlink() {
	m.helpBlinkTicking = false
	if time.Now().After(m.helpBlinkUntil) {
		m.helpBlinkUntil = time.Time{}
		m.helpBlinkOn = false
		m.markHelpBlinkSeen()
		m.helpInvalidate()
		return
	}
	m.helpBlinkOn = !m.helpBlinkOn
	m.helpInvalidate()
}

// markHelpBlinkSeen records that the control has been shown. Any failure means
// the next start shows it again, which is the harmless direction.
func (m *OS) markHelpBlinkSeen() {
	path := helpBlinkSeenPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return
	}
	_ = os.WriteFile(path, []byte(time.Now().Format(time.RFC3339)+"\n"), 0o600)
}
