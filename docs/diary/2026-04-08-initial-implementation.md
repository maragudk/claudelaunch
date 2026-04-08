# Diary: Initial implementation of claudelaunch

Build an HTTP server that launches persistent Claude Code sessions inside tmux, with a web UI for triggering launches.

## Step 1: Scaffold the app from template

### Prompt Context

**Verbatim prompt:** You're in a template repo for a new app. Set up the app to be called "claudelaunch" instead of "template". This app will do one thing: Start an HTTP server on a predefined port. When it gets a POST request on that port, check the body for a name. That name is the name for a new session for claude code, which should be started in tmux. Essentially, it's a claude code launcher for a persistent claude code process inside tmux.
**Interpretation:** Rename the Go template repo and implement an HTTP server that spawns tmux sessions running claude.
**Inferred intent:** A lightweight launcher to spin up isolated, persistent Claude Code sessions from a browser or curl.

### What I did

- Renamed module from `maragu.dev/template` to `maragu.dev/claudelaunch` in `/go.mod`
- Deleted `/template.go`, created `/claudelaunch.go` with `Server` struct, POST handler, tmux session launching, and name validation
- Created `/cmd/app/main.go` (later renamed to `/cmd/claudelaunch/main.go`) with HTTP server on port 6677 (configurable via `PORT` env)
- Created `/claudelaunch_test.go` with handler tests
- Updated `/README.md`, `/Makefile`, `/.gitignore`

### Why

Starting from the template repo meant renaming everything first, then building the core feature: receive a name, validate it, create a tmux session running claude.

### What worked

The initial JSON POST API worked immediately. Tests passed on first run.

### What didn't work

Nothing notable at this stage.

### What I learned

Nothing surprising -- straightforward scaffolding.

### What was tricky

Nothing at this stage.

### What warrants review

The name validation regex and tmux command construction are the security-sensitive parts.

### Future work

None identified at this step.

## Step 2: Add HTML form UI with gomponents

### Prompt Context

**Verbatim prompt:** Actually, also add a GET handler with a simple HTML page and a form taking the name. Use your gomponents skill and the TailwindCSS CDN.
**Interpretation:** Add a browser-friendly UI alongside the API.
**Inferred intent:** Make it easy to launch sessions from a browser without needing curl.

### What I did

- Created `/html/html.go` with gomponents-based pages: `IndexPage`, `SuccessPage`, `ErrorPage`, sharing a `page` layout helper
- Added GET `/` handler serving the form, updated POST `/` to detect form submissions and return HTML instead of JSON
- Added `/.golangci.yml` with dot-import whitelist for gomponents (copied pattern from `../app`)
- Later dropped JSON POST entirely per user request, simplifying to form-only

### Why

A web form is more accessible than curl for quick session launches.

### What worked

gomponents + Tailwind CDN made the UI fast to build. The dark theme looks good. Dropping JSON simplified the code significantly -- `r.FormValue("name")` replaced all the content-type detection and JSON parsing.

### What didn't work

The linter initially complained about gomponents dot imports. Started with `//nolint:staticcheck` directives, but checking `../app/.golangci.yml` revealed the proper `dot-import-whitelist` config.

### What I learned

golangci-lint v2 uses `dot-import-whitelist` under `linters.settings.staticcheck` to suppress ST1001 for specific packages.

### What was tricky

Nothing notable.

### What warrants review

The HTML pages in `/html/html.go` -- check that the form validation pattern matches the server-side regex.

### Future work

None identified.

## Step 3: Working directory and session uniqueness

### Prompt Context

**Verbatim prompt:** The name should also be the working directory for claude code inside the session, as a subdir in ~/Developer.
**Interpretation:** Map the session name to `~/Developer/<name>` and start claude there.
**Inferred intent:** Each session should operate in its own project directory.

### What I did

- Added `os.MkdirAll` to create `~/Developer/<name>` if it doesn't exist
- Added unix timestamp suffix to session names for uniqueness (e.g. `app-1775645606`)
- Added `filepath.IsLocal` check to prevent `..` traversal, plus explicit `name == "."` check
- Allowed underscores and dots in names (regex: `^[a-zA-Z0-9._-]+$`)

### Why

Each claude session needs its own project directory. Timestamp suffixes mean you can launch multiple sessions for the same project without conflicts.

### What worked

