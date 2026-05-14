package user

import (
	"nbhr/internal/constants"
	apperrors "nbhr/internal/domain/errors"
	"nbhr/internal/dto"
	"nbhr/internal/shared/filter"
	"nbhr/internal/shared/pagination"
	"nbhr/internal/shared/response"
	"nbhr/internal/shared/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the user module.
type Handler struct {
	svc Service
}

// NewHandler constructs a user handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constants.MsgBadRequest, err.Error())
		return
	}
	if errs := validator.Validate(&req); errs != nil {
		response.ValidationError(c, errs)
		return
	}
	user, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, constants.MsgCreated, user)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	user, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, constants.MsgSuccess, user)
}

func (h *Handler) GetAll(c *gin.Context) {
	p := pagination.FromContext(c)
	f := filter.FromContext(c)
	users, total, err := h.svc.GetAll(c.Request.Context(), p, f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, constants.MsgSuccess, users, pagination.NewMeta(total, p))
}

func (h *Handler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constants.MsgBadRequest, err.Error())
		return
	}
	if errs := validator.Validate(&req); errs != nil {
		response.ValidationError(c, errs)
		return
	}
	user, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, constants.MsgUpdated, user)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, constants.MsgDeleted, nil)
}

func (h *Handler) AssignRoles(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constants.MsgBadRequest, err.Error())
		return
	}
	if errs := validator.Validate(&req); errs != nil {
		response.ValidationError(c, errs)
		return
	}
	user, err := h.svc.AssignRoles(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, constants.MsgUpdated, user)
}

func (h *Handler) ToggleActive(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	user, err := h.svc.ToggleActive(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, constants.MsgUpdated, user)
}

// Me returns the currently authenticated user's profile.
func (h *Handler) Me(c *gin.Context) {
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
	user, err := h.svc.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, constants.MsgSuccess, user)
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperrors.Wrap(apperrors.ErrBadRequest, "invalid UUID: "+param, err)
	}
	return id, nil
}