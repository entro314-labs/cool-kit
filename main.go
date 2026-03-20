package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/entro314-labs/cool-kit/cmd"
)

var (
	version = "1.0.2"
	commit  = "none"
	date    = "unknown"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := cmd.Execute(ctx, version, commit, date); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
