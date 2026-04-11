package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/pagination"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type ServiceHandler struct {
	service *service.ServiceService
}

type createServiceRequest struct {
	TypeName     string `json:"type_name"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Notes        string `json:"notes"`
	Price        string `json:"price"`
	DiscountRate string `json:"discount_rate"`
	ImageURL     string `json:"image_url"`
	IsActive     *bool  `json:"is_active"`
}

type updateServiceRequest struct {
	TypeName     *string `json:"type_name"`
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	Notes        *string `json:"notes"`
	Price        *string `json:"price"`
	DiscountRate *string `json:"discount_rate"`
	ImageURL     *string `json:"image_url"`
	IsActive     *bool   `json:"is_active"`
}

type serviceResponse struct {
	ID           string  `json:"id"`
	TypeID       string  `json:"type_id"`
	TypeName     string  `json:"type_name"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	Notes        *string `json:"notes,omitempty"`
	Price        string  `json:"price"`
	DiscountRate string  `json:"discount_rate"`
	ImageURL     *string `json:"image_url,omitempty"`
	IsActive     bool    `json:"is_active"`
}

func NewServiceHandler(service *service.ServiceService) *ServiceHandler {
	return &ServiceHandler{service: service}
}

// List godoc
// @Summary List services
// @Description Returns services from the authenticated tenant.
// @Tags services
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ServiceListResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /services [get]
func (h *ServiceHandler) List(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	params := pagination.ParseParams(c)
	items, err := h.service.ListServicesByCompanyID(c.Request.Context(), companyID, params)
	if err != nil {
		middleware.JSONError(c, http.StatusInternalServerError, "list_services_failed", "failed to list services")
		return
	}

	total := 0
	if len(items) > 0 {
		total = int(items[0].TotalCount)
	}

	middleware.JSONPaginated(c, http.StatusOK, mapServiceList(items), pagination.NewMeta(total, params.Page, params.Limit))
}

// GetByID godoc
// @Summary Get service by ID
// @Description Returns a single service from the authenticated tenant.
// @Tags services
// @Security BearerAuth
// @Produce json
// @Param id path string true "Service ID"
// @Success 200 {object} ServiceItemResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /services/{id} [get]
func (h *ServiceHandler) GetByID(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	serviceID, err := parseUUID(c.Param("id"))
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_service_id", "invalid service id")
		return
	}

	item, err := h.service.GetServiceByID(c.Request.Context(), companyID, serviceID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_service_failed", "failed to get service")
		return
	}

	middleware.JSONData(c, http.StatusOK, mapServiceItem(item))
}

// Create godoc
// @Summary Create service
// @Description Creates a service for the authenticated tenant.
// @Tags services
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ServiceCreateRequestDoc true "Service payload"
// @Success 201 {object} ServiceItemResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 409 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /services [post]
func (h *ServiceHandler) Create(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	var req createServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_request_body", "invalid request body")
		return
	}

	price, err := parseRequiredNumeric(req.Price)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_price", "invalid price")
		return
	}

	discountRate, err := parseOptionalNumeric(req.DiscountRate)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_discount_rate", "invalid discount_rate")
		return
	}

	item, err := h.service.CreateService(c.Request.Context(), service.CreateServiceInput{
		CompanyID:    companyID,
		TypeName:     strings.TrimSpace(req.TypeName),
		Title:        strings.TrimSpace(req.Title),
		Description:  strings.TrimSpace(req.Description),
		Notes:        textValue(strings.TrimSpace(req.Notes)),
		Price:        price,
		DiscountRate: discountRate,
		ImageURL:     textValue(strings.TrimSpace(req.ImageURL)),
		IsActive:     req.IsActive == nil || *req.IsActive,
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "create_service_failed", "failed to create service")
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionCreate,
		EntityTable: "services",
		EntityID:    item.ID,
		CompanyID:   companyID,
		OldData:     nil,
		NewData:     item,
	})

	middleware.JSONData(c, http.StatusCreated, mapServiceItem(item))
}

