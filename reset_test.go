package vt10x

import (
	"bytes"
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

func TestRISResetLeavesTerminalOnClearedMainScreen(t *testing.T) {
	term := New(WithSize(8, 3))

	if _, err := term.Write([]byte("main")); err != nil {
		t.Fatal(err)
	}
	if _, err := term.Write([]byte("\x1b[?1049h\x1b[2J\x1b[Halt")); err != nil {
		t.Fatal(err)
	}
	if _, err := term.Write([]byte("\x1bc")); err != nil {
		t.Fatal(err)
	}

	view := term.(View)
	if view.Mode()&ModeAltScreen != 0 {
		t.Fatal("expected RIS to leave alt screen mode")
	}
	for y := 0; y < 3; y++ {
		for x := 0; x < 8; x++ {
			if got := term.Cell(x, y).Char; got != ' ' {
				t.Fatalf("expected reset screen cell (%d,%d) to be blank, got %q", x, y, got)
			}
		}
	}
}

func TestCPRReportsRowRelativeToOriginMode(t *testing.T) {
	var out bytes.Buffer
	term := New(WithWriter(&out), WithSize(10, 6))

	if _, err := term.Write([]byte("\x1b[2;5r\x1b[?6h\x1b[3;4H\x1b[6n")); err != nil {
		t.Fatal(err)
	}

	if got := out.String(); got != "\x1b[3;4R" {
		t.Fatalf("expected relative CPR reply %q, got %q", "\x1b[3;4R", got)
	}
}

func TestPrivateCPRIncludesQuestionMarkPrefix(t *testing.T) {
	var out bytes.Buffer
	term := New(WithWriter(&out), WithSize(10, 6))

	if _, err := term.Write([]byte("\x1b[?6n")); err != nil {
		t.Fatal(err)
	}

	if got := out.String(); got != "\x1b[?1;1R" {
		t.Fatalf("expected private CPR reply %q, got %q", "\x1b[?1;1R", got)
	}
}
