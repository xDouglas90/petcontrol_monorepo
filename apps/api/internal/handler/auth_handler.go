package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type AuthHandler struct {
	service *service.AuthService
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.service.Login(c.Request.Context(), req.Email, req.Password, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		status := apperror.HTTPStatus(err)
		message := "internal error"
		switch {
		case errors.Is(err, apperror.ErrInvalidCredentials):
			message = "invalid credentials"
		case errors.Is(err, apperror.ErrAccountInactive):
			message = "account inactive"
		case errors.Is(err, apperror.ErrAccountLocked):
			message = "account locked"
		case errors.Is(err, apperror.ErrEmailNotVerified):
			message = "email not verified"
		case errors.Is(err, apperror.ErrUnprocessableEntity):
			message = "invalid credentials payload"
		}
		c.JSON(status, gin.H{"error": message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}
