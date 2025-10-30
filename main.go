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
	// Define CLI flags
	helpFlag := flag.Bool("help", false, "Show help message")
    flag.BoolVar(helpFlag, "h", false, "Show help message (shorthand)")

	versionFlag := flag.Bool("v", false, "Show version")
	taskFlag := flag.Bool("t", false, "Directly start in Task List View")

	flag.Parse()

	ui.Init()

    // Handle help/version flags BEFORE running the program
    if *helpFlag {
        fmt.Println(ui.Strings["HelpHeader"])
        fmt.Println(ui.Strings["HelpUsage"] + ": jira [flags]")
        fmt.Println(ui.Strings["HelpFlags"] + ":")
        fmt.Println("  --help/ -h    " + ui.Strings["HelpFlagHelp"])
        fmt.Println("  -v           " + ui.Strings["HelpFlagVersion"])
        fmt.Println("  -t           " + ui.Strings["HelpFlagTasks"])
        return
    }

    if *versionFlag {
        fmt.Println("jira version", ui.AppVersion)
        return
    }

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