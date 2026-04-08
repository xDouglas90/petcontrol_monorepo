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
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type ScheduleHandler struct {
	service *service.ScheduleService
}

type createScheduleRequest struct {
	ClientID     string  `json:"client_id"`
	PetID        string  `json:"pet_id"`
	ScheduledAt  string  `json:"scheduled_at"`
	EstimatedEnd *string `json:"estimated_end"`
	Notes        string  `json:"notes"`
	Status       string  `json:"status"`
	StatusNotes  string  `json:"status_notes"`
}

type updateScheduleRequest struct {
	ClientID     *string `json:"client_id"`
	PetID        *string `json:"pet_id"`
	ScheduledAt  *string `json:"scheduled_at"`
	EstimatedEnd *string `json:"estimated_end"`
	Notes        *string `json:"notes"`
	Status       *string `json:"status"`
	StatusNotes  *string `json:"status_notes"`
}

func NewScheduleHandler(service *service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{service: service}
}

func (h *ScheduleHandler) List(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "company context required"})
		return
	}

	items, err := h.service.ListSchedulesByCompanyID(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list schedules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *ScheduleHandler) GetByID(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "company context required"})
		return
	}

	scheduleID, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid schedule id"})
		return
	}

	item, err := h.service.GetScheduleByID(c.Request.Context(), companyID, scheduleID)
	if err != nil {
		c.JSON(apperror.HTTPStatus(err), gin.H{"error": "failed to get schedule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *ScheduleHandler) Create(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "company context required"})
		return
	}

	claims, ok := middleware.GetClaims(c)
	if !ok || claims.UserID == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "user context required"})
		return
	}

	createdBy, err := parseUUID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid user_id in token"})
		return
	}

	var req createScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid request body"})
		return
	}

	clientID, err := parseUUID(req.ClientID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid client_id"})
		return
	}

	petID, err := parseUUID(req.PetID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid pet_id"})
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid scheduled_at"})
		return
	}

	estimatedEnd, err := parseOptionalTime(req.EstimatedEnd)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid estimated_end"})
		return
	}

	status := sqlc.ScheduleStatus(strings.TrimSpace(req.Status))
	item, err := h.service.CreateSchedule(c.Request.Context(), service.CreateScheduleInput{
		CompanyID:    companyID,
		ClientID:     clientID,
		PetID:        petID,
		ScheduledAt:  scheduledAt,
		EstimatedEnd: estimatedEnd,
		Notes:        strings.TrimSpace(req.Notes),
		CreatedBy:    createdBy,
		Status:       status,
		StatusNotes:  strings.TrimSpace(req.StatusNotes),
	})
	if err != nil {
		c.JSON(apperror.HTTPStatus(err), gin.H{"error": "failed to create schedule"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *ScheduleHandler) Update(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "company context required"})
		return
	}

	claims, ok := middleware.GetClaims(c)
	if !ok || claims.UserID == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "user context required"})
		return
	}

	changedBy, err := parseUUID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid user_id in token"})
		return
	}

	scheduleID, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid schedule id"})
		return
	}

	var req updateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid request body"})
		return
	}

	clientID, err := parseOptionalUUID(req.ClientID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid client_id"})
		return
	}

	petID, err := parseOptionalUUID(req.PetID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid pet_id"})
		return
	}

	scheduledAt, err := parseOptionalRFC3339(req.ScheduledAt)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid scheduled_at"})
		return
	}

	estimatedEnd, err := parseOptionalRFC3339(req.EstimatedEnd)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid estimated_end"})
		return
	}

	item, err := h.service.UpdateSchedule(c.Request.Context(), service.UpdateScheduleInput{
		CompanyID:    companyID,
		ScheduleID:   scheduleID,
		ClientID:     clientID,
		PetID:        petID,
		ScheduledAt:  scheduledAt,
		EstimatedEnd: estimatedEnd,
		Notes:        parseOptionalTrimmed(req.Notes),
		Status:       parseOptionalScheduleStatus(req.Status),
		StatusNotes:  parseOptionalTrimmed(req.StatusNotes),
		ChangedBy:    changedBy,
	})
	if err != nil {
		c.JSON(apperror.HTTPStatus(err), gin.H{"error": "failed to update schedule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *ScheduleHandler) Delete(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "company context required"})
		return
	}

	scheduleID, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid schedule id"})
		return
	}

	if err := h.service.DeleteSchedule(c.Request.Context(), companyID, scheduleID); err != nil {
		c.JSON(apperror.HTTPStatus(err), gin.H{"error": "failed to delete schedule"})
		return
	}

	c.Status(http.StatusNoContent)
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
