package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/realtime"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type AdminSystemChatHandler struct {
	service                 *service.AdminSystemChatService
	hub                     *realtime.InternalChatHub
	websocketAllowedOrigins []string
}

type createAdminSystemMessageRequest struct {
	Message string `json:"message"`
}

const (
	internalChatSocketSubprotocol = "petcontrol.internal-chat.v1"
	internalChatMaxMessageBytes   = 4096
	internalChatWriteTimeout      = 10 * time.Second
)

func NewAdminSystemChatHandler(
	service *service.AdminSystemChatService,
	hub *realtime.InternalChatHub,
	websocketAllowedOrigins []string,
) *AdminSystemChatHandler {
	return &AdminSystemChatHandler{
		service:                 service,
		hub:                     hub,
		websocketAllowedOrigins: websocketAllowedOrigins,
	}
}

// ListMessages godoc
// @Summary List persisted admin-system chat messages
// @Description Returns the persisted text conversation between the authenticated admin/system user and the selected counterpart in the same tenant.
// @Tags chat
// @Security BearerAuth
// @Produce json
// @Param user_id path string true "Counterpart user id"
// @Success 200 {object} AdminSystemChatMessageListResponseDoc
// @Failure 400 {object} APIErrorResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /chat/system/{user_id}/messages [get]
func (h *AdminSystemChatHandler) ListMessages(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	claims, ok := middleware.GetClaims(c)
	if !ok || claims.UserID == "" {
		middleware.JSONError(c, http.StatusUnauthorized, "auth_claims_required", "auth claims required")
		return
	}

	currentUserID, err := parseUUID(claims.UserID)
	if err != nil {
		middleware.JSONError(c, http.StatusUnauthorized, "invalid_user_context", "invalid user context")
		return
	}

	contactUserID, err := parseUUID(c.Param("user_id"))
	if err != nil {
		middleware.JSONError(c, http.StatusBadRequest, "invalid_contact_user_id", "invalid contact user id")
		return
	}

	items, err := h.service.ListMessages(c.Request.Context(), companyID, currentUserID, contactUserID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "list_admin_system_messages_failed", "failed to list chat messages")
		return
	}

	middleware.JSONData(c, http.StatusOK, mapAdminSystemChatMessages(items))
}

// CreateMessage godoc
// @Summary Create persisted admin-system chat message
// @Description Persists a text message in the admin-system tenant conversation.
// @Tags chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user_id path string true "Counterpart user id"
// @Param payload body AdminSystemChatMessageCreateRequestDoc true "Message payload"
// @Success 201 {object} AdminSystemChatMessageItemResponseDoc
// @Failure 400 {object} APIErrorResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /chat/system/{user_id}/messages [post]
func (h *AdminSystemChatHandler) CreateMessage(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	claims, ok := middleware.GetClaims(c)
	if !ok || claims.UserID == "" {
		middleware.JSONError(c, http.StatusUnauthorized, "auth_claims_required", "auth claims required")
		return
	}

	currentUserID, err := parseUUID(claims.UserID)
	if err != nil {
		middleware.JSONError(c, http.StatusUnauthorized, "invalid_user_context", "invalid user context")
		return
	}

	contactUserID, err := parseUUID(c.Param("user_id"))
	if err != nil {
		middleware.JSONError(c, http.StatusBadRequest, "invalid_contact_user_id", "invalid contact user id")
		return
	}

	var input createAdminSystemMessageRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.JSONError(c, http.StatusBadRequest, "invalid_chat_message_payload", "invalid chat message payload")
		return
	}

	message := strings.TrimSpace(input.Message)
	if message == "" {
		middleware.JSONError(c, http.StatusBadRequest, "chat_message_required", "chat message required")
		return
	}

	item, err := h.service.SendMessage(c.Request.Context(), companyID, currentUserID, contactUserID, message)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "send_admin_system_message_failed", "failed to send chat message")
		return
	}

	h.hub.BroadcastConversationEvent(
		c.Request.Context(),
		uuidToString(companyID),
		uuidToString(currentUserID),
		uuidToString(contactUserID),
		"",
		func(connection realtime.InternalChatConnection) map[string]any {
			return h.newMessageCreatedEvent(connection, item)
		},
	)

	middleware.JSONData(c, http.StatusCreated, mapAdminSystemChatMessage(item))
}

