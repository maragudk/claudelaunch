package claudelaunch_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"maragu.dev/is"

	"maragu.dev/claudelaunch"
)

func TestServer_Handler(t *testing.T) {
	t.Run("returns the index page on GET", func(t *testing.T) {
		s := newTestServer()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)

		is.Equal(t, http.StatusOK, res.Code)
		is.True(t, strings.Contains(res.Body.String(), "claudelaunch"))
	})

	t.Run("returns error page when name is missing", func(t *testing.T) {
		s := newTestServer()

		res := doPost(t, s, "")
		is.Equal(t, http.StatusOK, res.Code)
		is.True(t, strings.Contains(res.Body.String(), "name is required"))
	})

	t.Run("returns error page when name contains spaces", func(t *testing.T) {
		s := newTestServer()

		res := doPost(t, s, "oh no spaces")
		is.Equal(t, http.StatusOK, res.Code)
		is.True(t, strings.Contains(res.Body.String(), "alphanumeric"))
	})

	t.Run("rejects dot-dot as a name because that would be too clever", func(t *testing.T) {
		s := newTestServer()

		res := doPost(t, s, "..")
		is.Equal(t, http.StatusOK, res.Code)
		is.True(t, strings.Contains(res.Body.String(), "alphanumeric"))
	})

	t.Run("rejects single dot as a name", func(t *testing.T) {
		s := newTestServer()

		res := doPost(t, s, ".")
		is.Equal(t, http.StatusOK, res.Code)
		is.True(t, strings.Contains(res.Body.String(), "alphanumeric"))
	})

	t.Run("shows recent names on the index page", func(t *testing.T) {
		s := newTestServer()
		s.AddRecentName("totally-real-session")

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)

		is.Equal(t, http.StatusOK, res.Code)
		is.True(t, strings.Contains(res.Body.String(), "totally-real-session"))
		is.True(t, strings.Contains(res.Body.String(), "Recent sessions"))
	})

	t.Run("index page has no recent section when list is empty", func(t *testing.T) {
		s := newTestServer()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()
		s.Handler().ServeHTTP(res, req)

		is.Equal(t, http.StatusOK, res.Code)
		is.True(t, !strings.Contains(res.Body.String(), "Recent sessions"))
	})
}

func newTestServer() *claudelaunch.Server {
	return &claudelaunch.Server{
		Log: slog.Default(),
	}
}

func doPost(t *testing.T, s *claudelaunch.Server, name string) *httptest.ResponseRecorder {
	t.Helper()

	form := url.Values{}
	if name != "" {
		form.Set("name", name)
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	s.Handler().ServeHTTP(res, req)
	return res
}
