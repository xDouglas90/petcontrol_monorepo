package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type AdminSystemChatHandler struct {
	service *service.AdminSystemChatService
}

type createAdminSystemMessageRequest struct {
	Message string `json:"message"`
}

func NewAdminSystemChatHandler(service *service.AdminSystemChatService) *AdminSystemChatHandler {
	return &AdminSystemChatHandler{service: service}
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

	middleware.JSONData(c, http.StatusCreated, mapAdminSystemChatMessage(item))
}
