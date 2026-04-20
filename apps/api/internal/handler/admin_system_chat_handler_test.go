package handler

import (
	"context"
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
	return router
}

func websocketURL(baseURL string, userID pgtype.UUID) string {
	return "ws" + strings.TrimPrefix(baseURL, "http") + "/api/v1/chat/system/" + userID.String() + "/ws"
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
	require.NoError(t, wsjson.Read(ctx, conn, &event))
	require.Equal(t, "chat.connected", event["type"])
	require.Equal(t, companyID.String(), event["company_id"])
	require.Equal(t, contactUserID.String(), event["counterpart_user_id"])
	require.Equal(t, currentUserID.String(), event["viewer_user_id"])
	require.Equal(t, "admin", event["viewer_role"])

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
