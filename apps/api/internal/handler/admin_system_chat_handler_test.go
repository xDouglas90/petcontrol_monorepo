package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	appjwt "github.com/xdouglas90/petcontrol_monorepo/internal/jwt"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/realtime"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

func adminSystemChatServiceWithMock(t *testing.T) (*service.AdminSystemChatService, pgxmock.PgxPoolIface) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	return service.NewAdminSystemChatService(sqlc.New(mock)), mock
}

func chatToken(t *testing.T, companyID pgtype.UUID, userID pgtype.UUID, role string) string {
	t.Helper()

	token, err := appjwt.GenerateToken("secret", time.Hour, appjwt.Claims{
		UserID:    userID.String(),
		CompanyID: companyID.String(),
		Role:      role,
		Kind:      "owner",
	})
	require.NoError(t, err)

	return token
}

func chatUserRow(id pgtype.UUID, email string, role sqlc.UserRoleType) *pgxmock.Rows {
	return pgxmock.NewRows([]string{
		"id",
		"email",
		"email_verified",
		"email_verified_at",
		"role",
		"is_active",
		"created_at",
		"updated_at",
		"deleted_at",
	}).AddRow(id, email, true, time.Now().Add(-time.Hour), role, true, time.Now().Add(-time.Hour), nil, nil)
}

func chatCompanyUserRow(t *testing.T, companyID pgtype.UUID, userID pgtype.UUID, active bool) *pgxmock.Rows {
	return pgxmock.NewRows([]string{
		"id",
		"company_id",
		"user_id",
		"kind",
		"is_owner",
		"is_active",
		"created_at",
		"updated_at",
		"deleted_at",
	}).AddRow(newHandlerUUID(t), companyID, userID, sqlc.UserKindOwner, false, active, time.Now().Add(-time.Hour), nil, nil)
}

func newChatSocketRouter(serviceUnderTest *service.AdminSystemChatService, hub *realtime.InternalChatHub) *gin.Engine {
	router := gin.New()
	handler := NewAdminSystemChatHandler(serviceUnderTest, hub, []string{"http://localhost:*", "http://127.0.0.1:*"})
	protected := router.Group("/")
	protected.Use(middleware.Auth("secret"), middleware.Tenant())
	protected.GET("/api/v1/chat/system/:user_id/ws", handler.Connect)
	protected.POST("/api/v1/chat/system/:user_id/messages", handler.CreateMessage)
	return router
}

func websocketURL(baseURL string, userID pgtype.UUID) string {
	return "ws" + strings.TrimPrefix(baseURL, "http") + "/api/v1/chat/system/" + userID.String() + "/ws"
}

func messagesURL(baseURL string, userID pgtype.UUID) string {
	return baseURL + "/api/v1/chat/system/" + userID.String() + "/messages"
}

func readSocketEvent(t *testing.T, ctx context.Context, conn *websocket.Conn) map[string]any {
	t.Helper()

	var event map[string]any
	require.NoError(t, wsjson.Read(ctx, conn, &event))
	return event
}

func expectParticipantValidation(
	t *testing.T,
	mock pgxmock.PgxPoolIface,
	companyID pgtype.UUID,
	currentUserID pgtype.UUID,
	contactUserID pgtype.UUID,
	currentRole sqlc.UserRoleType,
	contactRole sqlc.UserRoleType,
) {
	mock.ExpectQuery(`(?s)name: GetCompanyUser`).
		WithArgs(companyID, currentUserID).
		WillReturnRows(chatCompanyUserRow(t, companyID, currentUserID, true))
	mock.ExpectQuery(`(?s)name: GetCompanyUser`).
		WithArgs(companyID, contactUserID).
		WillReturnRows(chatCompanyUserRow(t, companyID, contactUserID, true))
	mock.ExpectQuery(`(?s)name: GetUserByID`).
		WithArgs(currentUserID).
		WillReturnRows(chatUserRow(currentUserID, "admin@petcontrol.local", currentRole))
	mock.ExpectQuery(`(?s)name: GetUserByID`).
		WithArgs(contactUserID).
		WillReturnRows(chatUserRow(contactUserID, "system@petcontrol.local", contactRole))
}

