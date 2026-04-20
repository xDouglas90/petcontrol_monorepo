package handler

import (
	"fmt"
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

type CompanySystemConfigHandler struct {
	service *service.CompanySystemConfigService
}

type updateCompanySystemConfigRequest struct {
	ScheduleInitTime      string   `json:"schedule_init_time"`
	SchedulePauseInitTime *string  `json:"schedule_pause_init_time"`
	SchedulePauseEndTime  *string  `json:"schedule_pause_end_time"`
	ScheduleEndTime       string   `json:"schedule_end_time"`
	MinSchedulesPerDay    int16    `json:"min_schedules_per_day"`
	MaxSchedulesPerDay    int16    `json:"max_schedules_per_day"`
	ScheduleDays          []string `json:"schedule_days"`
	DynamicCages          bool     `json:"dynamic_cages"`
	TotalSmallCages       int16    `json:"total_small_cages"`
	TotalMediumCages      int16    `json:"total_medium_cages"`
	TotalLargeCages       int16    `json:"total_large_cages"`
	TotalGiantCages       int16    `json:"total_giant_cages"`
	WhatsappNotifications bool     `json:"whatsapp_notifications"`
	WhatsappConversation  bool     `json:"whatsapp_conversation"`
	WhatsappBusinessPhone *string  `json:"whatsapp_business_phone"`
}

func NewCompanySystemConfigHandler(service *service.CompanySystemConfigService) *CompanySystemConfigHandler {
	return &CompanySystemConfigHandler{service: service}
}

// Current godoc
// @Summary Get current company system config
// @Description Returns the system configuration resolved from the authenticated tenant context.
// @Tags company-system-configs
// @Security BearerAuth
// @Produce json
// @Success 200 {object} CompanySystemConfigResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /company-system-configs/current [get]
func (h *CompanySystemConfigHandler) Current(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	item, err := h.service.GetCurrent(c.Request.Context(), companyID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_company_system_config_failed", "failed to get company system config")
		return
	}

	middleware.JSONData(c, http.StatusOK, gin.H{
		"company_id":               uuidToString(item.CompanyID),
		"schedule_init_time":       formatTime(item.ScheduleInitTime),
		"schedule_pause_init_time": formatTime(item.SchedulePauseInitTime),
		"schedule_pause_end_time":  formatTime(item.SchedulePauseEndTime),
		"schedule_end_time":        formatTime(item.ScheduleEndTime),
		"min_schedules_per_day":    item.MinSchedulesPerDay,
		"max_schedules_per_day":    item.MaxSchedulesPerDay,
		"schedule_days":            weekDaysToStrings(item.ScheduleDays),
		"dynamic_cages":            item.DynamicCages,
		"total_small_cages":        item.TotalSmallCages,
		"total_medium_cages":       item.TotalMediumCages,
		"total_large_cages":        item.TotalLargeCages,
		"total_giant_cages":        item.TotalGiantCages,
		"whatsapp_notifications":   item.WhatsappNotifications,
		"whatsapp_conversation":    item.WhatsappConversation,
		"whatsapp_business_phone":  nullableText(item.WhatsappBusinessPhone),
	})
}

// Update godoc
// @Summary Update current company system config
// @Description Updates the authenticated tenant system configuration.
// @Tags company-system-configs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CompanySystemConfigUpdateRequestDoc true "Update payload"
// @Success 200 {object} CompanySystemConfigResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /company-system-configs/current [patch]
func (h *CompanySystemConfigHandler) Update(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	var req updateCompanySystemConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_request_body", "invalid request body")
		return
	}

	params, err := buildUpdateCompanySystemConfigParams(companyID, req)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_company_system_config", err.Error())
		return
	}

	item, err := h.service.UpdateCurrent(c.Request.Context(), params)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "update_company_system_config_failed", "failed to update company system config")
		return
	}

	middleware.JSONData(c, http.StatusOK, gin.H{
		"company_id":               uuidToString(item.CompanyID),
		"schedule_init_time":       formatTime(item.ScheduleInitTime),
		"schedule_pause_init_time": formatTime(item.SchedulePauseInitTime),
		"schedule_pause_end_time":  formatTime(item.SchedulePauseEndTime),
		"schedule_end_time":        formatTime(item.ScheduleEndTime),
		"min_schedules_per_day":    item.MinSchedulesPerDay,
		"max_schedules_per_day":    item.MaxSchedulesPerDay,
		"schedule_days":            weekDaysToStrings(item.ScheduleDays),
		"dynamic_cages":            item.DynamicCages,
		"total_small_cages":        item.TotalSmallCages,
		"total_medium_cages":       item.TotalMediumCages,
		"total_large_cages":        item.TotalLargeCages,
		"total_giant_cages":        item.TotalGiantCages,
		"whatsapp_notifications":   item.WhatsappNotifications,
		"whatsapp_conversation":    item.WhatsappConversation,
		"whatsapp_business_phone":  nullableText(item.WhatsappBusinessPhone),
	})
}