func (h *AdminSystemChatHandler) Connect(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	claims, ok := middleware.GetClaims(c)
	if !ok || claims.UserID == "" {
		middleware.JSONError(c, http.StatusUnauthorized, "auth_claims_required", "auth claims required")
		return
	}

	currentUserID, err := parseUUID(claims.UserID)
	if err != nil {
		middleware.JSONError(c, http.StatusUnauthorized, "invalid_user_context", "invalid user context")
		return
	}

	contactUserID, err := parseUUID(c.Param("user_id"))
	if err != nil {
		middleware.JSONError(c, http.StatusBadRequest, "invalid_contact_user_id", "invalid contact user id")
		return
	}

	if _, _, err := h.service.ResolveParticipants(c.Request.Context(), companyID, currentUserID, contactUserID); err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "connect_admin_system_chat_failed", "failed to connect chat socket")
		return
	}

	socket, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		Subprotocols:   []string{internalChatSocketSubprotocol},
		OriginPatterns: h.websocketAllowedOrigins,
	})
	if err != nil {
		return
	}

	if socket.Subprotocol() == "" {
		_ = socket.Close(websocket.StatusPolicyViolation, "subprotocol required")
		return
	}

	socket.SetReadLimit(internalChatMaxMessageBytes)

	connectionID := uuid.NewString()
	connection := realtime.InternalChatConnection{
		ID:                connectionID,
		CompanyID:         uuidToString(companyID),
		UserID:            uuidToString(currentUserID),
		CounterpartUserID: uuidToString(contactUserID),
		UserRole:          claims.Role,
		ConnectedAt:       time.Now().UTC(),
		Socket:            socket,
	}

	_, statusChanged := h.hub.Register(connection)
	defer func() {
		disconnectedConnection, presence, becameOffline := h.hub.Unregister(connectionID)
		if becameOffline {
			h.hub.BroadcastConversationEvent(
				context.Background(),
				disconnectedConnection.CompanyID,
				disconnectedConnection.UserID,
				disconnectedConnection.CounterpartUserID,
				"",
				func(target realtime.InternalChatConnection) map[string]any {
					return h.newPresenceUpdatedEvent(target, presence)
				},
			)
		}
	}()
	defer func() {
		_ = socket.Close(websocket.StatusNormalClosure, "connection closed")
	}()

	if err := h.writeSocketEvent(c.Request.Context(), socket, map[string]any{
		"type":                "chat.connected",
		"company_id":          connection.CompanyID,
		"counterpart_user_id": connection.CounterpartUserID,
		"emitted_at":          connection.ConnectedAt.Format(time.RFC3339),
		"connection_id":       connection.ID,
		"viewer_user_id":      connection.UserID,
		"viewer_role":         connection.UserRole,
	}); err != nil {
		_ = socket.Close(websocket.StatusInternalError, "failed to initialize connection")
		return
	}

	if err := h.writeSocketEvent(c.Request.Context(), socket, h.newPresenceSnapshotEvent(connection)); err != nil {
		_ = socket.Close(websocket.StatusInternalError, "failed to initialize presence snapshot")
		return
	}

	if statusChanged {
		h.hub.BroadcastConversationEvent(
			c.Request.Context(),
			connection.CompanyID,
			connection.UserID,
			connection.CounterpartUserID,
			connection.ID,
			func(target realtime.InternalChatConnection) map[string]any {
				return h.newPresenceUpdatedEvent(target, h.hub.Presence(connection.CompanyID, connection.UserID))
			},
		)
	}

	ctx := socket.CloseRead(c.Request.Context())
	<-ctx.Done()
}

func (h *AdminSystemChatHandler) writeSocketEvent(ctx context.Context, socket *websocket.Conn, payload map[string]any) error {
	writeCtx, cancel := context.WithTimeout(ctx, internalChatWriteTimeout)
	defer cancel()

	return wsjson.Write(writeCtx, socket, payload)
}

func (h *AdminSystemChatHandler) newMessageCreatedEvent(
	connection realtime.InternalChatConnection,
	item service.AdminSystemChatMessage,
) map[string]any {
	return map[string]any{
		"type":                "chat.message.created",
		"company_id":          uuidToString(item.CompanyID),
		"counterpart_user_id": connection.CounterpartUserID,
		"emitted_at":          formatTimestamptz(item.CreatedAt),
		"message":             mapAdminSystemChatMessage(item),
	}
}

func (h *AdminSystemChatHandler) newPresenceSnapshotEvent(connection realtime.InternalChatConnection) map[string]any {
	presences := h.hub.ConversationSnapshot(connection.CompanyID, connection.UserID, connection.CounterpartUserID)
	items := make([]map[string]any, 0, len(presences))
	for _, presence := range presences {
		items = append(items, mapInternalChatPresence(presence))
	}

	return map[string]any{
		"type":                "chat.presence.snapshot",
		"company_id":          connection.CompanyID,
		"counterpart_user_id": connection.CounterpartUserID,
		"emitted_at":          time.Now().UTC().Format(time.RFC3339),
		"presences":           items,
	}
}

func (h *AdminSystemChatHandler) newPresenceUpdatedEvent(
	connection realtime.InternalChatConnection,
	presence realtime.InternalChatPresence,
) map[string]any {
	return map[string]any{
		"type":                "chat.presence.updated",
		"company_id":          connection.CompanyID,
		"counterpart_user_id": connection.CounterpartUserID,
		"emitted_at":          time.Now().UTC().Format(time.RFC3339),
		"presence":            mapInternalChatPresence(presence),
	}
}

func mapInternalChatPresence(presence realtime.InternalChatPresence) map[string]any {
	return map[string]any{
		"user_id":         presence.UserID,
		"status":          presence.Status,
		"connections":     presence.Connections,
		"last_changed_at": presence.LastChangedAt.Format(time.RFC3339),
	}
}
