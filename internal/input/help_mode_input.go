package input

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Gaurav-Gosain/tuios/internal/app"
)

// handleHelpModeKey routes the keys help mode owns. Everything else is
// swallowed: a letter typed while the mode is on is not text for the shell, and
// a leader chord in it is not a second mode. The mode is read-only, so no key
// here changes the layout it explains; the rail chords the pane lists are
// things to press after leaving. Navigation and scrolling take a chord each,
// because the bare keys they used to take are things the user may be reading
// about, and h is the one bare letter help mode keeps.
func handleHelpModeKey(msg tea.KeyPressMsg, o *app.OS) (*app.OS, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c", "shift+f1":
		o.ExitHelpMode()
	case "shift+tab":
		o.HelpNext(1)
	case "shift+down":
		o.HelpDeeper()
	case "shift+up":
		o.HelpShallower()
	case "up":
		o.HelpScroll(-1)
	case "down":
		o.HelpScroll(1)
	case "h":
		o.HelpToggleKeys()
	}
	return o, nil
}
