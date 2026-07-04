package commands

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
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

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running daemon (the one auto-started by any command)",
	RunE:  runDaemonStop,
}

func init() {
	rootCmd.AddCommand(daemonCmd)

	daemonCmd.Flags().StringP("addr", "a", client.DefaultAddr(), "Address to listen on")
	daemonCmd.Flags().Bool("mdns", false, "Advertise <host>.local via mDNS and bind all interfaces")
	daemonCmd.Flags().String("host", "pomo", "mDNS hostname (advertised as <host>.local)")
	daemonCmd.Flags().BoolP("verbose", "v", false, "Verbose output")

	daemonCmd.AddCommand(daemonStopCmd)
}

func runDaemonStop(_ *cobra.Command, _ []string) error {
	addr := client.DefaultAddr()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	c := client.New(addr)
	if err := c.Health(ctx); err != nil {
		fmt.Printf("no daemon running at %s\n", addr)
		return nil
	}
	if err := c.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to stop daemon: %w", err)
	}
	fmt.Printf("daemon stopped (%s)\n", addr)
	return nil
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

	srv := server.New(st, nil, rootCmd.Version)

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
				url := fmt.Sprintf("http://%s.local", mdnsHost)
				if port != 80 {
					url += fmt.Sprintf(":%d", port)
				}
				slog.Info("mdns: reachable at", "url", url)
				defer func() { _ = ad.Close() }()
			}
		}
	}

	select {
	case err := <-errCh:
		return privilegedPortHint(addr, err)
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case <-srv.ShutdownRequested():
		slog.Info("shutdown requested via API")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	slog.Info("daemon stopped")
	return nil
}

// privilegedPortHint augments a bind failure on a privileged port (<1024, e.g.
// :80 so pomo.local needs no port in the URL) with how to grant the permission.
func privilegedPortHint(addr string, err error) error {
	if err == nil {
		return nil
	}
	_, portStr, splitErr := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)
	if splitErr != nil || port == 0 || port >= 1024 || !errors.Is(err, os.ErrPermission) {
		return err
	}
	return fmt.Errorf("cannot bind privileged port %d without extra permission: %w\n"+
		"  grant it once with either:\n"+
		"    sudo setcap 'cap_net_bind_service=+ep' %s   (per-binary; redo after rebuild)\n"+
		"    echo 'net.ipv4.ip_unprivileged_port_start=%d' | sudo tee /etc/sysctl.d/50-pomo.conf && sudo sysctl --system",
		port, err, selfPath(), port)
}

func selfPath() string {
	if p, err := os.Executable(); err == nil {
		return p
	}
	return "/path/to/pomo"
}

func isLoopback(host string) bool {
	if host == "localhost" || host == "" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
