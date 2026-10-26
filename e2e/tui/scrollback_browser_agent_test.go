package tuie2e

import (
	"strings"
	"testing"

	"github.com/Gaurav-Gosain/tuitest"
)

// agentTurns prints two turns the way an agent TUI submits them.
//
// omp marks the input a person submits as an OSC 133 zone of its own: the A
// mark opens on the card's first row, the B mark closes on its last, and C and
// D are emitted at that same cell so the zone closes without latching the
// terminal's cursor-to-input state. Nothing sits between B and C, which a
// shell's marks always have (the typed line), so the command a shell would be
// read from is empty here. Every row of the card carries one column of inset,
// like omp's.
//
// The second turn is printed twice at the same rows, the way omp's row diff
// repaints every row it has drawn when the viewport top moves. That re-emits
// its marks.
const agentTurns = `printf '\033]133;A\007            \r\n SUBMITTEDONE turn one\r\n            \033]133;B\007\033]133;C\007\033]133;D;0\007\r\n'; printf 'ANSWERONE the reply to turn one\r\n'; printf '\033]133;A\007            \r\n SUBMITTEDTWO turn two\r\n            \033]133;B\007\033]133;C\007\033]133;D;0\007\r\n'; printf '\033[3A'; printf '\033]133;A\007            \r\n SUBMITTEDTWO turn two\r\n            \033]133;B\007\033]133;C\007\033]133;D;0\007\r\n'; printf 'ANSWER%s\r\n' TWO`

// TestScrollbackBrowserSegmentsAnAgentTUIsTurns opens the scrollback browser on
// a pane whose OSC 133 marks are an agent TUI's, not a shell's.
//
// How this could pass wrongly, written down first:
//   - The browser could open from the regex fallback, which matches prompt
//     shaped lines and would list something that is not a turn. The header's
//     parse method is read: it must say osc133, which only marks produce.
//   - The turns could be listed twice, once per emission. omp repaints rows it
//     has drawn, so the second turn's marks arrive twice at one row; moving to
//     the entry after the newest must reach the *older turn*, which cannot
//     happen if a repeat of the newest sits there.
//   - The listed turn could be the card's padding rather than the input. The
//     command copied out of it is compared whole.
//   - The output pane could be empty, which is the whole point of the browser.
//     The reply to the older turn must come back on the clipboard.
//   - The pane's own text could satisfy the assertions. Every check reads a
//     clipboard write or the browser's own header, never the screen text.
//
// Negative control: on origin/main the marks produce no blocks, the browser
// falls back to prompt matching, and its header reads (regex, 12 markers) over
// a list holding the pane's echoed command line as a turn. Recorded in
// NEGATIVE_CONTROLS.md.
func TestScrollbackBrowserSegmentsAnAgentTUIsTurns(t *testing.T) {
	out := &lockedBuffer{}
	term, _ := start(t, startOpts{out: out})
	waitBoot(t, term)
	newWindow(t, term)
	enterTerminalMode(t, term)

	runInShell(t, term, agentTurns, "ANSWERTWO", shellTimeout)

	if err := term.SendKeys(tuitest.Ctrl('b'), "s"); err != nil {
		t.Fatalf("send the prefix and s: %v", err)
	}
	if err := term.WaitForText("Scrollback browser", uiTimeout); err != nil {
		t.Fatalf("the browser did not open on a pane whose turns carry an agent TUI's marks: %v\n%s",
			err, term.Snapshot())
	}
	// The parse method is the browser's own header, so this cannot be the
	// pane's text and cannot be satisfied by the regex fallback.
	if got := term.Screen().Text(); !strings.Contains(got, "(osc133") {
		t.Fatalf("the browser did not cut the turns from their marks\n%s", term.Snapshot())
	}

	// The newest turn is selected. Its command is the input's text, not the
	// card's padding and not the empty run between B and C.
	from := len(clipboardWrites(out))
	if err := term.SendKeys("c"); err != nil {
		t.Fatalf("copy the newest turn's command: %v", err)
	}
	if got, want := pipeWrite(t, term, out, from, "c on the newest turn"), "SUBMITTEDTWO turn two"; got != want {
		t.Fatalf("the newest turn's command copied as %q, want %q\n%s", got, want, term.Snapshot())
	}

	// One entry older is the first turn, not a repeat of the newest.
	from = len(clipboardWrites(out))
	if err := term.SendKeys("j"); err != nil {
		t.Fatalf("step to the older turn: %v", err)
	}
	if err := term.SendKeys("c"); err != nil {
		t.Fatalf("copy the older turn's command: %v", err)
	}
	if got, want := pipeWrite(t, term, out, from, "c on the older turn"), "SUBMITTEDONE turn one"; got != want {
		t.Fatalf("the entry after the newest copied as %q, want the first turn %q; a repeat of the "+
			"newest turn was listed beside it\n%s", got, want, term.Snapshot())
	}

	// And that turn's output is the reply that followed it, which is what the
	// browser is for.
	from = len(clipboardWrites(out))
	if err := term.SendKeys("y"); err != nil {
		t.Fatalf("copy the older turn's output: %v", err)
	}
	if got, want := pipeWrite(t, term, out, from, "y on the older turn"), "ANSWERONE the reply to turn one"; got != want {
		t.Fatalf("the first turn's output copied as %q, want %q\n%s", got, want, term.Snapshot())
	}
	alive(t, term, "after reading the turns from the browser")
}
