package mail

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/xdouglas90/petcontrol_monorepo/worker/internal/config"
	"github.com/xdouglas90/petcontrol_monorepo/worker/internal/queue"
)

func TestSMTPSender_SendPersonAccessCredentials_WithMailHog(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "mailhog/mailhog:v1.0.1",
		ExposedPorts: []string{"1025/tcp", "8025/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("1025/tcp"),
			wait.ForListeningPort("8025/tcp"),
		).WithDeadline(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		if isDockerUnavailable(err) {
			t.Skipf("skipping MailHog integration test; Docker not available: %v", err)
		}
		t.Fatalf("failed to start MailHog container: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)
	smtpPort, err := container.MappedPort(ctx, "1025/tcp")
	require.NoError(t, err)
	apiPort, err := container.MappedPort(ctx, "8025/tcp")
	require.NoError(t, err)

	sender := NewSMTPSender(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		config.Config{
			SMTPHost:      host,
			SMTPPort:      smtpPort.Port(),
			SMTPFromEmail: "no-reply@petcontrol.local",
			SMTPFromName:  "PetControl",
			AppBaseURL:    "http://localhost:5173",
		},
	)

	payload := queue.PersonAccessCredentialsPayload{
		Version:           1,
		CompanyID:         "company-1",
		PersonID:          "person-1",
		UserID:            "user-1",
		RecipientName:     "Maria Silva",
		RecipientEmail:    "maria.silva@petcontrol.local",
		TemporaryPassword: "senha-temporaria",
		Role:              "system",
		OccurredAt:        time.Now().UTC(),
	}

	err = sender.SendPersonAccessCredentials(ctx, payload)
	require.NoError(t, err)

	baseURL := fmt.Sprintf("http://%s:%s", host, apiPort.Port())
	var responseBody string
	require.Eventually(t, func() bool {
		res, reqErr := http.Get(baseURL + "/api/v2/messages")
		if reqErr != nil {
			return false
		}
		defer res.Body.Close()

		body, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			return false
		}
		responseBody = string(body)
		return strings.Contains(responseBody, payload.RecipientEmail) &&
			strings.Contains(responseBody, payload.TemporaryPassword)
	}, 10*time.Second, 250*time.Millisecond)
}

func isDockerUnavailable(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "docker") &&
		(strings.Contains(msg, "permission denied") ||
			strings.Contains(msg, "cannot connect") ||
			strings.Contains(msg, "no such file or directory") ||
			strings.Contains(msg, "daemon"))
}
