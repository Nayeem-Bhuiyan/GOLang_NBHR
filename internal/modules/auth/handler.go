package auth

import (
	"strings"

	"nbhr/internal/constants"
	apperrors "nbhr/internal/domain/errors"
	"nbhr/internal/dto"
	"nbhr/internal/shared/response"
	"nbhr/internal/shared/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the auth module.
type Handler struct {
	svc Service
}

// NewHandler constructs an auth handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Login godoc
// POST /api/v1/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constants.MsgBadRequest, err.Error())
		return
	}

	if errs := validator.Validate(&req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	tokens, err := h.svc.Login(
		c.Request.Context(),
		&req,
		c.Request.UserAgent(),
		c.ClientIP(),
	)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "login successful", tokens)
}

// Register godoc
// POST /api/v1/auth/register
func (h *Handler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constants.MsgBadRequest, err.Error())
		return
	}

	if errs := validator.Validate(&req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	user, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "registration successful", user)
}

// RefreshToken godoc
// POST /api/v1/auth/refresh
func (h *Handler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constants.MsgBadRequest, err.Error())
		return
	}

	if errs := validator.Validate(&req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	tokens, err := h.svc.RefreshToken(
		c.Request.Context(),
		&req,
		c.Request.UserAgent(),
		c.ClientIP(),
	)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "token refreshed", tokens)
}

// Logout godoc
// POST /api/v1/auth/logout
func (h *Handler) Logout(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constants.MsgBadRequest, err.Error())
		return
	}

	if err := h.svc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "logged out successfully", nil)
}

// LogoutAll godoc
// POST /api/v1/auth/logout-all
func (h *Handler) LogoutAll(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constants.MsgBadRequest, err.Error())
		return
	}

	if err := h.svc.LogoutAll(c.Request.Context(), req.RefreshToken); err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "logged out from all devices", nil)
}

// ChangePassword godoc
// POST /api/v1/auth/change-password
func (h *Handler) ChangePassword(c *gin.Context) {
	userIDRaw, exists := c.Get(constants.ContextKeyUserID)
	if !exists {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}

	userID, ok := userIDRaw.(uuid.UUID)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constants.MsgBadRequest, err.Error())
		return
	}

	if errs := validator.Validate(&req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), userID.String(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "password changed successfully", nil)
}

// extractBearerToken strips "Bearer " prefix from the Authorization header.
func extractBearerToken(c *gin.Context) string {
	header := c.GetHeader(constants.HeaderAuthorization)
	if strings.HasPrefix(header, constants.BearerPrefix) {
		return strings.TrimPrefix(header, constants.BearerPrefix)
	}
	return ""
}