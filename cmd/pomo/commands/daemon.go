package commands

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"pomo.local/internal/client"
	"pomo.local/internal/mdns"
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

	daemonCmd.Flags().StringP("addr", "a", client.DefaultAddr, "Address to listen on")
	daemonCmd.Flags().Bool("mdns", false, "Advertise <host>.local via mDNS and bind all interfaces")
	daemonCmd.Flags().String("host", "pomo", "mDNS hostname (advertised as <host>.local)")
	daemonCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
}

func runDaemon(cmd *cobra.Command, _ []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	addr, _ := cmd.Flags().GetString("addr")
	mdnsOn, _ := cmd.Flags().GetBool("mdns")
	mdnsHost, _ := cmd.Flags().GetString("host")

	// mDNS only helps if the daemon accepts connections from the LAN, so when
	// it's on and we'd otherwise bind loopback-only, listen on all interfaces.
	if mdnsOn {
		if host, port, err := net.SplitHostPort(addr); err == nil && isLoopback(host) {
			addr = net.JoinHostPort("0.0.0.0", port)
			slog.Info("mdns: binding all interfaces so the .local name is reachable", "addr", addr)
		}
	}

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

	if mdnsOn {
		if _, portStr, err := net.SplitHostPort(addr); err == nil {
			port, _ := strconv.Atoi(portStr)
			if ad, err := mdns.Advertise(mdnsHost, port); err != nil {
				slog.Error("mdns advertise failed", "err", err)
			} else {
				slog.Info("mdns: reachable at", "url", fmt.Sprintf("http://%s.local:%d", mdnsHost, port))
				defer func() { _ = ad.Close() }()
			}
		}
	}

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

func isLoopback(host string) bool {
	if host == "localhost" || host == "" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
