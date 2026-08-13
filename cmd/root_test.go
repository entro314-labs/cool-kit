package cmd

import "testing"

func TestRootCommandHasUniqueSubcommands(t *testing.T) {
	seen := make(map[string]struct{})
	for _, sub := range rootCmd.Commands() {
		name := sub.Name()
		if _, exists := seen[name]; exists {
			t.Fatalf("duplicate root subcommand detected: %s", name)
		}
		seen[name] = struct{}{}
	}

	for _, required := range []string{"deploy", "instances", "update", "destroy"} {
		if _, exists := seen[required]; !exists {
			t.Fatalf("missing required root subcommand: %s", required)
		}
	}
}
