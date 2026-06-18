package cmd

import (
	"fmt"
	"os"

	"github.com/adityarizkyramadhan/cloudflared-setup-cli/ui"
)

func Execute() {
	if err := ui.Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
