package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/pagination"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type ClientHandler struct {
	service *service.ClientService
}

type createClientRequest struct {
	FullName       string `json:"full_name"`
	ShortName      string `json:"short_name"`
	GenderIdentity string `json:"gender_identity"`
	MaritalStatus  string `json:"marital_status"`
	BirthDate      string `json:"birth_date"`
	CPF            string `json:"cpf"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Cellphone      string `json:"cellphone"`
	HasWhatsapp    bool   `json:"has_whatsapp"`
	ClientSince    string `json:"client_since"`
	Notes          string `json:"notes"`
}

type updateClientRequest struct {
	FullName       *string `json:"full_name"`
	ShortName      *string `json:"short_name"`
	GenderIdentity *string `json:"gender_identity"`
	MaritalStatus  *string `json:"marital_status"`
	BirthDate      *string `json:"birth_date"`
	CPF            *string `json:"cpf"`
	Email          *string `json:"email"`
	Phone          *string `json:"phone"`
	Cellphone      *string `json:"cellphone"`
	HasWhatsapp    *bool   `json:"has_whatsapp"`
	ClientSince    *string `json:"client_since"`
	Notes          *string `json:"notes"`
}

func NewClientHandler(service *service.ClientService) *ClientHandler {
	return &ClientHandler{service: service}
}

// List godoc
// @Summary List clients
// @Description Returns clients from the authenticated tenant.
// @Tags clients
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ClientListResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /clients [get]
func (h *ClientHandler) List(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	params := pagination.ParseParams(c)
	items, err := h.service.ListClientsByCompanyID(c.Request.Context(), companyID, params)
	if err != nil {
		middleware.JSONError(c, http.StatusInternalServerError, "list_clients_failed", "failed to list clients")
		return
	}

	total := 0
	if len(items) > 0 {
		total = int(items[0].TotalCount)
	}

	middleware.JSONPaginated(c, http.StatusOK, items, pagination.NewMeta(total, params.Page, params.Limit))
}

// GetByID godoc
// @Summary Get client by ID
// @Description Returns a single client from the authenticated tenant.
// @Tags clients
// @Security BearerAuth
// @Produce json
// @Param id path string true "Client ID"
// @Success 200 {object} ClientItemResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /clients/{id} [get]
func (h *ClientHandler) GetByID(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	clientID, err := parseUUID(c.Param("id"))
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_client_id", "invalid client id")
		return
	}

	item, err := h.service.GetClientByID(c.Request.Context(), companyID, clientID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_client_failed", "failed to get client")
		return
	}

	middleware.JSONData(c, http.StatusOK, item)
}

// Create godoc
// @Summary Create client
// @Description Creates a client for the authenticated tenant.
// @Tags clients
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ClientCreateRequestDoc true "Client payload"
// @Success 201 {object} ClientItemResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 409 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /clients [post]
func (h *ClientHandler) Create(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	var req createClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_request_body", "invalid request body")
		return
	}

	birthDate, err := parseRequiredDate(req.BirthDate)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_birth_date", "invalid birth_date")
		return
	}

	clientSince, err := parseOptionalDate(req.ClientSince)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_client_since", "invalid client_since")
		return
	}

	genderIdentity, err := parseGenderIdentity(req.GenderIdentity)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_gender_identity", "invalid gender_identity")
		return
	}

	maritalStatus, err := parseMaritalStatus(req.MaritalStatus)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_marital_status", "invalid marital_status")
		return
	}

	item, err := h.service.CreateClient(c.Request.Context(), service.CreateClientInput{
		CompanyID:      companyID,
		FullName:       strings.TrimSpace(req.FullName),
		ShortName:      strings.TrimSpace(req.ShortName),
		GenderIdentity: genderIdentity,
		MaritalStatus:  maritalStatus,
		BirthDate:      birthDate,
		CPF:            onlyDigits(req.CPF),
		Email:          strings.TrimSpace(req.Email),
		Phone:          textValue(strings.TrimSpace(req.Phone)),
		Cellphone:      strings.TrimSpace(req.Cellphone),
		HasWhatsapp:    req.HasWhatsapp,
		ClientSince:    clientSince,
		Notes:          textValue(strings.TrimSpace(req.Notes)),
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "create_client_failed", "failed to create client")
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionCreate,
		EntityTable: "clients",
		EntityID:    item.ID,
		CompanyID:   companyID,
		OldData:     nil,
		NewData:     item,
	})

	middleware.JSONData(c, http.StatusCreated, item)
}

// Update godoc
// @Summary Update client
// @Description Updates a client from the authenticated tenant.
// @Tags clients
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Client ID"
// @Param request body ClientUpdateRequestDoc true "Client payload"
// @Success 200 {object} ClientItemResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 409 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /clients/{id} [put]
func (h *ClientHandler) Update(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	clientID, err := parseUUID(c.Param("id"))
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_client_id", "invalid client id")
		return
	}

	before, err := h.service.GetClientByID(c.Request.Context(), companyID, clientID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_client_failed", "failed to get client")
		return
	}

	var req updateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_request_body", "invalid request body")
		return
	}

	if !hasClientUpdatePayload(req) {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "empty_update_payload", "at least one field must be provided")
		return
	}

	birthDate, err := parseOptionalDatePointer(req.BirthDate)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_birth_date", "invalid birth_date")
		return
	}

	clientSince, err := parseOptionalDatePointer(req.ClientSince)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_client_since", "invalid client_since")
		return
	}

	genderIdentity, err := parseOptionalGenderIdentity(req.GenderIdentity)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_gender_identity", "invalid gender_identity")
		return
	}

	maritalStatus, err := parseOptionalMaritalStatus(req.MaritalStatus)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_marital_status", "invalid marital_status")
		return
	}

	item, err := h.service.UpdateClient(c.Request.Context(), service.UpdateClientInput{
		CompanyID:      companyID,
		ClientID:       clientID,
		FullName:       parseOptionalTrimmed(req.FullName),
		ShortName:      parseOptionalTrimmed(req.ShortName),
		GenderIdentity: genderIdentity,
		MaritalStatus:  maritalStatus,
		BirthDate:      birthDate,
		CPF:            sanitizeOptionalDigits(req.CPF),
		Email:          parseOptionalTrimmed(req.Email),
		Phone:          parseOptionalTrimmed(req.Phone),
		Cellphone:      parseOptionalTrimmed(req.Cellphone),
		HasWhatsapp:    req.HasWhatsapp,
		ClientSince:    clientSince,
		Notes:          parseOptionalTrimmed(req.Notes),
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "update_client_failed", "failed to update client")
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionUpdate,
		EntityTable: "clients",
		EntityID:    item.ID,
		CompanyID:   companyID,
		OldData:     before,
		NewData:     item,
	})

	middleware.JSONData(c, http.StatusOK, item)
}

// Delete godoc
// @Summary Deactivate client
// @Description Deactivates the tenant association for a client.
// @Tags clients
// @Security BearerAuth
// @Param id path string true "Client ID"
// @Success 204
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /clients/{id} [delete]
func (h *ClientHandler) Delete(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	clientID, err := parseUUID(c.Param("id"))
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_client_id", "invalid client id")
		return
	}

	before, err := h.service.GetClientByID(c.Request.Context(), companyID, clientID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_client_failed", "failed to get client")
		return
	}

	if err := h.service.DeactivateClient(c.Request.Context(), companyID, clientID); err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "delete_client_failed", "failed to deactivate client")
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionDeactivate,
		EntityTable: "clients",
		EntityID:    before.ID,
		CompanyID:   companyID,
		OldData:     before,
		NewData: gin.H{
			"id":          before.ID,
			"is_active":   false,
			"deactivated": true,
		},
	})

	c.Status(http.StatusNoContent)
}

func parseRequiredDate(raw string) (pgtype.Date, error) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return pgtype.Date{}, err
	}
	return pgtype.Date{Time: parsed, Valid: true}, nil
}

func parseOptionalDate(raw string) (pgtype.Date, error) {
	if strings.TrimSpace(raw) == "" {
		return pgtype.Date{}, nil
	}
	return parseRequiredDate(raw)
}

func parseOptionalDatePointer(raw *string) (*pgtype.Date, error) {
	if raw == nil {
		return nil, nil
	}
	if strings.TrimSpace(*raw) == "" {
		value := pgtype.Date{}
		return &value, nil
	}
	parsed, err := parseRequiredDate(*raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseGenderIdentity(raw string) (sqlc.GenderIdentity, error) {
	value := sqlc.GenderIdentity(strings.TrimSpace(raw))
	switch value {
	case sqlc.GenderIdentityManCisgender,
		sqlc.GenderIdentityWomanCisgender,
		sqlc.GenderIdentityTransgender,
		sqlc.GenderIdentityNonBinary,
		sqlc.GenderIdentityGenderFluid,
		sqlc.GenderIdentityGenderQueer,
		sqlc.GenderIdentityAgender,
		sqlc.GenderIdentityGenderNonConforming,
		sqlc.GenderIdentityNotToExpose:
		return value, nil
	default:
		return "", apperror.ErrUnprocessableEntity
	}
}

func parseOptionalGenderIdentity(raw *string) (*sqlc.GenderIdentity, error) {
	if raw == nil {
		return nil, nil
	}
	value, err := parseGenderIdentity(*raw)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func parseMaritalStatus(raw string) (sqlc.MaritalStatus, error) {
	value := sqlc.MaritalStatus(strings.TrimSpace(raw))
	switch value {
	case sqlc.MaritalStatusSingle,
		sqlc.MaritalStatusMarried,
		sqlc.MaritalStatusDivorced,
		sqlc.MaritalStatusWidowed,
		sqlc.MaritalStatusSeparated:
		return value, nil
	default:
		return "", apperror.ErrUnprocessableEntity
	}
}

func parseOptionalMaritalStatus(raw *string) (*sqlc.MaritalStatus, error) {
	if raw == nil {
		return nil, nil
	}
	value, err := parseMaritalStatus(*raw)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func onlyDigits(raw string) string {
	var out strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			out.WriteRune(r)
		}
	}
	return out.String()
}

func sanitizeOptionalDigits(raw *string) *string {
	if raw == nil {
		return nil
	}
	value := onlyDigits(*raw)
	return &value
}

func hasClientUpdatePayload(req updateClientRequest) bool {
	return req.FullName != nil ||
		req.ShortName != nil ||
		req.GenderIdentity != nil ||
		req.MaritalStatus != nil ||
		req.BirthDate != nil ||
		req.CPF != nil ||
		req.Email != nil ||
		req.Phone != nil ||
		req.Cellphone != nil ||
		req.HasWhatsapp != nil ||
		req.ClientSince != nil ||
		req.Notes != nil
}

func textValue(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}
