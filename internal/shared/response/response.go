package response

import (
	"net/http"

	apperrors "nbhr/internal/domain/errors"
	"nbhr/internal/shared/pagination"

	"github.com/gin-gonic/gin"
)

// Response is the standard API envelope.
type Response struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Error      interface{} `json:"error,omitempty"`
	Meta       interface{} `json:"meta,omitempty"`
	RequestID  string      `json:"request_id,omitempty"`
}

// ErrorDetail provides structured error information.
type ErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// OK sends a 200 success response.
func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// Created sends a 201 created response.
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// Paginated sends a paginated success response.
func Paginated(c *gin.Context, message string, data interface{}, meta *pagination.Meta) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Meta:      meta,
		RequestID: getRequestID(c),
	})
}

// Error sends a structured error response.
func Error(c *gin.Context, err error) {
	var appErr *apperrors.AppError
	code := http.StatusInternalServerError
	message := "internal server error"
	detail := ""

	if e, ok := err.(*apperrors.AppError); ok {
		appErr = e
		code = appErr.Code
		message = appErr.Message
		detail = appErr.Detail
	}

	c.JSON(code, Response{
		Success: false,
		Message: message,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Detail:  detail,
		},
		RequestID: getRequestID(c),
	})
}

// BadRequest sends a 400 bad request response.
func BadRequest(c *gin.Context, message string, detail interface{}) {
	c.JSON(http.StatusBadRequest, Response{
		Success:   false,
		Message:   message,
		Error:     detail,
		RequestID: getRequestID(c),
	})
}

// Unauthorized sends a 401 response.
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Success:   false,
		Message:   message,
		RequestID: getRequestID(c),
	})
}

// Forbidden sends a 403 response.
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Success:   false,
		Message:   message,
		RequestID: getRequestID(c),
	})
}

// NotFound sends a 404 response.
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Success:   false,
		Message:   message,
		RequestID: getRequestID(c),
	})
}

// ValidationError sends a 422 validation error response.
func ValidationError(c *gin.Context, errors interface{}) {
	c.JSON(http.StatusUnprocessableEntity, Response{
		Success:   false,
		Message:   "validation failed",
		Error:     errors,
		RequestID: getRequestID(c),
	})
}

func getRequestID(c *gin.Context) string {
	if id, exists := c.Get("request_id"); exists {
		if sid, ok := id.(string); ok {
			return sid
		}
	}
	return ""
}