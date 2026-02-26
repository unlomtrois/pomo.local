package mail

import (
	"fmt"
	"net"
	"net/smtp"
	"strconv"

	"github.com/zalando/go-keyring"
	"pomo.local/internal/config"
)

func SendMail(subject, body string) error {
	cfg := &config.MailConfig{}
	if err := cfg.Load(); err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("Invalid config (try \"pomo auth --email\" again): %v", err)
	}

	password, err := keyring.Get("pomo-smtp", cfg.Sender)
	if err != nil {
		return fmt.Errorf("Failed to get keyring for sender: %v, %w. (try to \"pomo auth --email\" again)", cfg.Sender, err)
	}

	fmt.Println("send email to", cfg.Receiver)
	err = sendMail(cfg, password, "Pomo:"+subject, body)
	if err != nil {
		return fmt.Errorf("Failed to send an email: %v", err)
	}

	fmt.Println("Mail sended!")
	return nil
}

func sendMail(cfg *config.MailConfig, password, subject, body string) error {
	auth := smtp.PlainAuth("", cfg.Sender, password, cfg.Host)

	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	msg := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", cfg.Receiver, subject, body)

	return smtp.SendMail(addr, auth, cfg.Sender, []string{cfg.Receiver}, []byte(msg))
}
