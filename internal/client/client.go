// Package client is the HTTP client the CLI (and later the MCP server) use to
// talk to the pomo daemon. It is the counterpart to internal/server.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

// DefaultAddr is the daemon's listen/dial address, shared by the daemon command
// and its clients so they agree. Overridable via $POMO_ADDR (e.g. ":80" so
// pomo.local needs no port in the URL).
func DefaultAddr() string {
	if v := os.Getenv("POMO_ADDR"); v != "" {
		return v
	}
	return "127.0.0.1:7420"
}

// Sentinel errors mirroring the daemon's status codes.
var (
	// ErrActiveExists is returned when a session is already active (HTTP 409).
	ErrActiveExists = errors.New("an active pomodoro session already exists")
	// ErrNoActive is returned when there is no active session (HTTP 404).
	ErrNoActive = errors.New("no active pomodoro session")
)

// Session mirrors the daemon's session JSON.
type Session struct {
	ID        int64     `json:"id"`
	Hash      string    `json:"hash"`
	Topic     string    `json:"topic"`
	StartTime time.Time `json:"start_time"`
	StopTime  time.Time `json:"stop_time"`
	Duration  int64     `json:"duration"` // nanoseconds
	Status    string    `json:"status"`
	Source    string    `json:"source"`
	Message   string    `json:"message"`
	Hint      string    `json:"hint"`
	Email     bool      `json:"email"`
	Open      bool      `json:"open"`
}

// StartParams is the payload for StartSession. Duration is a Go duration string.
type StartParams struct {
	Topic       string `json:"topic"`
	Duration    string `json:"duration"`
	Source      string `json:"source"`
	Message     string `json:"message"`
	Hint        string `json:"hint"`
	Email       bool   `json:"email"`
	Project     string `json:"project,omitempty"`      // .pomo stable id
	ProjectName string `json:"project_name,omitempty"` // .pomo display name
	Open        bool   `json:"open,omitempty"`         // open-ended stopwatch
}

// Client talks to a pomo daemon over HTTP.
type Client struct {
	base string
	http *http.Client
}

// New returns a Client for the daemon at addr (host:port).
func New(addr string) *Client {
	return &Client{
		base: "http://" + dialHost(addr),
		http: &http.Client{Timeout: 5 * time.Second},
	}
}

// dialHost makes a listen address dialable: a wildcard/empty host (":80",
// "0.0.0.0:80", "[::]:80") becomes 127.0.0.1 for connecting from this machine.
func dialHost(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port)
}

// Stats mirrors the daemon's DB aggregate summary.
type Stats struct {
	TotalSessions int            `json:"total_sessions"`
	ByStatus      map[string]int `json:"by_status"`
	Projects      int            `json:"projects"`
	TrackedNanos  int64          `json:"tracked_nanos"`
}

// StatusInfo is the daemon status: version, DB stats, and active session.
type StatusInfo struct {
	Version  string   `json:"version"`
	Sessions Stats    `json:"sessions"`
	Active   *Session `json:"active"`
}

// Status fetches the daemon's version and DB summary.
func (c *Client) Status(ctx context.Context) (*StatusInfo, error) {
	var info StatusInfo
	if err := c.do(ctx, http.MethodGet, "/api/stats", nil, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// Shutdown asks the daemon to stop gracefully.
func (c *Client) Shutdown(ctx context.Context) error {
	return c.do(ctx, http.MethodPost, "/api/shutdown", nil, nil)
}

// Health reports whether the daemon is reachable and healthy.
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/healthz", nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("daemon unhealthy: %s", resp.Status)
	}
	return nil
}

