package vt10x

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
)

// Terminal represents the virtual terminal emulator.
type Terminal interface {
	// View displays the virtual terminal.
	View

	// Write parses input and writes terminal changes to state.
	io.Writer

	// Parse blocks on read on pty or io.Reader, then parses sequences until
	// buffer empties. State is locked as soon as first rune is read, and unlocked
	// when buffer is empty.
	Parse(bf *bufio.Reader) error
}

// View represents the view of the virtual terminal emulator.
type View interface {
	// String dumps the virtual terminal contents.
	fmt.Stringer

	// Size returns the size of the virtual terminal.
	Size() (cols, rows int)

	// Resize changes the size of the virtual terminal.
	Resize(cols, rows int)

	// Mode returns the current terminal mode.//
	Mode() ModeFlag

	// Title represents the title of the console window.
	Title() string

	// Cell returns the glyph containing the character code, foreground color, and
	// background color at position (x, y) relative to the top left of the terminal.
	Cell(x, y int) Glyph

	// Cursor returns the current position of the cursor.
	Cursor() Cursor

	// CursorVisible returns the visible state of the cursor.
	CursorVisible() bool

	// Lock locks the state object's mutex.
	Lock()

	// Unlock resets change flags and unlocks the state object's mutex.
	Unlock()
}

type TerminalOption func(*TerminalInfo)

type TerminalInfo struct {
	w           io.Writer
	cols, rows  int
	onScrollUp  OnScrollUpFunc
}

// OnScrollUpFunc is invoked when one or more lines are about to be evicted
// from the top of the primary scroll region (i.e. content that is leaving
// the visible grid). lines holds a defensive copy of the evicted rows in
// top-to-bottom order; the slice and its elements are owned by the
// callback and may be retained. altScreen reports whether the eviction
// happened on the alternate screen — callers maintaining chat-style
// scrollback typically only want to retain primary-screen evictions, but
// for TUIs that live entirely in the alternate screen (e.g. Claude Code,
// Codex) alt-screen evictions are the only source of history.
//
// The callback runs while the terminal's internal mutex is held. It must
// not call any method on the Terminal/View or block on anything that
// could; do queued work or copy out and process asynchronously.
type OnScrollUpFunc func(lines [][]Glyph, altScreen bool)

func WithWriter(w io.Writer) TerminalOption {
	return func(info *TerminalInfo) {
		info.w = w
	}
}

func WithSize(cols, rows int) TerminalOption {
	return func(info *TerminalInfo) {
		info.cols = cols
		info.rows = rows
	}
}

// WithOnScrollUp registers a callback that fires when lines are about to
// be evicted from the top of the primary scroll region. Only full-region
// scrolls (orig == top) trigger it; sub-region scrolls produced by
// DECSTBM-bounded apps are intentionally ignored, since their "evicted"
// content remains visible elsewhere on screen. See OnScrollUpFunc for
// locking constraints.
func WithOnScrollUp(fn OnScrollUpFunc) TerminalOption {
	return func(info *TerminalInfo) {
		info.onScrollUp = fn
	}
}

// New returns a new virtual terminal emulator.
func New(opts ...TerminalOption) Terminal {
	info := TerminalInfo{
		w:    ioutil.Discard,
		cols: 80,
		rows: 24,
	}
	for _, opt := range opts {
		opt(&info)
	}
	return newTerminal(info)
}
