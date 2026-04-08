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

	session, err := s.launchSession(name)
	if err != nil {
		s.Log.Error("Failed to launch session", "name", name, "error", err)
		s.render(w, r, chtml.ErrorPage("failed to launch session"))
		return
	}

	s.Log.Info("Launched session", "session", session)
	s.render(w, r, chtml.SuccessPage(session))
}

func (s *Server) render(w http.ResponseWriter, r *http.Request, n g.Node) {
	ghttp.Adapt(func(w http.ResponseWriter, r *http.Request) (g.Node, error) {
		return n, nil
	})(w, r)
}

func (s *Server) launchSession(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home dir: %w", err)
	}

	dir := filepath.Join(home, "Developer", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating directory: %w", err)
	}

	session := fmt.Sprintf("%v-%v", name, time.Now().Unix())

	cmd := exec.Command("tmux", "new-session", "-d", "-s", session,
		fmt.Sprintf("cd %v && claude --dangerously-skip-permissions", dir))
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("tmux: %w: %s", err, output)
	}
	return session, nil
}
