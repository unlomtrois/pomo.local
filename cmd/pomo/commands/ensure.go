package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/adrg/xdg"
	"pomo.local/internal/client"
)

// ensureDaemon returns a client to the daemon at addr, auto-spawning a detached
// `pomo daemon` process if none is reachable.
//
// The spawned daemon is placed in its own session (Setsid) so it outlives this
// CLI invocation and the terminal, and its stdout/stderr are redirected to a
// persistent log file so its activity is inspectable after the fact.
func ensureDaemon(ctx context.Context, addr string) (*client.Client, error) {
	c := client.New(addr)

	if err := c.Health(ctx); err == nil {
		return c, nil
	}

	slog.Info("daemon not reachable, auto-spawning", "addr", addr)

	logPath, err := spawnDaemon(addr)
	if err != nil {
		return nil, err
	}
	slog.Info("daemon spawned", "addr", addr, "log", logPath)

	// Wait for the freshly spawned daemon to come up.
	if err := waitHealthy(ctx, c, 5*time.Second); err != nil {
		return nil, fmt.Errorf("daemon did not become healthy (see %s): %w", logPath, err)
	}
	return c, nil
}

// spawnDaemon starts a detached `pomo daemon` and returns the log file path.
func spawnDaemon(addr string) (string, error) {
	bin, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not find pomo executable: %w", err)
	}

	logPath, err := xdg.StateFile("pomo/daemon.log")
	if err != nil {
		return "", fmt.Errorf("resolve daemon log path: %w", err)
	}
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("open daemon log: %w", err)
	}
	defer func() { _ = logFile.Close() }()

	cmd := exec.Command(bin, "daemon", "--addr", addr, "--verbose")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	// Detach: new session so the daemon survives this CLI exiting / the TTY closing.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start daemon: %w", err)
	}
	// Release the child so it is not left as a zombie when this process exits.
	if err := cmd.Process.Release(); err != nil {
		slog.Warn("release daemon process", "err", err)
	}
	return logPath, nil
}

// waitHealthy polls the daemon's health endpoint until it succeeds or times out.
func waitHealthy(ctx context.Context, c *client.Client, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		if err := c.Health(ctx); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timed out after %s", timeout)
			}
		}
	}
}