func buildUpdateCompanySystemConfigParams(companyID pgtype.UUID, req updateCompanySystemConfigRequest) (sqlc.UpdateCompanySystemConfigParams, error) {
	scheduleInitTime, err := parseRequiredClock(req.ScheduleInitTime)
	if err != nil {
		return sqlc.UpdateCompanySystemConfigParams{}, err
	}
	scheduleEndTime, err := parseRequiredClock(req.ScheduleEndTime)
	if err != nil {
		return sqlc.UpdateCompanySystemConfigParams{}, err
	}
	if scheduleInitTime.Microseconds >= scheduleEndTime.Microseconds {
		return sqlc.UpdateCompanySystemConfigParams{}, fmt.Errorf("%w: schedule_init_time must be before schedule_end_time", apperror.ErrUnprocessableEntity)
	}

	schedulePauseInitTime, err := parseNullableClock(req.SchedulePauseInitTime)
	if err != nil {
		return sqlc.UpdateCompanySystemConfigParams{}, err
	}
	schedulePauseEndTime, err := parseNullableClock(req.SchedulePauseEndTime)
	if err != nil {
		return sqlc.UpdateCompanySystemConfigParams{}, err
	}

	if schedulePauseInitTime.Valid != schedulePauseEndTime.Valid {
		return sqlc.UpdateCompanySystemConfigParams{}, fmt.Errorf("%w: schedule_pause_init_time and schedule_pause_end_time must be provided together", apperror.ErrUnprocessableEntity)
	}
	if schedulePauseInitTime.Valid && schedulePauseEndTime.Valid {
		if schedulePauseInitTime.Microseconds >= schedulePauseEndTime.Microseconds {
			return sqlc.UpdateCompanySystemConfigParams{}, fmt.Errorf("%w: schedule_pause_init_time must be before schedule_pause_end_time", apperror.ErrUnprocessableEntity)
		}
		if schedulePauseInitTime.Microseconds <= scheduleInitTime.Microseconds || schedulePauseEndTime.Microseconds >= scheduleEndTime.Microseconds {
			return sqlc.UpdateCompanySystemConfigParams{}, fmt.Errorf("%w: pause window must be inside operational hours", apperror.ErrUnprocessableEntity)
		}
	}

	scheduleDays, err := parseRequiredWeekDays(req.ScheduleDays)
	if err != nil {
		return sqlc.UpdateCompanySystemConfigParams{}, err
	}

	if req.MinSchedulesPerDay < 0 || req.MaxSchedulesPerDay < 0 {
		return sqlc.UpdateCompanySystemConfigParams{}, fmt.Errorf("%w: schedule limits must be zero or positive", apperror.ErrUnprocessableEntity)
	}
	if req.MinSchedulesPerDay > req.MaxSchedulesPerDay {
		return sqlc.UpdateCompanySystemConfigParams{}, fmt.Errorf("%w: min_schedules_per_day must be less than or equal to max_schedules_per_day", apperror.ErrUnprocessableEntity)
	}

	if req.TotalSmallCages < 0 || req.TotalMediumCages < 0 || req.TotalLargeCages < 0 || req.TotalGiantCages < 0 {
		return sqlc.UpdateCompanySystemConfigParams{}, fmt.Errorf("%w: cage totals must be zero or positive", apperror.ErrUnprocessableEntity)
	}

	return sqlc.UpdateCompanySystemConfigParams{
		CompanyID:             companyID,
		ScheduleInitTime:      scheduleInitTime,
		SchedulePauseInitTime: schedulePauseInitTime,
		SchedulePauseEndTime:  schedulePauseEndTime,
		ScheduleEndTime:       scheduleEndTime,
		MinSchedulesPerDay:    req.MinSchedulesPerDay,
		MaxSchedulesPerDay:    req.MaxSchedulesPerDay,
		ScheduleDays:          scheduleDays,
		DynamicCages:          req.DynamicCages,
		TotalSmallCages:       req.TotalSmallCages,
		TotalMediumCages:      req.TotalMediumCages,
		TotalLargeCages:       req.TotalLargeCages,
		TotalGiantCages:       req.TotalGiantCages,
		WhatsappNotifications: req.WhatsappNotifications,
		WhatsappConversation:  req.WhatsappConversation,
		WhatsappBusinessPhone: textPointer(parseOptionalTrimmed(req.WhatsappBusinessPhone)),
	}, nil
}

