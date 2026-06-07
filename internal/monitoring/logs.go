package monitoring

import (
	"bufio"
	"io"
	"os/exec"
)

type LogStreamer struct {
	cmd     *exec.Cmd
	scanner *bufio.Scanner
	stdout  io.ReadCloser
}

func NewLogStreamer(tunnelName string) (*LogStreamer, error) {
	cmd := exec.Command("cloudflared", "tunnel", "--loglevel", "info", "run", tunnelName)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &LogStreamer{
		cmd:     cmd,
		scanner: bufio.NewScanner(stdout),
		stdout:  stdout,
	}, nil
}

func (ls *LogStreamer) NextLine() (string, error) {
	if ls.scanner.Scan() {
		return ls.scanner.Text(), nil
	}
	if err := ls.scanner.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}

func (ls *LogStreamer) Stop() error {
	if ls.cmd.Process != nil {
		return ls.cmd.Process.Kill()
	}
	return nil
}
