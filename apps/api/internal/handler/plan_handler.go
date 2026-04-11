package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type PlanHandler struct {
	service *service.PlanService
}

func NewPlanHandler(service *service.PlanService) *PlanHandler {
	return &PlanHandler{service: service}
}

func (h *PlanHandler) List(c *gin.Context) {
	plans, err := h.service.ListPlans(c.Request.Context())
	if err != nil {
		middleware.JSONError(c, 500, "list_plans_failed", "failed to list plans")
		return
	}

	middleware.JSONData(c, 200, plans)
}

// Current godoc
// @Summary Get current plan
// @Description Returns the active subscription plan for the authenticated tenant.
// @Tags plans
// @Security BearerAuth
// @Produce json
// @Success 200 {object} PlanItemResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /plans/current [get]
func (h *PlanHandler) Current(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, 403, "company_context_required", "company context required")
		return
	}

	plan, err := h.service.GetCurrentPlan(c.Request.Context(), companyID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_current_plan_failed", "failed to get current plan")
		return
	}

	middleware.JSONData(c, 200, plan)
}
