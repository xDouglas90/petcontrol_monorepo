package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type CompanyHandler struct {
	service        *service.CompanyService
	uploadResolver companyUploadResolver
}

type updateCompanyRequest struct {
	Name           *string `json:"name"`
	FantasyName    *string `json:"fantasy_name"`
	FoundationDate *string `json:"foundation_date"`
	LogoURL        *string `json:"logo_url"`
	UploadKey      *string `json:"upload_object_key"`
}

type companyUploadResolver interface {
	ResolveObjectKey(ctx context.Context, resource string, field string, objectKey string) (string, error)
}

func NewCompanyHandler(service *service.CompanyService, uploadResolver ...companyUploadResolver) *CompanyHandler {
	handler := &CompanyHandler{service: service}
	if len(uploadResolver) > 0 {
		handler.uploadResolver = uploadResolver[0]
	}
	return handler
}

func (h *CompanyHandler) List(c *gin.Context) {
	companies, err := h.service.ListCompanies(c.Request.Context())
	if err != nil {
		middleware.JSONError(c, http.StatusInternalServerError, "list_companies_failed", "failed to list companies")
		return
	}

	middleware.JSONData(c, http.StatusOK, companies)
}

// Current godoc
// @Summary Get current company
// @Description Returns the company resolved from the authenticated tenant context.
// @Tags companies
// @Security BearerAuth
// @Produce json
// @Success 200 {object} CompanyItemResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /companies/current [get]
func (h *CompanyHandler) Current(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	company, err := h.service.GetCurrentCompany(c.Request.Context(), companyID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_current_company_failed", "failed to get current company")
		return
	}

	middleware.JSONData(c, http.StatusOK, company)
}

// Update godoc
// @Summary Update current company
// @Description Updates the authenticated company's data.
// @Tags companies
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CompanyUpdateRequestDoc true "Update payload"
// @Success 200 {object} CompanyItemResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /companies/current [patch]
func (h *CompanyHandler) Update(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	claims, ok := middleware.GetClaims(c)
	if !ok {
		middleware.JSONError(c, http.StatusUnauthorized, "auth_claims_required", "auth claims required")
		return
	}
	if claims.Role != string(sqlc.UserRoleTypeAdmin) {
		middleware.JSONError(c, http.StatusForbidden, "company_admin_required", "company admin required")
		return
	}

	var req updateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_request_body", "invalid request body")
		return
	}

	logoURL, err := h.resolveOptionalUploadObjectKey(c.Request.Context(), req.UploadKey, req.LogoURL)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "invalid_upload_object_key", "invalid upload_object_key")
		return
	}

	// Aqui poderíamos ter mais validações se necessário
	item, err := h.service.UpdateCompany(c.Request.Context(), sqlc.UpdateCompanyParams{
		ID:          companyID,
		Name:        textPointer(req.Name),
		FantasyName: textPointer(req.FantasyName),
		LogoURL:     textPointer(logoURL),
		// Outros campos omitidos por brevidade ou por não serem editáveis via Current
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "update_company_failed", "failed to update company")
		return
	}

	middleware.JSONData(c, http.StatusOK, item)
}

func (h *CompanyHandler) resolveUploadObjectKey(ctx context.Context, objectKey string, fallback string) (string, error) {
	trimmedKey := strings.TrimSpace(objectKey)
	if trimmedKey == "" {
		return fallback, nil
	}
	if h.uploadResolver == nil {
		return "", apperror.ErrServiceUnavailable
	}
	return h.uploadResolver.ResolveObjectKey(ctx, "companies", "logo_url", trimmedKey)
}

func (h *CompanyHandler) resolveOptionalUploadObjectKey(ctx context.Context, objectKey *string, fallback *string) (*string, error) {
	if objectKey == nil {
		return parseOptionalTrimmed(fallback), nil
	}

	resolved, err := h.resolveUploadObjectKey(ctx, *objectKey, "")
	if err != nil {
		return nil, err
	}
	return &resolved, nil
}
