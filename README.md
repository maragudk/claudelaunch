# claudelaunch

[![CI](https://github.com/maragudk/claudelaunch/actions/workflows/ci.yml/badge.svg)](https://github.com/maragudk/claudelaunch/actions/workflows/ci.yml)

Claude Code launcher -- starts persistent Claude Code sessions inside tmux.

Made with sparkles by [maragu](https://www.maragu.dev/).

> **Disclaimer:** This project is 100% vibe-coded. No humans were harmed in the making of this software, but no humans reviewed it either.

## Usage

Start the server:

```shell
go run ./cmd/app
```

Then open http://localhost:6677 in your browser, type a session name, and hit Launch.

This creates a new tmux session running `claude --dangerously-skip-permissions` inside `~/Developer/<name>`. Attach with `tmux attach -t <name>`.
