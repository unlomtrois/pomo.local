package commands

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"pomo.local/internal/server"
	"pomo.local/internal/store"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run the pomo daemon (HTTP API + timer engine)",
	Long: "Runs the long-lived pomo daemon that owns the SQLite store, serves the " +
		"HTTP API, and fires completion notifications in-process. Intended to run " +
		"under a systemd --user service.",
	RunE: runDaemon,
}

func init() {
	rootCmd.AddCommand(daemonCmd)

	daemonCmd.Flags().StringP("addr", "a", "127.0.0.1:7420", "Address to listen on")
	daemonCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
}

func runDaemon(cmd *cobra.Command, _ []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	addr, _ := cmd.Flags().GetString("addr")

	dbPath, err := xdg.DataFile("pomo/pomo.db")
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}

	st, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()
	slog.Info("store opened", "path", dbPath)

	srv := server.New(st, nil)

	// Signal-aware context: Ctrl-C / SIGTERM (systemd stop) trigger shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := srv.Reconcile(ctx); err != nil {
		slog.Error("reconcile failed", "err", err)
	}

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("daemon listening", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	slog.Info("daemon stopped")
	return nil
}
