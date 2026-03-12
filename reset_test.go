package vt10x

import (
	"strings"
	"testing"
)

func TestRISResetClearsFullWidth(t *testing.T) {
	term := New(WithSize(10, 3))

	input := strings.Join([]string{
		"abcdefghij",
		"klmnopqrst",
		"uvwxyzABCD",
	}, "\r\n")
	if _, err := term.Write([]byte(input)); err != nil {
		t.Fatal(err)
	}
	if _, err := term.Write([]byte("\x1bc")); err != nil {
		t.Fatal(err)
	}

	for y := 0; y < 3; y++ {
		for x := 0; x < 10; x++ {
			if got := term.Cell(x, y).Char; got != ' ' {
				t.Fatalf("expected reset screen cell (%d,%d) to be blank, got %q", x, y, got)
			}
		}
	}
}
