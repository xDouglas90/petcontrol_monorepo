package handler

import (
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
		middleware.JSONError(c, 500, "list_modules_failed", "failed to list modules")
		return
	}

	middleware.JSONData(c, 200, modules)
}

// Active godoc
// @Summary List active modules
// @Description Lists the active modules available to the authenticated tenant.
// @Tags modules
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ModuleListResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /modules/active [get]
func (h *ModuleHandler) Active(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, 403, "company_context_required", "company context required")
		return
	}

	modules, err := h.service.ListActiveModulesByCompanyID(c.Request.Context(), companyID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "list_active_modules_failed", "failed to list active modules")
		return
	}

	middleware.JSONData(c, 200, modules)
}

// Access godoc
// @Summary Check module access
// @Description Checks whether the authenticated tenant has access to the requested module code.
// @Tags modules
// @Security BearerAuth
// @Produce json
// @Param code path string true "Module code"
// @Success 200 {object} ModuleAccessResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Router /modules/{code}/access [get]
func (h *ModuleHandler) Access(c *gin.Context) {
	middleware.JSONData(c, 200, gin.H{"allowed": true, "module": c.Param("code")})
}
