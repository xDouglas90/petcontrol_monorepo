package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type ModuleHandler struct {
	service *service.ModuleService
}

func NewModuleHandler(service *service.ModuleService) *ModuleHandler {
	return &ModuleHandler{service: service}
}

func (h *ModuleHandler) List(c *gin.Context) {
	modules, err := h.service.ListModules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list modules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": modules})
}

func (h *ModuleHandler) Active(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "company context required"})
		return
	}

	modules, err := h.service.ListActiveModulesByCompanyID(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(apperror.HTTPStatus(err), gin.H{"error": "failed to list active modules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": modules})
}
