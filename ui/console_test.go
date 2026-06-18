package ui

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

func newTestConsole(input string) *Console {
	return &Console{r: bufio.NewReader(strings.NewReader(input)), out: io.Discard}
}

func TestConfirm(t *testing.T) {
	cases := map[string]bool{
		"y\n":    true,
		"yes\n":  true,
		"Y\n":    true,
		"YES\n":  true,
		"n\n":    false,
		"\n":     false,
		"nope\n": false,
	}
	for in, want := range cases {
		if got := newTestConsole(in).confirm("ok?"); got != want {
			t.Errorf("confirm(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestPromptDefault(t *testing.T) {
	if got := newTestConsole("\n").promptDefault("x", "def"); got != "def" {
		t.Errorf("blank input: got %q, want %q", got, "def")
	}
	if got := newTestConsole("val\n").promptDefault("x", "def"); got != "val" {
		t.Errorf("value input: got %q, want %q", got, "val")
	}
}

func TestReadChoiceTrims(t *testing.T) {
	if got := newTestConsole("  3  \n").readChoice(); got != "3" {
		t.Errorf("readChoice = %q, want %q", got, "3")
	}
}

func TestReadLineSetsEOF(t *testing.T) {
	c := newTestConsole("")
	if got := c.readLine(); got != "" {
		t.Errorf("readLine on empty = %q, want empty", got)
	}
	if !c.eof {
		t.Error("expected eof to be set after empty input")
	}
}

func TestReadLineFinalUnterminatedLine(t *testing.T) {
	c := newTestConsole("2") // no trailing newline
	if got := c.readLine(); got != "2" {
		t.Errorf("readLine = %q, want %q", got, "2")
	}
	if !c.eof {
		t.Error("expected eof after consuming final unterminated line")
	}
	if got := c.readLine(); got != "" {
		t.Errorf("readLine after eof = %q, want empty", got)
	}
}
