package cli

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"syscall"

	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

// AuthCommand is basically a wrapper around notify-send
type AuthCommand struct {
	forMail  bool
	forToggl bool
}

func ParseAuth(args []string) *AuthCommand {
	cmd := AuthCommand{}
	fs := flag.NewFlagSet("auth", flag.ExitOnError)
	fs.BoolVar(&cmd.forMail, "mail", false, "Auth for SMTP notifications")
	fs.BoolVar(&cmd.forToggl, "toggl", false, "Auth for Toggl Track")
	fs.Parse(args)

	return &cmd
}

func (cmd *AuthCommand) Run() error {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return fmt.Errorf("your terminal in non-interactive")
	}

	if cmd.forMail {
		if err := cmd.authService("pomo-smtp", "SMTP App Password"); err != nil {
			return err
		}
	}
	if cmd.forToggl {
		if err := cmd.authService("pomo-toggl", "Toggl API Token"); err != nil {
			return err
		}
	}

	return nil
}

func (cmd *AuthCommand) authService(service, label string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Enter Username/Email for %s: ", service)
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)

	fmt.Printf("Enter %s (input hidden): ", label)
	bytePassword, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return err
	}
	defer func() {
		for i := range bytePassword {
			bytePassword[i] = 0
		}
	}()

	password := strings.TrimSpace(string(bytePassword))
	err = keyring.Set(service, user, password)
	if err != nil {
		return fmt.Errorf("could not save to keyring: %v", err)
	}
	fmt.Printf("\nSuccess! Credentials for %s saved securely.\n", service)

	if service == "pomo-smtp" {
		fmt.Println("Testing SMTP connection...")

		pass, err := keyring.Get(service, user)
		if err != nil {
			return err
		}

		host := "smtp.gmail.com"
		auth := smtp.PlainAuth("", user, pass, host)

		client, err := smtp.Dial(host + ":587")
		if err != nil {
			return fmt.Errorf("could not connect to SMTP: %v", err)
		}

		if err = client.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return fmt.Errorf("TLS error: %v", err)
		}

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("Auth failed: invalid credentials or App Password required")
		}
		fmt.Println("Success! SMTP credentials verified successfully!")
	}

	return nil
}
