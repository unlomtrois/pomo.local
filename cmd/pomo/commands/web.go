package commands

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"pomo.local/internal/client"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Open the pomo dashboard in your browser",
	RunE:  runWeb,
}

func init() {
	rootCmd.AddCommand(webCmd)
}

func runWeb(_ *cobra.Command, _ []string) error {
	// If a daemon is already running (any address, per the state file), open its
	// recorded URL — this respects an mDNS daemon on pomo.local or a custom port.
	if st, ok := readDaemonState(); ok {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := client.New(st.Addr).Health(ctx)
		cancel()
		if err == nil {
			return openURL(webURLFromState(st))
		}
	}

	// Otherwise make sure a daemon is up on the default address, then open it.
	if _, err := ensureDaemon(context.Background(), client.DefaultAddr()); err != nil {
		return err
	}
	url := browserURL(client.DefaultAddr(), "127.0.0.1")
	if st, ok := readDaemonState(); ok && st.URL != "" {
		url = st.URL
	}
	return openURL(url)
}

func webURLFromState(st *daemonState) string {
	if st.URL != "" {
		return st.URL
	}
	return browserURL(st.Addr, "127.0.0.1")
}

// openURL opens url in the default browser via xdg-open, printing it either way.
func openURL(url string) error {
	fmt.Printf("opening %s\n", url)
	bin, err := exec.LookPath("xdg-open")
	if err != nil {
		fmt.Printf("(no xdg-open found — open it manually: %s)\n", url)
		return nil
	}
	return exec.Command(bin, url).Start()
}
