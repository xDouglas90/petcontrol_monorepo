package handler

import (
	"net/http"

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list plans"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": plans})
}

func (h *PlanHandler) Current(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "company context required"})
		return
	}

	plan, err := h.service.GetCurrentPlan(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(apperror.HTTPStatus(err), gin.H{"error": "failed to get current plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": plan})
}
