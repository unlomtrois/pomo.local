package mail

import (
	"fmt"
	"net"
	"net/smtp"
	"strconv"

	"pomo.local/internal/config"
)

func Send(cfg *config.MailConfig, password, subject, body string) error {
	auth := smtp.PlainAuth("", cfg.Sender, password, cfg.Host)

	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	msg := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", cfg.Receiver, subject, body)

	return smtp.SendMail(addr, auth, cfg.Sender, []string{cfg.Receiver}, []byte(msg))
}
