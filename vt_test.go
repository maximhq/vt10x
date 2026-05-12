package vt10x

import (
	"io"
	"strings"
	"testing"
)

func extractStr(term Terminal, x0, x1, row int) string {
	var s []rune
	for i := x0; i <= x1; i++ {
		attr := term.Cell(i, row)
		s = append(s, attr.Char)
	}
	return string(s)
}

func TestPlainChars(t *testing.T) {
	term := New()
	expected := "Hello world!"
	_, err := term.Write([]byte(expected))
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	actual := extractStr(term, 0, len(expected)-1, 0)
	if expected != actual {
		t.Fatal(actual)
	}
}

func TestNewline(t *testing.T) {
	term := New()
	expected := "Hello world!\n...and more."
	_, err := term.Write([]byte("\033[20h")) // set CRLF mode
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	_, err = term.Write([]byte(expected))
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}

	split := strings.Split(expected, "\n")
	actual := extractStr(term, 0, len(split[0])-1, 0)
	actual += "\n"
	actual += extractStr(term, 0, len(split[1])-1, 1)
	if expected != actual {
		t.Fatal(actual)
	}

	// A newline with a color set should not make the next line that color,
	// which used to happen if it caused a scroll event.
	st := (term.(*terminal))
	st.moveTo(0, st.rows-1)
	_, err = term.Write([]byte("\033[1;37m\n$ \033[m"))
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	cur := term.Cursor()
	attr := term.Cell(cur.X, cur.Y)
	if attr.FG != DefaultFG {
		t.Fatal(st.cur.X, st.cur.Y, attr.FG, attr.BG)
	}
}

// TestOnScrollUpEvictsTopLine ensures the WithOnScrollUp hook fires when
// LF at the bottom row pushes the top line off the grid, with the evicted
// content delivered before the in-place clear.
func TestOnScrollUpEvictsTopLine(t *testing.T) {
	var evicted [][]Glyph
	var sawAlt []bool
	term := New(
		WithSize(20, 3),
		WithOnScrollUp(func(lines [][]Glyph, altScreen bool) {
			for _, ln := range lines {
				cp := make([]Glyph, len(ln))
				copy(cp, ln)
				evicted = append(evicted, cp)
				sawAlt = append(sawAlt, altScreen)
			}
		}),
	)

	// Three lines fill the 3-row grid; the 4th LF should evict "line-1".
	_, err := term.Write([]byte("\033[20h")) // CRLF mode
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if _, err := term.Write([]byte("line-1\nline-2\nline-3\nline-4")); err != nil && err != io.EOF {
		t.Fatal(err)
	}

	if len(evicted) != 1 {
		t.Fatalf("expected exactly 1 evicted line, got %d", len(evicted))
	}
	got := glyphsToString(evicted[0])
	if !strings.HasPrefix(got, "line-1") {
		t.Fatalf("expected evicted line to start with %q, got %q", "line-1", got)
	}
	if sawAlt[0] {
		t.Fatalf("expected altScreen=false on primary screen, got true")
	}
}

// TestOnScrollUpReportsAltScreen ensures the alt-screen flag is true for
// evictions that happen while the alternate screen is active.
func TestOnScrollUpReportsAltScreen(t *testing.T) {
	var sawAlt []bool
	term := New(
		WithSize(10, 2),
		WithOnScrollUp(func(lines [][]Glyph, altScreen bool) {
			for range lines {
				sawAlt = append(sawAlt, altScreen)
			}
		}),
	)

	// Enter alt screen, force several LFs at the bottom row.
	if _, err := term.Write([]byte("\033[20h\033[?1049h")); err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if _, err := term.Write([]byte("a\nb\nc\nd")); err != nil && err != io.EOF {
		t.Fatal(err)
	}

	if len(sawAlt) == 0 {
		t.Fatalf("expected at least one alt-screen eviction, got none")
	}
	for i, alt := range sawAlt {
		if !alt {
			t.Fatalf("expected altScreen=true for eviction %d, got false", i)
		}
	}
}

// TestOnScrollUpIgnoresSubRegionScrolls ensures DECSTBM-bounded sub-region
// scrolls don't fire the hook, since content stays visible on screen.
func TestOnScrollUpIgnoresSubRegionScrolls(t *testing.T) {
	var count int
	term := New(
		WithSize(10, 6),
		WithOnScrollUp(func(lines [][]Glyph, altScreen bool) {
			count += len(lines)
		}),
	)

	// Restrict scroll region to rows 2-4 (1-based), then write enough LFs
	// to scroll inside that region. orig will be 1 (0-based row 2), not
	// t.top (0), so the hook must stay silent.
	if _, err := term.Write([]byte("\033[20h\033[2;4r")); err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if _, err := term.Write([]byte("\033[2;1H")); err != nil && err != io.EOF { // park cursor at row 2
		t.Fatal(err)
	}
	if _, err := term.Write([]byte("a\nb\nc\nd\ne")); err != nil && err != io.EOF {
		t.Fatal(err)
	}

	if count != 0 {
		t.Fatalf("expected no evictions from sub-region scrolls, got %d", count)
	}
}

// glyphsToString turns a row's glyphs into a trimmed string for assertions.
func glyphsToString(row []Glyph) string {
	var b strings.Builder
	for _, g := range row {
		if g.Char == 0 {
			b.WriteByte(' ')
		} else {
			b.WriteRune(g.Char)
		}
	}
	return strings.TrimRight(b.String(), " ")
}
