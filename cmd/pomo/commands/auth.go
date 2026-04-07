package commands

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"golang.org/x/term"
	"pomo.local/internal/config"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Setup credentials for integrations",
	RunE:  runAuth,
}

func init() {
	rootCmd.AddCommand(authCmd)

	authCmd.Flags().Bool("email", false, "Auth for SMTP notifications")
	authCmd.Flags().Bool("toggl", false, "Auth for Toggl Track")
}

func runAuth(cmd *cobra.Command, args []string) error {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return fmt.Errorf("your terminal in non-interactive")
	}

	forEmail, _ := cmd.Flags().GetBool("email")
	forToggl, _ := cmd.Flags().GetBool("toggl")

	if forEmail {
		if err := authService("pomo-smtp", "SMTP App Password"); err != nil {
			return err
		}
	}
	if forToggl {
		if err := authService("pomo-toggl", "Toggl API Token"); err != nil {
			return err
		}
	}

	return nil
}

func authService(service, label string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Enter Username/Email for %s: ", service)
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)

	fmt.Printf("Enter %s (input hidden): ", label)
	bytePassword, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return err
	}

	password := strings.TrimSpace(string(bytePassword))
	err = keyring.Set(service, user, password)
	if err != nil {
		return fmt.Errorf("could not save to keyring: %v", err)
	}
	fmt.Printf("\nSuccess! Credentials for %s saved securely.\n", service)

	if service == "pomo-smtp" {
		pass, err := keyring.Get(service, user)
		if err != nil {
			return err
		}

		fmt.Printf("Enter your smtp host: ")
		host, _ := reader.ReadString('\n')
		host = strings.TrimSpace(host)

		fmt.Printf("Enter your smtp port: ")
		portString, _ := reader.ReadString('\n')
		portString = strings.TrimSpace(portString)
		port, err := strconv.ParseInt(portString, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse port: %v", err)
		}

		cfg := config.MailConfig{
			Host:     host,
			Port:     int(port),
			Sender:   user,
			Receiver: user,
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config, due: %v", err)
		}
		fmt.Println("New config successfully saved!")

		fmt.Println("Testing SMTP connection...")
		auth := smtp.PlainAuth("", cfg.Sender, pass, cfg.Host)

		addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
		client, err := smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("could not connect to SMTP: %v", err)
		}

		if err = client.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return fmt.Errorf("TLS error: %v", err)
		}

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("auth failed: invalid credentials or App Password required")
		}
		fmt.Println("Success! SMTP credentials verified successfully!")
	}

	return nil
}
