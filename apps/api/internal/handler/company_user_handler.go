package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func (h *CompanyUserHandler) List(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "company context required"})
		return
	}

	users, err := h.service.ListCompanyUsers(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list company users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h *CompanyUserHandler) Create(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "company context required"})
		return
	}

	var req createCompanyUserRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid request body"})
		return
	}

	userID, err := parseUUID(req.UserID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid user_id"})
		return
	}

	created, err := h.service.CreateCompanyUser(c.Request.Context(), sqlc.CreateCompanyUserParams{
		CompanyID: companyID,
		UserID:    userID,
		IsOwner:   req.IsOwner,
		IsActive:  true,
	})
	if err != nil {
		c.JSON(apperror.HTTPStatus(err), gin.H{"error": "failed to create company user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": created})
}

func (h *CompanyUserHandler) Deactivate(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "company context required"})
		return
	}

	userID, err := parseUUID(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid user_id"})
		return
	}

	if err := h.service.DeactivateCompanyUser(c.Request.Context(), companyID, userID); err != nil {
		c.JSON(apperror.HTTPStatus(err), gin.H{"error": "failed to deactivate company user"})
		return
	}

	c.Status(http.StatusNoContent)
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
