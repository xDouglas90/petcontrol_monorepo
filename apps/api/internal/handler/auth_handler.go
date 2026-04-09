package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
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
		middleware.JSONError(c, 422, "invalid_request_body", "invalid request body")
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
		case errors.Is(err, apperror.ErrForbidden):
			message = "no active company membership"
		case errors.Is(err, apperror.ErrUnprocessableEntity):
			message = "invalid credentials payload"
		}
		middleware.JSONError(c, status, "login_failed", message)
		return
	}

	middleware.JSONData(c, 200, result)
}
