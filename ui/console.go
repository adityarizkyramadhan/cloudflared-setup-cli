// Package ui implements a plain, interactive numbered-menu CLI over
// stdin/stdout. It replaces the former Bubbletea TUI: every screen is a simple
// menu loop that prints with fmt and reads lines from the user.
package ui

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Console holds the shared input/output for the interactive session.
type Console struct {
	r   *bufio.Reader
	out io.Writer
	eof bool // set once input reaches EOF (Ctrl+D / end of piped input)
}

// Run starts the interactive CLI, reading from in and writing to out. It
// returns when the user exits or input reaches EOF.
func Run(in io.Reader, out io.Writer) error {
	c := &Console{r: bufio.NewReader(in), out: out}
	c.mainMenu()
	return nil
}

func (c *Console) println(a ...any)            { fmt.Fprintln(c.out, a...) }
func (c *Console) printf(f string, a ...any)   { fmt.Fprintf(c.out, f, a...) }
func (c *Console) info(s string)               { fmt.Fprintln(c.out, s) }
func (c *Console) ok(s string)                 { fmt.Fprintln(c.out, "✓ "+s) }
func (c *Console) fail(s string)               { fmt.Fprintln(c.out, "✗ "+s) }
func (c *Console) warn(s string)               { fmt.Fprintln(c.out, "⚠ "+s) }

// readLine reads one trimmed line. After EOF it sets eof and returns "".
func (c *Console) readLine() string {
	if c.eof {
		return ""
	}
	line, err := c.r.ReadString('\n')
	if err != nil {
		c.eof = true
	}
	return strings.TrimSpace(line)
}

// prompt prints label and reads a trimmed line.
func (c *Console) prompt(label string) string {
	fmt.Fprint(c.out, label)
	return c.readLine()
}

// promptDefault returns def when the user enters a blank line.
func (c *Console) promptDefault(label, def string) string {
	if v := c.prompt(label); v != "" {
		return v
	}
	return def
}

// confirm asks a yes/no question; anything other than y/yes is false.
func (c *Console) confirm(label string) bool {
	v := strings.ToLower(c.prompt(label + " [y/N]: "))
	return v == "y" || v == "yes"
}

// readChoice reads a menu selection.
func (c *Console) readChoice() string { return c.prompt("> ") }
