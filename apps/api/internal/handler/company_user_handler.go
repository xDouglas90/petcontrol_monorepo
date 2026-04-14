package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type CompanyUserHandler struct {
	service *service.CompanyUserService
}

type createCompanyUserRequest struct {
	UserID  string `json:"user_id"`
	IsOwner bool   `json:"is_owner"`
}

func NewCompanyUserHandler(service *service.CompanyUserService) *CompanyUserHandler {
	return &CompanyUserHandler{service: service}
}

// List godoc
// @Summary List company users
// @Description Lists users linked to the authenticated tenant company.
// @Tags company_users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} CompanyUserListResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /company-users [get]
func (h *CompanyUserHandler) List(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, 403, "company_context_required", "company context required")
		return
	}

	users, err := h.service.ListCompanyUsers(c.Request.Context(), companyID)
	if err != nil {
		middleware.JSONError(c, 500, "list_company_users_failed", "failed to list company users")
		return
	}

	middleware.JSONData(c, 200, users)
}

// Create godoc
// @Summary Create company user link
// @Description Links an existing user to the authenticated tenant company. Requires company owner access.
// @Tags company_users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CompanyUserCreateRequestDoc true "Company user payload"
// @Success 201 {object} CompanyUserItemResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /company-users [post]
func (h *CompanyUserHandler) Create(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, 403, "company_context_required", "company context required")
		return
	}

	var req createCompanyUserRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == "" {
		middleware.JSONError(c, 422, "invalid_request_body", "invalid request body")
		return
	}

	userID, err := parseUUID(req.UserID)
	if err != nil {
		middleware.JSONError(c, 422, "invalid_user_id", "invalid user_id")
		return
	}

	kind := sqlc.UserKindEmployee
	if req.IsOwner {
		kind = sqlc.UserKindOwner
	}

	created, err := h.service.CreateCompanyUser(c.Request.Context(), sqlc.CreateCompanyUserParams{
		CompanyID: companyID,
		UserID:    userID,
		Kind:      kind,
		IsOwner:   req.IsOwner,
		IsActive:  pgtype.Bool{Bool: true, Valid: true},
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "create_company_user_failed", "failed to create company user")
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionCreate,
		EntityTable: "company_users",
		EntityID:    created.ID,
		CompanyID:   companyID,
		OldData:     nil,
		NewData:     created,
	})

	middleware.JSONData(c, 201, created)
}

// Deactivate godoc
// @Summary Deactivate company user link
// @Description Deactivates a user link for the authenticated tenant company. Requires company owner access.
// @Tags company_users
// @Security BearerAuth
// @Produce json
// @Param user_id path string true "User ID"
// @Success 204
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /company-users/{user_id} [delete]
func (h *CompanyUserHandler) Deactivate(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, 403, "company_context_required", "company context required")
		return
	}

	userID, err := parseUUID(c.Param("user_id"))
	if err != nil {
		middleware.JSONError(c, 422, "invalid_user_id", "invalid user_id")
		return
	}

	before, err := h.service.GetCompanyUser(c.Request.Context(), companyID, userID)
	if err != nil {
		status := apperror.HTTPStatus(err)
		code := "get_company_user_failed"
		if errors.Is(err, apperror.ErrNotFound) || errors.Is(err, pgx.ErrNoRows) {
			code = "company_user_not_found"
		}
		middleware.JSONError(c, status, code, "failed to load company user")
		return
	}

	if err := h.service.DeactivateCompanyUser(c.Request.Context(), companyID, userID); err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "deactivate_company_user_failed", "failed to deactivate company user")
		return
	}

	after, err := h.service.GetCompanyUser(c.Request.Context(), companyID, userID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_company_user_failed", "failed to load company user")
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionDeactivate,
		EntityTable: "company_users",
		EntityID:    before.ID,
		CompanyID:   companyID,
		OldData:     before,
		NewData:     after,
	})

	c.Status(204)
}

func parseUUID(raw string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return pgtype.UUID{}, err
	}

	var out pgtype.UUID
	copy(out.Bytes[:], parsed[:])
	out.Valid = true
	return out, nil
}
