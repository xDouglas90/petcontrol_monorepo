package handler

import (
	"net/http"
	"strconv"
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
	TypeName     string                     `json:"type_name"`
	Title        string                     `json:"title"`
	Description  string                     `json:"description"`
	Notes        string                     `json:"notes"`
	Price        string                     `json:"price"`
	DiscountRate string                     `json:"discount_rate"`
	ImageURL     string                     `json:"image_url"`
	IsActive     *bool                      `json:"is_active"`
	SubServices  []serviceSubServiceRequest `json:"sub_services"`
}

type updateServiceRequest struct {
	TypeName     *string                     `json:"type_name"`
	Title        *string                     `json:"title"`
	Description  *string                     `json:"description"`
	Notes        *string                     `json:"notes"`
	Price        *string                     `json:"price"`
	DiscountRate *string                     `json:"discount_rate"`
	ImageURL     *string                     `json:"image_url"`
	IsActive     *bool                       `json:"is_active"`
	SubServices  *[]serviceSubServiceRequest `json:"sub_services"`
}

type serviceResponse struct {
	ID                string                      `json:"id"`
	TypeID            string                      `json:"type_id"`
	TypeName          string                      `json:"type_name"`
	Title             string                      `json:"title"`
	Description       string                      `json:"description"`
	Notes             *string                     `json:"notes,omitempty"`
	Price             string                      `json:"price"`
	DiscountRate      string                      `json:"discount_rate"`
	ImageURL          *string                     `json:"image_url,omitempty"`
	IsActive          bool                        `json:"is_active"`
	SubServicesCount  int64                       `json:"sub_services_count"`
	AverageTimesCount int64                       `json:"average_times_count"`
	SubServices       []serviceSubServiceResponse `json:"sub_services,omitempty"`
}

type serviceSubServiceRequest struct {
	TypeName     string                      `json:"type_name"`
	Title        string                      `json:"title"`
	Description  string                      `json:"description"`
	Notes        string                      `json:"notes"`
	Price        string                      `json:"price"`
	DiscountRate string                      `json:"discount_rate"`
	ImageURL     string                      `json:"image_url"`
	IsActive     *bool                       `json:"is_active"`
	AverageTimes []serviceAverageTimeRequest `json:"average_times"`
}

type serviceAverageTimeRequest struct {
	PetSize            string `json:"pet_size"`
	PetKind            string `json:"pet_kind"`
	PetTemperament     string `json:"pet_temperament"`
	AverageTimeMinutes int    `json:"average_time_minutes"`
}

type serviceSubServiceResponse struct {
	ID           string                       `json:"id"`
	TypeID       string                       `json:"type_id"`
	Title        string                       `json:"title"`
	Description  string                       `json:"description"`
	Notes        *string                      `json:"notes,omitempty"`
	Price        string                       `json:"price"`
	DiscountRate string                       `json:"discount_rate"`
	ImageURL     *string                      `json:"image_url,omitempty"`
	IsActive     bool                         `json:"is_active"`
	AverageTimes []serviceAverageTimeResponse `json:"average_times"`
}

