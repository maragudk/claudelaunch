// Package claudelaunch launches persistent Claude Code sessions inside tmux.
package claudelaunch

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	g "maragu.dev/gomponents"
	ghttp "maragu.dev/gomponents/http"

	chtml "maragu.dev/claudelaunch/html"
)

// Server serves HTTP requests that launch Claude Code sessions in tmux.
type Server struct {
	Log *slog.Logger
}

var validName = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// Handler returns the HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.index)
	mux.HandleFunc("POST /", s.launch)
	return mux
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, chtml.IndexPage())
}

func (s *Server) launch(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")

	if name == "" {
		s.render(w, r, chtml.ErrorPage("name is required"))
		return
	}

	if !validName.MatchString(name) || !filepath.IsLocal(name) || name == "." {
		s.render(w, r, chtml.ErrorPage("name must be alphanumeric (dashes, underscores, and dots allowed)"))
		return
	}

	result, err := s.launchSession(name)
	if err != nil {
		s.Log.Error("Failed to launch session", "name", name, "error", err)
		s.render(w, r, chtml.ErrorPage("failed to launch session"))
		return
	}

	s.Log.Info("Launched session", "session", result.Session, "url", result.URL)
	s.render(w, r, chtml.SuccessPage(chtml.LaunchResult{
		Session: result.Session,
		URL:     result.URL,
	}))
}

func (s *Server) render(w http.ResponseWriter, r *http.Request, n g.Node) {
	ghttp.Adapt(func(w http.ResponseWriter, r *http.Request) (g.Node, error) {
		return n, nil
	})(w, r)
}

// LaunchResult holds the result of launching a session.
type LaunchResult struct {
	Session string
	URL     string
}

func (s *Server) launchSession(name string) (LaunchResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return LaunchResult{}, fmt.Errorf("getting home dir: %w", err)
	}

	dir := filepath.Join(home, "Developer", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return LaunchResult{}, fmt.Errorf("creating directory: %w", err)
	}

	session := fmt.Sprintf("%v-%v", name, time.Now().Unix())

	cmd := exec.Command("tmux", "new-session", "-d", "-s", session,
		fmt.Sprintf("cd %v && claude --dangerously-skip-permissions --remote-control %q", dir, name))
	if output, err := cmd.CombinedOutput(); err != nil {
		return LaunchResult{}, fmt.Errorf("tmux: %w: %s", err, output)
	}

	url, err := pollForSessionURL(session, 30*time.Second)
	if err != nil {
		s.Log.Warn("Could not capture session URL", "session", session, "error", err)
	}

	return LaunchResult{Session: session, URL: url}, nil
}

var sessionURLPattern = regexp.MustCompile(`https://claude\.ai/code/session_[a-zA-Z0-9]+`)

func pollForSessionURL(session string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	interval := 500 * time.Millisecond

	for time.Now().Before(deadline) {
		cmd := exec.Command("tmux", "capture-pane", "-t", session, "-p")
		output, err := cmd.CombinedOutput()
		if err != nil {
			time.Sleep(interval)
			continue
		}

		if url := sessionURLPattern.FindString(strings.TrimSpace(string(output))); url != "" {
			return url, nil
		}

		time.Sleep(interval)
	}

	return "", fmt.Errorf("timed out waiting for session URL")
}
