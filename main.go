package main

import (
	"github.com/entro314-labs/cool-kit/cmd"
)

var (
	version = "1.0.0"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(version, commit, date)
}
