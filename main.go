package main

import (
	"fmt"
	"os"

	"bubble-jira/config"
	"bubble-jira/jira"
	"bubble-jira/ui"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	// Create Jira client
	jc := jira.NewClient(cfg)

	// Create Bubble Tea program
	pr := ui.NewProgram(cfg, jc)

	// Start TUI
	if err := pr.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}