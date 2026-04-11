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
	"github.com/xdouglas90/petcontrol_monorepo/internal/queue"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type ScheduleHandler struct {
	service   *service.ScheduleService
	publisher queue.Publisher
}

type createScheduleRequest struct {
	ClientID     string   `json:"client_id"`
	PetID        string   `json:"pet_id"`
	ServiceIDs   []string `json:"service_ids"`
	ScheduledAt  string   `json:"scheduled_at"`
	EstimatedEnd *string  `json:"estimated_end"`
	Notes        string   `json:"notes"`
	Status       string   `json:"status"`
	StatusNotes  string   `json:"status_notes"`
}

type updateScheduleRequest struct {
	ClientID     *string   `json:"client_id"`
	PetID        *string   `json:"pet_id"`
	ServiceIDs   *[]string `json:"service_ids"`
	ScheduledAt  *string   `json:"scheduled_at"`
	EstimatedEnd *string   `json:"estimated_end"`
	Notes        *string   `json:"notes"`
	Status       *string   `json:"status"`
	StatusNotes  *string   `json:"status_notes"`
}

type scheduleResponse struct {
	ID            string     `json:"id"`
	CompanyID     string     `json:"company_id"`
	ClientID      string     `json:"client_id"`
	PetID         string     `json:"pet_id"`
	ClientName    string     `json:"client_name"`
	PetName       string     `json:"pet_name"`
	ServiceIDs    []string   `json:"service_ids"`
	ServiceTitles []string   `json:"service_titles"`
	ScheduledAt   time.Time  `json:"scheduled_at"`
	EstimatedEnd  *time.Time `json:"estimated_end,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	CurrentStatus string     `json:"current_status"`
}

func NewScheduleHandler(service *service.ScheduleService, publisher ...queue.Publisher) *ScheduleHandler {
	handler := &ScheduleHandler{service: service}
	if len(publisher) > 0 {
		handler.publisher = publisher[0]
	}
	return handler
}

// List godoc
// @Summary List schedules
// @Description Returns schedules from the authenticated tenant.
// @Tags schedules
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ScheduleListResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /schedules [get]
func (h *ScheduleHandler) List(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, 403, "company_context_required", "company context required")
		return
	}

	params := pagination.ParseParams(c)
	items, err := h.service.ListSchedulesByCompanyID(c.Request.Context(), companyID, params)
	if err != nil {
		middleware.JSONError(c, 500, "list_schedules_failed", "failed to list schedules")
		return
	}

	total := 0
	if len(items) > 0 {
		total = int(items[0].TotalCount)
	}

	middleware.JSONPaginated(c, 200, mapScheduleList(items), pagination.NewMeta(total, params.Page, params.Limit))
}

// GetByID godoc
// @Summary Get schedule by ID
// @Description Returns a single schedule from the authenticated tenant.
// @Tags schedules
// @Security BearerAuth
// @Produce json
// @Param id path string true "Schedule ID"
// @Success 200 {object} ScheduleItemResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /schedules/{id} [get]
func (h *ScheduleHandler) GetByID(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, 403, "company_context_required", "company context required")
		return
	}

	scheduleID, err := parseUUID(c.Param("id"))
	if err != nil {
		middleware.JSONError(c, 422, "invalid_schedule_id", "invalid schedule id")
		return
	}

	item, err := h.service.GetScheduleByID(c.Request.Context(), companyID, scheduleID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_schedule_failed", "failed to get schedule")
		return
	}

	middleware.JSONData(c, 200, mapScheduleItem(item))
}

// History godoc
// @Summary Get schedule history
// @Description Returns status history for a schedule in the authenticated tenant.
// @Tags schedules
// @Security BearerAuth
// @Produce json
// @Param id path string true "Schedule ID"
// @Success 200 {object} ScheduleHistoryResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /schedules/{id}/history [get]
func (h *ScheduleHandler) History(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, 403, "company_context_required", "company context required")
		return
	}

	scheduleID, err := parseUUID(c.Param("id"))
	if err != nil {
		middleware.JSONError(c, 422, "invalid_schedule_id", "invalid schedule id")
		return
	}

	items, err := h.service.ListScheduleStatusHistory(c.Request.Context(), companyID, scheduleID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "list_schedule_history_failed", "failed to list schedule history")
		return
	}

	middleware.JSONData(c, 200, items)
}

// Create godoc
// @Summary Create schedule
// @Description Creates a schedule for the authenticated tenant.
// @Tags schedules
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ScheduleCreateRequestDoc true "Schedule payload"
// @Success 201 {object} ScheduleItemResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /schedules [post]
func (h *ScheduleHandler) Create(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, 403, "company_context_required", "company context required")
		return
	}

	claims, ok := middleware.GetClaims(c)
	if !ok || claims.UserID == "" {
		middleware.JSONError(c, 403, "user_context_required", "user context required")
		return
	}

	createdBy, err := parseUUID(claims.UserID)
	if err != nil {
		middleware.JSONError(c, 403, "invalid_user_id", "invalid user_id in token")
		return
	}

	var req createScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONError(c, 422, "invalid_request_body", "invalid request body")
		return
	}

	clientID, err := parseUUID(req.ClientID)
	if err != nil {
		middleware.JSONError(c, 422, "invalid_client_id", "invalid client_id")
		return
	}

	petID, err := parseUUID(req.PetID)
	if err != nil {
		middleware.JSONError(c, 422, "invalid_pet_id", "invalid pet_id")
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		middleware.JSONError(c, 422, "invalid_scheduled_at", "invalid scheduled_at")
		return
	}

	estimatedEnd, err := parseOptionalTime(req.EstimatedEnd)
	if err != nil {
		middleware.JSONError(c, 422, "invalid_estimated_end", "invalid estimated_end")
		return
	}

	serviceIDs, err := parseUUIDList(req.ServiceIDs)
	if err != nil {
		middleware.JSONError(c, 422, "invalid_service_ids", "invalid service_ids")
		return
	}

	status := sqlc.ScheduleStatus(strings.TrimSpace(req.Status))
	item, err := h.service.CreateSchedule(c.Request.Context(), service.CreateScheduleInput{
		CompanyID:    companyID,
		ClientID:     clientID,
		PetID:        petID,
		ServiceIDs:   serviceIDs,
		ScheduledAt:  scheduledAt,
		EstimatedEnd: estimatedEnd,
		Notes:        strings.TrimSpace(req.Notes),
		CreatedBy:    createdBy,
		Status:       status,
		StatusNotes:  strings.TrimSpace(req.StatusNotes),
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "create_schedule_failed", "failed to create schedule")
		return
	}

	if err := h.publishScheduleConfirmationIfNeeded(c, nil, item, createdBy, strings.TrimSpace(req.StatusNotes)); err != nil {
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionCreate,
		EntityTable: "schedules",
		EntityID:    item.ID,
		CompanyID:   companyID,
		OldData:     nil,
		NewData:     item,
	})

	middleware.JSONData(c, 201, mapScheduleItem(item))
}

// Update godoc
// @Summary Update schedule
// @Description Updates a schedule for the authenticated tenant.
// @Tags schedules
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Schedule ID"
// @Param request body ScheduleUpdateRequestDoc true "Schedule payload"
// @Success 200 {object} ScheduleItemResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /schedules/{id} [put]
func (h *ScheduleHandler) Update(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, 403, "company_context_required", "company context required")
		return
	}

	claims, ok := middleware.GetClaims(c)
	if !ok || claims.UserID == "" {
		middleware.JSONError(c, 403, "user_context_required", "user context required")
		return
	}

	changedBy, err := parseUUID(claims.UserID)
	if err != nil {
		middleware.JSONError(c, 403, "invalid_user_id", "invalid user_id in token")
		return
	}

	scheduleID, err := parseUUID(c.Param("id"))
	if err != nil {
		middleware.JSONError(c, 422, "invalid_schedule_id", "invalid schedule id")
		return
	}

	before, err := h.service.GetScheduleByID(c.Request.Context(), companyID, scheduleID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_schedule_failed", "failed to get schedule")
		return
	}

	var req updateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONError(c, 422, "invalid_request_body", "invalid request body")
		return
	}

	clientID, err := parseOptionalUUID(req.ClientID)
	if err != nil {
		middleware.JSONError(c, 422, "invalid_client_id", "invalid client_id")
		return
	}

	petID, err := parseOptionalUUID(req.PetID)
	if err != nil {
		middleware.JSONError(c, 422, "invalid_pet_id", "invalid pet_id")
		return
	}

	scheduledAt, err := parseOptionalRFC3339(req.ScheduledAt)
	if err != nil {
		middleware.JSONError(c, 422, "invalid_scheduled_at", "invalid scheduled_at")
		return
	}

	estimatedEnd, err := parseOptionalRFC3339(req.EstimatedEnd)
	if err != nil {
		middleware.JSONError(c, 422, "invalid_estimated_end", "invalid estimated_end")
		return
	}

	serviceIDs, err := parseOptionalUUIDList(req.ServiceIDs)
	if err != nil {
		middleware.JSONError(c, 422, "invalid_service_ids", "invalid service_ids")
		return
	}

	item, err := h.service.UpdateSchedule(c.Request.Context(), service.UpdateScheduleInput{
		CompanyID:    companyID,
		ScheduleID:   scheduleID,
		ClientID:     clientID,
		PetID:        petID,
		ServiceIDs:   serviceIDs,
		ScheduledAt:  scheduledAt,
		EstimatedEnd: estimatedEnd,
		Notes:        parseOptionalTrimmed(req.Notes),
		Status:       parseOptionalScheduleStatus(req.Status),
		StatusNotes:  parseOptionalTrimmed(req.StatusNotes),
		ChangedBy:    changedBy,
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "update_schedule_failed", "failed to update schedule")
		return
	}

	if err := h.publishScheduleConfirmationIfNeeded(c, &before, item, changedBy, parseOptionalString(req.StatusNotes)); err != nil {
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionUpdate,
		EntityTable: "schedules",
		EntityID:    item.ID,
		CompanyID:   companyID,
		OldData:     before,
		NewData:     item,
	})

	middleware.JSONData(c, 200, mapScheduleItem(item))
}

// Delete godoc
// @Summary Delete schedule
// @Description Soft deletes a schedule from the authenticated tenant.
// @Tags schedules
// @Security BearerAuth
// @Param id path string true "Schedule ID"
// @Success 204
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /schedules/{id} [delete]
func (h *ScheduleHandler) Delete(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, 403, "company_context_required", "company context required")
		return
	}

	scheduleID, err := parseUUID(c.Param("id"))
	if err != nil {
		middleware.JSONError(c, 422, "invalid_schedule_id", "invalid schedule id")
		return
	}

	before, err := h.service.GetScheduleByID(c.Request.Context(), companyID, scheduleID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_schedule_failed", "failed to get schedule")
		return
	}

	if err := h.service.DeleteSchedule(c.Request.Context(), companyID, scheduleID); err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "delete_schedule_failed", "failed to delete schedule")
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionDelete,
		EntityTable: "schedules",
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

func (h *ScheduleHandler) publishScheduleConfirmationIfNeeded(c *gin.Context, before *sqlc.GetScheduleByIDAndCompanyIDRow, after sqlc.GetScheduleByIDAndCompanyIDRow, changedBy pgtype.UUID, statusNotes string) error {
	if h.publisher == nil {
		return nil
	}
	if after.CurrentStatus != sqlc.ScheduleStatusConfirmed {
		return nil
	}
	if before != nil && before.CurrentStatus == sqlc.ScheduleStatusConfirmed {
		return nil
	}

	if err := h.publisher.EnqueueScheduleConfirmation(c.Request.Context(), queue.ScheduleConfirmationPayload{
		Version:       2,
		ScheduleID:    after.ID.String(),
		CompanyID:     after.CompanyID.String(),
		ChangedBy:     changedBy.String(),
		ClientName:    after.ClientName,
		PetName:       after.PetName,
		ServiceTitles: stringSliceFromAny(after.ServiceTitles),
		Status:        string(after.CurrentStatus),
		StatusNotes:   statusNotes,
		OccurredAt:    time.Now().UTC(),
	}); err != nil {
		middleware.JSONError(c, 500, "enqueue_schedule_confirmation_failed", "failed to enqueue schedule confirmation")
		return err
	}

	return nil
}

func parseOptionalString(raw *string) string {
	if raw == nil {
		return ""
	}
	return strings.TrimSpace(*raw)
}

func parseUUIDList(raw []string) ([]pgtype.UUID, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	items := make([]pgtype.UUID, 0, len(raw))
	for _, item := range raw {
		parsed, err := parseUUID(strings.TrimSpace(item))
		if err != nil {
			return nil, err
		}
		items = append(items, parsed)
	}

	return items, nil
}

func parseOptionalUUIDList(raw *[]string) (*[]pgtype.UUID, error) {
	if raw == nil {
		return nil, nil
	}

	items, err := parseUUIDList(*raw)
	if err != nil {
		return nil, err
	}

	return &items, nil
}

func parseOptionalTime(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, *raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalUUID(raw *string) (*pgtype.UUID, error) {
	if raw == nil {
		return nil, nil
	}
	value, err := parseUUID(strings.TrimSpace(*raw))
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func parseOptionalRFC3339(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	value, err := time.Parse(time.RFC3339, strings.TrimSpace(*raw))
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func parseOptionalTrimmed(raw *string) *string {
	if raw == nil {
		return nil
	}
	value := strings.TrimSpace(*raw)
	return &value
}

func parseOptionalScheduleStatus(raw *string) *sqlc.ScheduleStatus {
	if raw == nil {
		return nil
	}
	value := sqlc.ScheduleStatus(strings.TrimSpace(*raw))
	return &value
}

func mapScheduleList(items []sqlc.ListSchedulesByCompanyIDRow) []scheduleResponse {
	mapped := make([]scheduleResponse, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, scheduleResponse{
			ID:            item.ID.String(),
			CompanyID:     item.CompanyID.String(),
			ClientID:      item.ClientID.String(),
			PetID:         item.PetID.String(),
			ClientName:    item.ClientName,
			PetName:       item.PetName,
			ServiceIDs:    stringSliceFromAny(item.ServiceIds),
			ServiceTitles: stringSliceFromAny(item.ServiceTitles),
			ScheduledAt:   item.ScheduledAt.Time,
			EstimatedEnd:  nullableTime(item.EstimatedEnd),
			Notes:         nullableText(item.Notes),
			CurrentStatus: string(item.CurrentStatus),
		})
	}
	return mapped
}

func mapScheduleItem(item sqlc.GetScheduleByIDAndCompanyIDRow) scheduleResponse {
	return scheduleResponse{
		ID:            item.ID.String(),
		CompanyID:     item.CompanyID.String(),
		ClientID:      item.ClientID.String(),
		PetID:         item.PetID.String(),
		ClientName:    item.ClientName,
		PetName:       item.PetName,
		ServiceIDs:    stringSliceFromAny(item.ServiceIds),
		ServiceTitles: stringSliceFromAny(item.ServiceTitles),
		ScheduledAt:   item.ScheduledAt.Time,
		EstimatedEnd:  nullableTime(item.EstimatedEnd),
		Notes:         nullableText(item.Notes),
		CurrentStatus: string(item.CurrentStatus),
	}
}

func nullableTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	timestamp := value.Time
	return &timestamp
}

func stringSliceFromAny(raw interface{}) []string {
	switch value := raw.(type) {
	case nil:
		return nil
	case []string:
		return append([]string(nil), value...)
	case []any:
		result := make([]string, 0, len(value))
		for _, item := range value {
			switch typed := item.(type) {
			case string:
				result = append(result, typed)
			case []byte:
				result = append(result, string(typed))
			}
		}
		return result
	default:
		return nil
	}
}
