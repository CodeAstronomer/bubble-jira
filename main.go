package main

import (
	"fmt"
	"os"
	"flag"

	"bubble-jira/config"
	"bubble-jira/jira"
	"bubble-jira/ui"
)

func main() {
	// Parse CLI flags
	taskFlag := flag.Bool("t", false, "Directly start in Task List View")
	flag.Parse()

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

    if *taskFlag {
    	if err := pr.StartWithTasks(cfg, jc); err != nil {
    		fmt.Println("Error running program:", err)
    		os.Exit(1)
    	}
    	return
    }

	// Start TUI
	if err := pr.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}