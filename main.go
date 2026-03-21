package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/entro314-labs/cool-kit/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := cmd.Execute(ctx, version, commit, date, builtBy); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
