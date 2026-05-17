package notifier

import (
	"fmt"
	"net/smtp"
)

// EmailConfig holds SMTP connection settings.
type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// buildEmailContent returns the subject and plain-text body for a notification email.
func buildEmailContent(sub Subscription) (subject, body string) {
	subject = fmt.Sprintf("Stadtbibliothek Leipzig: \"%s\" ist jetzt ausleihbar", sub.Title)
	mediaLabel := "Film"
	if sub.Type == "game" {
		mediaLabel = "Spiel"
	}
	platformInfo := ""
	if sub.Platform != "" {
		platformInfo = fmt.Sprintf(" (%s)", sub.Platform)
	}
	body = fmt.Sprintf(
		"Hallo,\n\nder %s \"%s\"%s ist jetzt in der Stadtbibliothek Leipzig ausleihbar.\n\nViel Spaß beim Ausleihen!\n",
		mediaLabel, sub.Title, platformInfo,
	)
	return subject, body
}

// SendNotification sends a notification email for the given subscription.
func SendNotification(cfg EmailConfig, sub Subscription) error {
	subject, body := buildEmailContent(sub)
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		cfg.From, sub.Email, subject, body,
	)
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}
	return smtp.SendMail(addr, auth, cfg.From, []string{sub.Email}, []byte(msg))
}

// EmailSender implements the Sender interface using SMTP.
type EmailSender struct {
	cfg EmailConfig
}

// NewEmailSender creates an EmailSender with the given config.
func NewEmailSender(cfg EmailConfig) *EmailSender {
	return &EmailSender{cfg: cfg}
}

// Send implements Sender.
func (es *EmailSender) Send(sub Subscription) error {
	return SendNotification(es.cfg, sub)
}