// StartSession starts a new session. Returns ErrActiveExists on 409.
func (c *Client) StartSession(ctx context.Context, p StartParams) (*Session, error) {
	var sess Session
	if err := c.do(ctx, http.MethodPost, "/api/sessions", p, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

// ActiveSession returns the active session, or ErrNoActive on 404.
func (c *Client) ActiveSession(ctx context.Context) (*Session, error) {
	var sess Session
	if err := c.do(ctx, http.MethodGet, "/api/sessions/active", nil, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

// StopActive cancels the active session, or ErrNoActive on 404.
func (c *Client) StopActive(ctx context.Context) (*Session, error) {
	var sess Session
	if err := c.do(ctx, http.MethodPost, "/api/sessions/active/stop", nil, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

// EndActive closes the active session, optionally overriding its topic (empty
// topic keeps the existing one). Returns ErrNoActive if none is active.
func (c *Client) EndActive(ctx context.Context, topic string) (*Session, error) {
	var body any
	if topic != "" {
		body = map[string]string{"topic": topic}
	}
	var sess Session
	if err := c.do(ctx, http.MethodPost, "/api/sessions/active/end", body, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

// ErrAmbiguousHash is returned when a hash prefix matches multiple sessions.
var ErrAmbiguousHash = errors.New("ambiguous session hash prefix")

// ErrSessionNotFound is returned when a hash prefix matches no session.
var ErrSessionNotFound = errors.New("session not found")

// SessionByHash resolves a session by hash prefix (own status handling, since
// 404/409 here mean not-found/ambiguous rather than no-active/active-exists).
func (c *Client) SessionByHash(ctx context.Context, prefix string) (*Session, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/api/sessions/by-hash/"+prefix, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		var sess Session
		if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil {
			return nil, err
		}
		return &sess, nil
	case http.StatusNotFound:
		return nil, ErrSessionNotFound
	case http.StatusConflict:
		return nil, ErrAmbiguousHash
	default:
		return nil, fmt.Errorf("lookup failed: %s: %s", resp.Status, readErr(resp.Body))
	}
}

// EditParams is a partial update; nil fields are omitted from the request and
// left unchanged. A non-nil pointer to "" clears the topic.
type EditParams struct {
	Topic    *string    `json:"topic,omitempty"`
	Start    *time.Time `json:"start_time,omitempty"`
	Duration *string    `json:"duration,omitempty"` // Go duration string
}

// EditByHash applies a partial update to a session resolved by hash prefix.
func (c *Client) EditByHash(ctx context.Context, prefix string, p EditParams) (*Session, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch,
		c.base+"/api/sessions/by-hash/"+prefix, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		var sess Session
		if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil {
			return nil, err
		}
		return &sess, nil
	case http.StatusNotFound:
		return nil, ErrSessionNotFound
	case http.StatusConflict:
		return nil, ErrAmbiguousHash
	default:
		return nil, fmt.Errorf("edit failed: %s: %s", resp.Status, readErr(resp.Body))
	}
}

// DeleteByHash deletes a session by hash prefix and returns the deleted row.
func (c *Client) DeleteByHash(ctx context.Context, prefix string) (*Session, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.base+"/api/sessions/by-hash/"+prefix, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		var sess Session
		if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil {
			return nil, err
		}
		return &sess, nil
	case http.StatusNotFound:
		return nil, ErrSessionNotFound
	case http.StatusConflict:
		return nil, ErrAmbiguousHash
	default:
		return nil, fmt.Errorf("delete failed: %s: %s", resp.Status, readErr(resp.Body))
	}
}

// ListSessions returns recent sessions, newest first. A non-empty projectExtID
// scopes the list to that project.
func (c *Client) ListSessions(ctx context.Context, limit int, projectExtID string) ([]Session, error) {
	path := fmt.Sprintf("/api/sessions?limit=%d", limit)
	if projectExtID != "" {
		path += "&project=" + url.QueryEscape(projectExtID)
	}
	var sessions []Session
	if err := c.do(ctx, http.MethodGet, path, nil, &sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

// do performs a request, mapping known status codes to sentinel errors and
// decoding a successful JSON response into out (if non-nil).
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.base+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusConflict:
		return ErrActiveExists
	case http.StatusNotFound:
		return ErrNoActive
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("daemon error: %s: %s", resp.Status, readErr(resp.Body))
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func readErr(r io.Reader) string {
	var e struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(r).Decode(&e); err == nil && e.Error != "" {
		return e.Error
	}
	return "unknown error"
}