type serviceAverageTimeResponse struct {
	ID                 string `json:"id"`
	PetSize            string `json:"pet_size"`
	PetKind            string `json:"pet_kind"`
	PetTemperament     string `json:"pet_temperament"`
	AverageTimeMinutes int16  `json:"average_time_minutes"`
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
// @Param search query string false "Search by title or description"
// @Param type_name query string false "Filter by service type name"
// @Param is_active query bool false "Filter by tenant service active state"
// @Param min_price query string false "Minimum base price"
// @Param max_price query string false "Maximum base price"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} ServiceListResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /services [get]
func (h *ServiceHandler) List(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	params := pagination.ParseParams(c)
	filters, err := parseServiceListFilters(c, params)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_service_filters", "invalid service filters")
		return
	}

	items, err := h.service.ListServicesByCompanyID(c.Request.Context(), companyID, filters)
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

	item, err := h.service.GetServiceDetailByID(c.Request.Context(), companyID, serviceID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_service_failed", "failed to get service")
		return
	}

	middleware.JSONData(c, http.StatusOK, mapServiceDetail(item))
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

	subServices, err := parseServiceSubServices(req.SubServices)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_sub_services", "invalid sub_services")
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
		SubServices:  subServices,
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "create_service_failed", "failed to create service")
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionCreate,
		EntityTable: "services",
		EntityID:    item.Item.ID,
		CompanyID:   companyID,
		OldData:     nil,
		NewData:     item,
	})

	middleware.JSONData(c, http.StatusCreated, mapServiceDetail(item))
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

	subServices, err := parseOptionalServiceSubServices(req.SubServices)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_sub_services", "invalid sub_services")
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
		SubServices:  subServices,
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "update_service_failed", "failed to update service")
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionUpdate,
		EntityTable: "services",
		EntityID:    item.Item.ID,
		CompanyID:   companyID,
		OldData:     before,
		NewData:     item,
	})

	middleware.JSONData(c, http.StatusOK, mapServiceDetail(item))
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

func parseServiceListFilters(c *gin.Context, params pagination.Params) (service.ServiceListFilters, error) {
	filters := service.ServiceListFilters{
		Params:   params,
		TypeName: strings.TrimSpace(c.Query("type_name")),
	}

	if rawIsActive := strings.TrimSpace(c.Query("is_active")); rawIsActive != "" {
		isActive, err := strconv.ParseBool(rawIsActive)
		if err != nil {
			return service.ServiceListFilters{}, err
		}
		filters.IsActive = &isActive
	}

	if rawMinPrice := strings.TrimSpace(c.Query("min_price")); rawMinPrice != "" {
		minPrice, err := parseRequiredNumeric(rawMinPrice)
		if err != nil {
			return service.ServiceListFilters{}, err
		}
		filters.MinPrice = &minPrice
	}

	if rawMaxPrice := strings.TrimSpace(c.Query("max_price")); rawMaxPrice != "" {
		maxPrice, err := parseRequiredNumeric(rawMaxPrice)
		if err != nil {
			return service.ServiceListFilters{}, err
		}
		filters.MaxPrice = &maxPrice
	}

	return filters, nil
}

func mapServiceList(items []sqlc.ListServicesByCompanyIDRow) []serviceResponse {
	mapped := make([]serviceResponse, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, serviceResponse{
			ID:                item.ID.String(),
			TypeID:            item.TypeID.String(),
			TypeName:          item.TypeName,
			Title:             item.Title,
			Description:       item.Description,
			Notes:             nullableText(item.Notes),
			Price:             numericToString(item.Price),
			DiscountRate:      numericToString(item.DiscountRate),
			ImageURL:          nullableText(item.ImageUrl),
			IsActive:          item.IsActive,
			SubServicesCount:  item.SubServicesCount,
			AverageTimesCount: item.AverageTimesCount,
		})
	}
	return mapped
}

func mapServiceItem(item sqlc.GetServiceByIDAndCompanyIDRow) serviceResponse {
	return serviceResponse{
		ID:                item.ID.String(),
		TypeID:            item.TypeID.String(),
		TypeName:          item.TypeName,
		Title:             item.Title,
		Description:       item.Description,
		Notes:             nullableText(item.Notes),
		Price:             numericToString(item.Price),
		DiscountRate:      numericToString(item.DiscountRate),
		ImageURL:          nullableText(item.ImageUrl),
		IsActive:          item.IsActive,
		SubServicesCount:  item.SubServicesCount,
		AverageTimesCount: item.AverageTimesCount,
	}
}