func parseRequiredClock(raw string) (pgtype.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return pgtype.Time{}, fmt.Errorf("%w: time fields are required", apperror.ErrUnprocessableEntity)
	}

	parsed, err := time.Parse("15:04", trimmed)
	if err != nil {
		return pgtype.Time{}, fmt.Errorf("%w: time fields must use HH:MM format", apperror.ErrUnprocessableEntity)
	}

	return pgtype.Time{
		Microseconds: int64(parsed.Hour())*int64(time.Hour/time.Microsecond) + int64(parsed.Minute())*int64(time.Minute/time.Microsecond),
		Valid:        true,
	}, nil
}

func parseNullableClock(raw *string) (pgtype.Time, error) {
	if raw == nil {
		return pgtype.Time{}, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return pgtype.Time{}, nil
	}
	return parseRequiredClock(trimmed)
}

func parseRequiredWeekDays(values []string) ([]sqlc.WeekDay, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("%w: schedule_days must contain at least one day", apperror.ErrUnprocessableEntity)
	}

	allowed := map[string]sqlc.WeekDay{
		string(sqlc.WeekDaySunday):    sqlc.WeekDaySunday,
		string(sqlc.WeekDayMonday):    sqlc.WeekDayMonday,
		string(sqlc.WeekDayTuesday):   sqlc.WeekDayTuesday,
		string(sqlc.WeekDayWednesday): sqlc.WeekDayWednesday,
		string(sqlc.WeekDayThursday):  sqlc.WeekDayThursday,
		string(sqlc.WeekDayFriday):    sqlc.WeekDayFriday,
		string(sqlc.WeekDaySaturday):  sqlc.WeekDaySaturday,
	}

	result := make([]sqlc.WeekDay, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := strings.ToLower(strings.TrimSpace(value))
		day, ok := allowed[normalized]
		if !ok {
			return nil, fmt.Errorf("%w: schedule_days contains an invalid week day", apperror.ErrUnprocessableEntity)
		}
		if _, duplicated := seen[normalized]; duplicated {
			return nil, fmt.Errorf("%w: schedule_days cannot contain duplicates", apperror.ErrUnprocessableEntity)
		}
		seen[normalized] = struct{}{}
		result = append(result, day)
	}

	return result, nil
}