`filepath.IsLocal` from stdlib handles the `..` case. The regex prevents slashes, so multi-segment traversal is impossible.

### What didn't work

`filepath.IsLocal(".")` returns `true` (it's a valid local path), so `.` had to be rejected explicitly.

### What I learned

`filepath.IsLocal` considers `.` local but `..` not local. Makes sense from a path perspective, but for our use case both are invalid names.

### What was tricky

Nothing notable.

### What warrants review

The path validation in `/claudelaunch.go` -- the combination of regex + `filepath.IsLocal` + explicit `.` check.

### Future work

None identified.

## Step 4: Getting tmux working directory right

### Prompt Context

**Verbatim prompt:** Try it out again with this repo name / That didn't work, claude launched in ~/Developer/claudelaunch
**Interpretation:** The tmux session's working directory wasn't being set correctly when launched from the server.
**Inferred intent:** Claude must start in the correct project directory.

### What I did

Went through several iterations:
1. Used `tmux new-session -c <dir>` -- didn't work
2. Added `cmd.Dir = dir` on exec.Command -- didn't work
3. Wrapped in `/bin/sh -c "cd <dir> && exec claude ..."` -- didn't work when args were passed separately to exec.Command
4. Tried `env -i` to strip inherited environment -- worked in isolation but not from the server
5. Discovered the root cause: `exec.Command("tmux", ..., "/bin/sh", "-c", script)` passes `/bin/sh`, `-c`, and `script` as separate tmux arguments, not as a single shell invocation
6. Fix: pass the entire command as a single string argument to tmux: `exec.Command("tmux", "new-session", "-d", "-s", session, "cd <dir> && claude --dangerously-skip-permissions")`

### Why

tmux runs the "shell command" argument through its default shell. When given as separate argv entries via exec.Command, tmux treats only the first as the command and ignores the rest as tmux args.

### What worked

Passing the whole `cd <dir> && claude ...` as a single string argument to tmux. This matches how tmux works on the command line.

### What didn't work

- `tmux -c <dir>`: sets the tmux session directory but claude still picked up the parent process's cwd
- `cmd.Dir`: only affects the tmux client process, not what the tmux server spawns
- `cmd.Env` / `cleanEnv`: only affects the tmux client, not the server-spawned shell (tmux is client-server)
- Separate `/bin/sh`, `-c`, `script` args: tmux doesn't interpret these as a shell invocation

The debugging was extensive. Key insight came from testing with `pwd > /tmp/debug.txt` in the script -- the file was never created, proving the script wasn't executing as written.

### What I learned

- tmux is client-server: env vars and cwd set on the `exec.Command` only affect the client, not the spawned process
- tmux's `new-session` takes a "shell-command" as a single string, which it runs through the default shell. Multiple separate args after the session options are treated differently
- When running claude from within another claude session, `pane_current_path` reflected the parent's cwd, not the child's, until the command string approach fixed it

### What was tricky

This was the trickiest part of the whole implementation. The failure mode was subtle: everything appeared to work (session created, claude running) but claude was in the wrong directory. Debugging required progressively isolating variables -- standalone Go program vs server, with/without claude, checking tmux pane metadata vs actual process state.

### What warrants review

The tmux command construction in `launchSession` in `/claudelaunch.go`. The single-string approach works but relies on the name being safe for shell interpolation (guaranteed by the regex validation).

### Future work

None identified.

## Step 5: Launchd and deployment

### Prompt Context

**Verbatim prompt:** Now we need to add a way for this server to automatically start on system startup. You're on a Mac mini.
**Interpretation:** Set up a launchd plist for auto-start.
**Inferred intent:** The server should survive reboots.

### What I did

- Renamed `/cmd/app` to `/cmd/claudelaunch` so `go install` produces a `claudelaunch` binary
- Created `/Users/maragubot/Library/LaunchAgents/dev.maragu.claudelaunch.plist` with `RunAtLoad` and `KeepAlive`
- Updated Makefile: replaced `build` target with `install` target using `go install`
- Ended up launching the server in a tmux session (`claudelaunch-server`) as a practical workaround for SSH

### Why

The Mac mini is accessed via SSH, and LaunchAgents are tied to the GUI login domain. `launchctl load` doesn't work from SSH sessions because the GUI domain isn't accessible.

### What worked

The tmux-based approach works immediately over SSH. The launchd plist is in place for when the machine is accessed locally or rebooted with auto-login.

### What didn't work

- `launchctl load` from SSH: silently fails (exit 134)
- `launchctl bootstrap gui/$(id -u)` from SSH: "Domain does not support specified action"
- `com.apple.provenance` xattr was present on the plist file (removed it, but loading still failed due to the SSH domain issue)

### What I learned

- LaunchAgents are GUI-domain-only on macOS. SSH sessions can't load or inspect them
- `launchctl list <label>` returns exit 113 "Could not find service" for all LaunchAgents when run from SSH, even ones that are loaded
- The `com.apple.provenance` xattr gets added to files created by sandboxed processes

### What was tricky

The launchd debugging was a dead end from SSH. No useful error messages, just silent failures.

### What warrants review

The plist at `~/Library/LaunchAgents/dev.maragu.claudelaunch.plist` should be tested after a GUI login or reboot to confirm `RunAtLoad` works.

### Future work

Consider switching to a LaunchDaemon (system-level) if the GUI domain constraint becomes a problem. Or just document the tmux approach for SSH-only setups.

## Step 6: Remote session support and browser access

### Prompt Context

**Verbatim prompt:** I actually won't use this through tmux, but the Claude app and remote sessions. Is it possible to link to the session directly?
**Interpretation:** The user wants to access launched sessions through claude.ai/code rather than tmux attach.
**Inferred intent:** Make claudelaunch a web-based launcher that links directly to remote Claude Code sessions, usable from any device (including iPad).

### What I did

- Researched Claude Code's `--remote-control` flag and `claude remote-control` server mode
- Added `--remote-control <name>` to the claude launch command in `/claudelaunch.go`
- Implemented `pollForSessionURL` that captures the tmux pane output every 500ms for up to 30 seconds, looking for the `https://claude.ai/code/session_*` URL
- Updated `/html/html.go` `SuccessPage` to accept a `LaunchResult` struct with both session name and URL
- Added an "Open Session" button that links to the claude.ai/code URL
- Briefly attempted a `claude-cli://` deep link for the desktop app, but it didn't work (possibly wrong URL format). Removed it in favor of browser-only link.

### Why

The primary use case shifted from "attach to tmux locally" to "open in browser from any device". The `--remote-control` flag makes sessions accessible at claude.ai/code, and polling the pane output is the only way to get the session URL programmatically.

### What worked

- `claude --remote-control "name"` works and prints the session URL to the terminal
- Polling `tmux capture-pane -p` reliably captures the URL after a few seconds
- The browser link works on iPad -- the actual use case the user wanted
- The `claude remote-control` server mode exists for dedicated remote-only processes, but the `--remote-control` flag on interactive sessions is the better fit here

### What didn't work

- `claude-cli://` deep links: the Claude Code URL Handler app registers the `claude-cli:` scheme, but clicking the link in the browser prompted for permission and then nothing happened. Couldn't debug further from SSH since `open` commands fail with "Domain does not support specified action". Removed the desktop app button.

### What I learned

- Claude Code has two remote modes: `claude remote-control` (dedicated server, supports `--capacity`, `--spawn`) and `claude --remote-control` (interactive session also accessible remotely)
- The `--remote-control-session-name-prefix` flag controls auto-generated session names, but you can pass a name directly: `--remote-control "My Session"`
- There's no programmatic API to get the session URL -- it's only printed to the terminal
- The Claude Code URL Handler app at `~/Applications/Claude Code URL Handler.app` is a wrapper around the claude binary that handles `claude-cli:` URLs, but the exact URL path format for session links is unclear
- `tmux capture-pane -t <session> -p` is a reliable way to read pane contents from outside the session

### What was tricky

The polling approach has an inherent tradeoff: the POST request blocks for several seconds while waiting for claude to start and print the URL. With a 30-second timeout and 500ms interval, the worst case is a 30-second page load. In practice it takes 3-5 seconds.

### What warrants review

- The polling timeout (30 seconds) and interval (500ms) in `pollForSessionURL` in `/claudelaunch.go` -- these were chosen by feel, not measurement
- The URL regex `https://claude\.ai/code/session_[a-zA-Z0-9]+` -- if the URL format ever changes, this breaks silently (returns no URL, falls back to tmux-only success page)

### Future work

- Could make the POST return immediately and poll via JavaScript on the success page, avoiding the blocking request
- The `claude-cli://` deep link could be revisited when testing from a local machine with GUI access
