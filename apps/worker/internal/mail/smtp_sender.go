package mail

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	"github.com/xdouglas90/petcontrol_monorepo/worker/internal/config"
	"github.com/xdouglas90/petcontrol_monorepo/worker/internal/queue"
)

type SMTPSender struct {
	logger *slog.Logger
	cfg    config.Config
}

func NewSMTPSender(logger *slog.Logger, cfg config.Config) *SMTPSender {
	return &SMTPSender{logger: logger, cfg: cfg}
}

func (s *SMTPSender) SendPersonAccessCredentials(_ context.Context, payload queue.PersonAccessCredentialsPayload) error {
	subject := "Seu acesso ao PetControl foi criado"
	body := buildAccessCredentialsBody(payload, s.cfg.AppBaseURL)
	message := strings.Join([]string{
		fmt.Sprintf("From: %s <%s>", s.cfg.SMTPFromName, s.cfg.SMTPFromEmail),
		fmt.Sprintf("To: %s", payload.RecipientEmail),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	address := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)
	var auth smtp.Auth
	if strings.TrimSpace(s.cfg.SMTPUsername) != "" {
		auth = smtp.PlainAuth("", s.cfg.SMTPUsername, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	}

	if err := smtp.SendMail(address, auth, s.cfg.SMTPFromEmail, []string{payload.RecipientEmail}, []byte(message)); err != nil {
		return err
	}

	s.logger.Info("person access credentials email sent",
		"user_id", payload.UserID,
		"person_id", payload.PersonID,
		"recipient_email", payload.RecipientEmail,
	)

	return nil
}

func buildAccessCredentialsBody(payload queue.PersonAccessCredentialsPayload, appBaseURL string) string {
	systemURL := strings.TrimSpace(payload.SystemURL)
	if systemURL == "" {
		systemURL = strings.TrimSpace(appBaseURL)
	}

	name := strings.TrimSpace(payload.RecipientName)
	if name == "" {
		name = "Olá"
	}

	return fmt.Sprintf(
		"%s,\n\nSeu acesso ao PetControl foi criado.\n\nEmail: %s\nSenha temporária: %s\nPerfil: %s\nAcesse: %s\n\nNo primeiro acesso, altere sua senha.\n",
		name,
		payload.RecipientEmail,
		payload.TemporaryPassword,
		payload.Role,
		systemURL,
	)
}
