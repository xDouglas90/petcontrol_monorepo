package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type CompanySystemConfigHandler struct {
	service *service.CompanySystemConfigService
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
