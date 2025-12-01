# 🫧 bubble-jira

---

# Overview

A blazing fast, interactive Terminal User Interface (TUI) for Jira.  
Built with ❤️ in Go using Bubble Tea.

bubble-jira allows developers to view, manage, and transition Jira issues directly from the command line. No more context switching—manage your sprint without ever leaving your terminal.

---

# Features

- 🖥️ Interactive Dashboard: Browse issues assigned to you or query specific projects.
- 📝 Issue Details: View full descriptions, comments, and metadata instantly.
- 🔄 Seamless Transitions: Move tickets (To Do → In Progress → Done) with a keystroke.
- 🔍 Power Search: Filter issues using full Jira Query Language (JQL) support.
- ⚡ Lightweight: Built in Go for instant startup and minimal resource usage.

---

# Description

A Bubble Tea-based Jira TUI that fetches issues assigned to the current user and displays them in a searchable list.

Key actions:
- Arrow keys or j/k to move
- Enter to open the selected issue in the browser
- Tab to toggle between search and list
- F2 to open the config editor (edit config.json in ~/.config/bubble-jira/config.json)

Configuration is stored in `~/.config/bubble-jira/config.json`. Edit from the CLI or by hand.

---

# Start / Development

start (development): `go run .`

---

# Build / Distributable Package

To build a macOS installer / distributable package, use the included build script:

`./build-pkg.sh [--version X.X.X]`

The script produces an installer package for distribution. Supplying --version is optional but recommended to tag the build.

---

# Prerequisites

- Go: Version 1.21 or higher.
- Jira Account: A Jira Cloud account and an API Token.

---

# Installation

## Option 1: Build from Source

Clone the repository:

git clone https://github.com/CodeAstronomer/bubble-jira.git
cd bubble-jira

Build for distribution

For development you can run:

`go run .`

For producing a distributable package (recommended), run:

`./build-pkg.sh [--version X.X.X]`

If you only need a local binary you can still build with:

`go build -o bubble-jira`

---

# Configuration

Before running the application, you must configure your Jira credentials.
Tip: To generate an API token, visit Atlassian API Tokens.

---

# Usage

Simply run the application from your terminal:

`bubble-jira`

---

# Keybindings

Key — Action

- ↑ / k — Move selection up
- ↓ / j — Move selection down
- Enter — Open selected issue / Confirm action
- / — Filter issues (Search)
- t — Transition issue status
- Esc — Go back
- q — Quit application

---

# Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository.
2. Create your feature branch (git checkout -b feature/AmazingFeature).
3. Commit your changes (git commit -m 'Add some AmazingFeature').
4. Push to the branch (git push origin feature/AmazingFeature).
5. Open a Pull Request.

---

# ~/.zshrc Alias

Füge folgendes in deine `~/.zshrc` ein:

```bash
function jira() { (cd ~/bubble-jira && go run . "$@") }
```

---

# 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.