// Update godoc
// @Summary Update service
// @Description Updates a service from the authenticated tenant.
// @Tags services
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Service ID"
// @Param request body ServiceUpdateRequestDoc true "Service payload"
// @Success 200 {object} ServiceItemResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 409 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /services/{id} [put]
func (h *ServiceHandler) Update(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	serviceID, err := parseUUID(c.Param("id"))
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_service_id", "invalid service id")
		return
	}

	before, err := h.service.GetServiceByID(c.Request.Context(), companyID, serviceID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_service_failed", "failed to get service")
		return
	}

	var req updateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_request_body", "invalid request body")
		return
	}

	if !hasServiceUpdatePayload(req) {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "empty_update_payload", "at least one field must be provided")
		return
	}

	price, err := parseOptionalNumericPointer(req.Price)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_price", "invalid price")
		return
	}

	discountRate, err := parseOptionalNumericPointer(req.DiscountRate)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_discount_rate", "invalid discount_rate")
		return
	}

	item, err := h.service.UpdateService(c.Request.Context(), service.UpdateServiceInput{
		CompanyID:    companyID,
		ServiceID:    serviceID,
		TypeName:     parseOptionalTrimmed(req.TypeName),
		Title:        parseOptionalTrimmed(req.Title),
		Description:  parseOptionalTrimmed(req.Description),
		Notes:        parseOptionalTrimmed(req.Notes),
		Price:        price,
		DiscountRate: discountRate,
		ImageURL:     parseOptionalTrimmed(req.ImageURL),
		IsActive:     req.IsActive,
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "update_service_failed", "failed to update service")
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionUpdate,
		EntityTable: "services",
		EntityID:    item.ID,
		CompanyID:   companyID,
		OldData:     before,
		NewData:     item,
	})

	middleware.JSONData(c, http.StatusOK, mapServiceItem(item))
}

// Delete godoc
// @Summary Delete service
// @Description Deactivates a service from the authenticated tenant.
// @Tags services
// @Security BearerAuth
// @Param id path string true "Service ID"
// @Success 204
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /services/{id} [delete]
func (h *ServiceHandler) Delete(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	serviceID, err := parseUUID(c.Param("id"))
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_service_id", "invalid service id")
		return
	}

	before, err := h.service.GetServiceByID(c.Request.Context(), companyID, serviceID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_service_failed", "failed to get service")
		return
	}

	if err := h.service.DeactivateService(c.Request.Context(), companyID, serviceID); err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "delete_service_failed", "failed to delete service")
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionDelete,
		EntityTable: "services",
		EntityID:    before.ID,
		CompanyID:   companyID,
		OldData:     before,
		NewData: gin.H{
			"id":      before.ID,
			"deleted": true,
		},
	})

	c.Status(http.StatusNoContent)
}

func mapServiceList(items []sqlc.ListServicesByCompanyIDRow) []serviceResponse {
	mapped := make([]serviceResponse, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, serviceResponse{
			ID:           item.ID.String(),
			TypeID:       item.TypeID.String(),
			TypeName:     item.TypeName,
			Title:        item.Title,
			Description:  item.Description,
			Notes:        nullableText(item.Notes),
			Price:        numericToString(item.Price),
			DiscountRate: numericToString(item.DiscountRate),
			ImageURL:     nullableText(item.ImageUrl),
			IsActive:     item.IsActive,
		})
	}
	return mapped
}

func mapServiceItem(item sqlc.GetServiceByIDAndCompanyIDRow) serviceResponse {
	return serviceResponse{
		ID:           item.ID.String(),
		TypeID:       item.TypeID.String(),
		TypeName:     item.TypeName,
		Title:        item.Title,
		Description:  item.Description,
		Notes:        nullableText(item.Notes),
		Price:        numericToString(item.Price),
		DiscountRate: numericToString(item.DiscountRate),
		ImageURL:     nullableText(item.ImageUrl),
		IsActive:     item.IsActive,
	}
}

func parseRequiredNumeric(raw string) (pgtype.Numeric, error) {
	var value pgtype.Numeric
	err := value.Scan(strings.TrimSpace(raw))
	return value, err
}

func parseOptionalNumeric(raw string) (pgtype.Numeric, error) {
	if strings.TrimSpace(raw) == "" {
		return parseRequiredNumeric("0.00")
	}
	return parseRequiredNumeric(raw)
}

func parseOptionalNumericPointer(raw *string) (*pgtype.Numeric, error) {
	if raw == nil {
		return nil, nil
	}
	value, err := parseRequiredNumeric(*raw)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func numericToString(value pgtype.Numeric) string {
	if !value.Valid {
		return ""
	}
	raw, err := value.Value()
	if err != nil || raw == nil {
		return ""
	}
	rendered, _ := raw.(string)
	return rendered
}

func nullableText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func hasServiceUpdatePayload(req updateServiceRequest) bool {
	return req.TypeName != nil ||
		req.Title != nil ||
		req.Description != nil ||
		req.Notes != nil ||
		req.Price != nil ||
		req.DiscountRate != nil ||
		req.ImageURL != nil ||
		req.IsActive != nil
}