func TestAdminSystemChatHandler_Connect_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceUnderTest, mock := adminSystemChatServiceWithMock(t)
	defer mock.Close()

	companyID := newHandlerUUID(t)
	currentUserID := newHandlerUUID(t)
	contactUserID := newHandlerUUID(t)
	token := chatToken(t, companyID, currentUserID, "admin")

	expectParticipantValidation(t, mock, companyID, currentUserID, contactUserID, sqlc.UserRoleTypeAdmin, sqlc.UserRoleTypeSystem)

	hub := realtime.NewInternalChatHub()
	server := httptest.NewServer(newChatSocketRouter(serviceUnderTest, hub))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, resp, err := websocket.Dial(ctx, websocketURL(server.URL, contactUserID), &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Authorization": []string{"Bearer " + token},
		},
		Subprotocols: []string{internalChatSocketSubprotocol},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

	var event map[string]any
	event = readSocketEvent(t, ctx, conn)
	require.Equal(t, "chat.connected", event["type"])
	require.Equal(t, companyID.String(), event["company_id"])
	require.Equal(t, contactUserID.String(), event["counterpart_user_id"])
	require.Equal(t, currentUserID.String(), event["viewer_user_id"])
	require.Equal(t, "admin", event["viewer_role"])

	event = readSocketEvent(t, ctx, conn)
	require.Equal(t, "chat.presence.snapshot", event["type"])
	presences, ok := event["presences"].([]any)
	require.True(t, ok)
	require.Len(t, presences, 2)

	require.Equal(t, 1, hub.TotalConnections())
	require.Equal(t, 1, hub.ConnectionCount(companyID.String(), currentUserID.String()))

	require.NoError(t, conn.Close(websocket.StatusNormalClosure, "test complete"))
	require.Eventually(t, func() bool {
		return hub.TotalConnections() == 0
	}, time.Second, 10*time.Millisecond)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminSystemChatHandler_Connect_RejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceUnderTest, mock := adminSystemChatServiceWithMock(t)
	defer mock.Close()

	contactUserID := newHandlerUUID(t)
	hub := realtime.NewInternalChatHub()
	server := httptest.NewServer(newChatSocketRouter(serviceUnderTest, hub))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, websocketURL(server.URL, contactUserID), &websocket.DialOptions{
		Subprotocols: []string{internalChatSocketSubprotocol},
	})
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body, readErr := io.ReadAll(resp.Body)
	require.NoError(t, readErr)
	require.Contains(t, string(body), "missing bearer token")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminSystemChatHandler_Connect_RejectsForbiddenPair(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceUnderTest, mock := adminSystemChatServiceWithMock(t)
	defer mock.Close()

	companyID := newHandlerUUID(t)
	currentUserID := newHandlerUUID(t)
	contactUserID := newHandlerUUID(t)
	token := chatToken(t, companyID, currentUserID, "admin")

	expectParticipantValidation(t, mock, companyID, currentUserID, contactUserID, sqlc.UserRoleTypeAdmin, sqlc.UserRoleTypeCommon)

	hub := realtime.NewInternalChatHub()
	server := httptest.NewServer(newChatSocketRouter(serviceUnderTest, hub))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, websocketURL(server.URL, contactUserID), &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Authorization": []string{"Bearer " + token},
		},
		Subprotocols: []string{internalChatSocketSubprotocol},
	})
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)

	body, readErr := io.ReadAll(resp.Body)
	require.NoError(t, readErr)
	require.Contains(t, string(body), "failed to connect chat socket")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminSystemChatHandler_CreateMessage_BroadcastsRealtimeEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceUnderTest, mock := adminSystemChatServiceWithMock(t)
	defer mock.Close()

	companyID := newHandlerUUID(t)
	currentUserID := newHandlerUUID(t)
	contactUserID := newHandlerUUID(t)
	conversationID := newHandlerUUID(t)
	messageID := newHandlerUUID(t)
	personID := newHandlerUUID(t)
	token := chatToken(t, companyID, currentUserID, "admin")

	expectParticipantValidation(t, mock, companyID, currentUserID, contactUserID, sqlc.UserRoleTypeAdmin, sqlc.UserRoleTypeSystem)
	expectParticipantValidation(t, mock, companyID, currentUserID, contactUserID, sqlc.UserRoleTypeAdmin, sqlc.UserRoleTypeSystem)
	mock.ExpectQuery(`(?s)name: UpsertAdminSystemConversation`).
		WithArgs(companyID, currentUserID, contactUserID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "admin_user_id", "system_user_id", "created_at", "updated_at"}).
			AddRow(conversationID, companyID, currentUserID, contactUserID, time.Now().Add(-time.Hour), time.Now()))
	mock.ExpectQuery(`(?s)name: InsertAdminSystemMessage`).
		WithArgs(conversationID, companyID, currentUserID, "Nova mensagem em tempo real").
		WillReturnRows(pgxmock.NewRows([]string{"id", "conversation_id", "company_id", "sender_user_id", "body", "created_at", "updated_at", "deleted_at"}).
			AddRow(messageID, conversationID, companyID, currentUserID, "Nova mensagem em tempo real", time.Now(), nil, nil))
	mock.ExpectQuery(`(?s)name: GetUserByID`).
		WithArgs(currentUserID).
		WillReturnRows(chatUserRow(currentUserID, "admin@petcontrol.local", sqlc.UserRoleTypeAdmin))
	mock.ExpectQuery(`(?s)name: GetUserProfile`).
		WithArgs(currentUserID).
		WillReturnRows(pgxmock.NewRows([]string{"user_id", "person_id", "created_at"}).AddRow(currentUserID, personID, time.Now().Add(-time.Hour)))
	mock.ExpectQuery(`(?s)name: GetPerson`).
		WithArgs(personID).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "kind", "is_active", "has_system_user", "created_at", "updated_at",
			"full_name", "short_name", "gender_identity", "marital_status", "image_url",
			"birth_date", "cpf", "identifications_created_at", "identifications_updated_at",
		}).AddRow(
			personID, sqlc.PersonKindEmployee, true, false, time.Now().Add(-24*time.Hour), nil,
			"Administrador PetControl", "Admin", nil, nil, nil, nil, nil, time.Now().Add(-24*time.Hour), nil,
		))

	hub := realtime.NewInternalChatHub()
	server := httptest.NewServer(newChatSocketRouter(serviceUnderTest, hub))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, resp, err := websocket.Dial(ctx, websocketURL(server.URL, contactUserID), &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Authorization": []string{"Bearer " + token},
		},
		Subprotocols: []string{internalChatSocketSubprotocol},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "done")
	}()

	connectedEvent := readSocketEvent(t, ctx, conn)
	require.Equal(t, "chat.connected", connectedEvent["type"])
	snapshotEvent := readSocketEvent(t, ctx, conn)
	require.Equal(t, "chat.presence.snapshot", snapshotEvent["type"])

	payload, err := json.Marshal(map[string]string{"message": "Nova mensagem em tempo real"})
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, messagesURL(server.URL, contactUserID), strings.NewReader(string(payload)))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusCreated, res.StatusCode)

	event := readSocketEvent(t, ctx, conn)
	require.Equal(t, "chat.message.created", event["type"])
	require.Equal(t, companyID.String(), event["company_id"])
	require.Equal(t, contactUserID.String(), event["counterpart_user_id"])

	messagePayload, ok := event["message"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, messageID.String(), messagePayload["id"])
	require.Equal(t, "Nova mensagem em tempo real", messagePayload["body"])
	require.Equal(t, currentUserID.String(), messagePayload["sender_user_id"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminSystemChatHandler_Connect_BroadcastsPresenceUpdates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceUnderTest, mock := adminSystemChatServiceWithMock(t)
	defer mock.Close()

	companyID := newHandlerUUID(t)
	adminUserID := newHandlerUUID(t)
	systemUserID := newHandlerUUID(t)
	adminToken := chatToken(t, companyID, adminUserID, "admin")
	systemToken := chatToken(t, companyID, systemUserID, "system")

	expectParticipantValidation(t, mock, companyID, adminUserID, systemUserID, sqlc.UserRoleTypeAdmin, sqlc.UserRoleTypeSystem)
	expectParticipantValidation(t, mock, companyID, systemUserID, adminUserID, sqlc.UserRoleTypeSystem, sqlc.UserRoleTypeAdmin)

	hub := realtime.NewInternalChatHub()
	server := httptest.NewServer(newChatSocketRouter(serviceUnderTest, hub))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	adminConn, resp, err := websocket.Dial(ctx, websocketURL(server.URL, systemUserID), &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Authorization": []string{"Bearer " + adminToken},
		},
		Subprotocols: []string{internalChatSocketSubprotocol},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	defer func() {
		_ = adminConn.Close(websocket.StatusNormalClosure, "done")
	}()

	require.Equal(t, "chat.connected", readSocketEvent(t, ctx, adminConn)["type"])
	adminSnapshot := readSocketEvent(t, ctx, adminConn)
	require.Equal(t, "chat.presence.snapshot", adminSnapshot["type"])

	systemConn, resp, err := websocket.Dial(ctx, websocketURL(server.URL, adminUserID), &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Authorization": []string{"Bearer " + systemToken},
		},
		Subprotocols: []string{internalChatSocketSubprotocol},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

	require.Equal(t, "chat.connected", readSocketEvent(t, ctx, systemConn)["type"])
	systemSnapshot := readSocketEvent(t, ctx, systemConn)
	require.Equal(t, "chat.presence.snapshot", systemSnapshot["type"])

	adminPresenceUpdated := readSocketEvent(t, ctx, adminConn)
	require.Equal(t, "chat.presence.updated", adminPresenceUpdated["type"])
	presence, ok := adminPresenceUpdated["presence"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, systemUserID.String(), presence["user_id"])
	require.Equal(t, "online", presence["status"])
	require.EqualValues(t, 1, presence["connections"])

	require.NoError(t, systemConn.Close(websocket.StatusNormalClosure, "disconnect"))

	adminPresenceOffline := readSocketEvent(t, ctx, adminConn)
	require.Equal(t, "chat.presence.updated", adminPresenceOffline["type"])
	presence, ok = adminPresenceOffline["presence"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, systemUserID.String(), presence["user_id"])
	require.Equal(t, "offline", presence["status"])
	require.EqualValues(t, 0, presence["connections"])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminSystemChatHandler_Connect_RejectsUnexpectedSocketPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceUnderTest, mock := adminSystemChatServiceWithMock(t)
	defer mock.Close()

	companyID := newHandlerUUID(t)
	currentUserID := newHandlerUUID(t)
	contactUserID := newHandlerUUID(t)
	token := chatToken(t, companyID, currentUserID, "admin")

	expectParticipantValidation(t, mock, companyID, currentUserID, contactUserID, sqlc.UserRoleTypeAdmin, sqlc.UserRoleTypeSystem)

	hub := realtime.NewInternalChatHub()
	server := httptest.NewServer(newChatSocketRouter(serviceUnderTest, hub))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, resp, err := websocket.Dial(ctx, websocketURL(server.URL, contactUserID), &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Authorization": []string{"Bearer " + token},
		},
		Subprotocols: []string{internalChatSocketSubprotocol},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "done")
	}()

	require.Equal(t, "chat.connected", readSocketEvent(t, ctx, conn)["type"])
	require.Equal(t, "chat.presence.snapshot", readSocketEvent(t, ctx, conn)["type"])

	require.NoError(t, wsjson.Write(ctx, conn, map[string]any{
		"type":    "client.message",
		"message": "not allowed",
	}))

	event := readSocketEvent(t, ctx, conn)
	require.Equal(t, "chat.error", event["type"])
	require.Equal(t, companyID.String(), event["company_id"])
	require.Equal(t, contactUserID.String(), event["counterpart_user_id"])
	require.Equal(t, "invalid_socket_payload", event["code"])

	_, _, err = conn.Read(ctx)
	require.Error(t, err)
	require.Equal(t, websocket.StatusPolicyViolation, websocket.CloseStatus(err))

	stats := hub.Stats()
	require.EqualValues(t, 1, stats.InvalidPayloads)
	require.NoError(t, mock.ExpectationsWereMet())
}
