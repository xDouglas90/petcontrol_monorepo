package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error         string `json:"error"`
	Code          string `json:"code"`
	CorrelationID string `json:"correlation_id,omitempty"`
}

func JSONData(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

func JSONError(c *gin.Context, status int, code string, message string) {
	if message == "" {
		message = http.StatusText(status)
	}
	if code == "" {
		code = defaultErrorCode(status)
	}

	c.AbortWithStatusJSON(status, ErrorResponse{
		Error:         message,
		Code:          code,
		CorrelationID: GetCorrelationID(c),
	})
}

func defaultErrorCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	case http.StatusUnprocessableEntity:
		return "unprocessable_entity"
	case http.StatusTooManyRequests:
		return "too_many_requests"
	default:
		if status >= 500 {
			return "internal_error"
		}
		return fmt.Sprintf("http_%d", status)
	}
}
