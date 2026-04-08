package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type CompanyHandler struct {
	service *service.CompanyService
}

func NewCompanyHandler(service *service.CompanyService) *CompanyHandler {
	return &CompanyHandler{service: service}
}

func (h *CompanyHandler) List(c *gin.Context) {
	companies, err := h.service.ListCompanies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list companies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": companies})
}

func (h *CompanyHandler) Current(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "company context required"})
		return
	}

	company, err := h.service.GetCurrentCompany(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(apperror.HTTPStatus(err), gin.H{"error": "failed to get current company"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": company})
}