func mapServiceDetail(detail service.ServiceDetail) serviceResponse {
	response := mapServiceItem(detail.Item)
	response.SubServices = make([]serviceSubServiceResponse, 0, len(detail.SubServices))
	for _, subService := range detail.SubServices {
		item := subService.Item
		mapped := serviceSubServiceResponse{
			ID:           item.ID.String(),
			TypeID:       item.TypeID.String(),
			Title:        item.Title,
			Description:  item.Description,
			Notes:        nullableText(item.Notes),
			Price:        numericToString(item.Price),
			DiscountRate: numericToString(item.DiscountRate),
			ImageURL:     nullableText(item.ImageUrl),
			IsActive:     item.IsActive,
			AverageTimes: make([]serviceAverageTimeResponse, 0, len(subService.AverageTimes)),
		}
		for _, averageTime := range subService.AverageTimes {
			mapped.AverageTimes = append(mapped.AverageTimes, serviceAverageTimeResponse{
				ID:                 averageTime.ID.String(),
				PetSize:            string(averageTime.PetSize),
				PetKind:            string(averageTime.PetKind),
				PetTemperament:     string(averageTime.PetTemperament),
				AverageTimeMinutes: averageTime.AverageTimeMinutes,
			})
		}
		response.SubServices = append(response.SubServices, mapped)
	}
	return response
}

func parseOptionalServiceSubServices(raw *[]serviceSubServiceRequest) (*[]service.ServiceSubServiceInput, error) {
	if raw == nil {
		return nil, nil
	}
	parsed, err := parseServiceSubServices(*raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseServiceSubServices(raw []serviceSubServiceRequest) ([]service.ServiceSubServiceInput, error) {
	items := make([]service.ServiceSubServiceInput, 0, len(raw))
	for _, subService := range raw {
		price, err := parseRequiredNumeric(subService.Price)
		if err != nil {
			return nil, err
		}
		discountRate, err := parseOptionalNumeric(subService.DiscountRate)
		if err != nil {
			return nil, err
		}

		averageTimes, err := parseServiceAverageTimes(subService.AverageTimes)
		if err != nil {
			return nil, err
		}

		items = append(items, service.ServiceSubServiceInput{
			TypeName:     strings.TrimSpace(subService.TypeName),
			Title:        strings.TrimSpace(subService.Title),
			Description:  strings.TrimSpace(subService.Description),
			Notes:        textValue(strings.TrimSpace(subService.Notes)),
			Price:        price,
			DiscountRate: discountRate,
			ImageURL:     textValue(strings.TrimSpace(subService.ImageURL)),
			IsActive:     subService.IsActive == nil || *subService.IsActive,
			AverageTimes: averageTimes,
		})
	}
	return items, nil
}

func parseServiceAverageTimes(raw []serviceAverageTimeRequest) ([]service.ServiceAverageTimeInput, error) {
	items := make([]service.ServiceAverageTimeInput, 0, len(raw))
	for _, averageTime := range raw {
		if averageTime.AverageTimeMinutes <= 0 || averageTime.AverageTimeMinutes > 32767 {
			return nil, strconv.ErrSyntax
		}
		petSize, err := parsePetSize(averageTime.PetSize)
		if err != nil {
			return nil, err
		}
		petKind, err := parsePetKind(averageTime.PetKind)
		if err != nil {
			return nil, err
		}
		petTemperament, err := parsePetTemperament(averageTime.PetTemperament)
		if err != nil {
			return nil, err
		}
		items = append(items, service.ServiceAverageTimeInput{
			PetSize:            petSize,
			PetKind:            petKind,
			PetTemperament:     petTemperament,
			AverageTimeMinutes: int16(averageTime.AverageTimeMinutes),
		})
	}
	return items, nil
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

func hasServiceUpdatePayload(req updateServiceRequest) bool {
	return req.TypeName != nil ||
		req.Title != nil ||
		req.Description != nil ||
		req.Notes != nil ||
		req.Price != nil ||
		req.DiscountRate != nil ||
		req.ImageURL != nil ||
		req.IsActive != nil ||
		req.SubServices != nil
}
