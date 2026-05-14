package role

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

// Handler handles HTTP requests for the role module.
type Handler struct {
	svc Service
}

// NewHandler constructs a role handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constants.MsgBadRequest, err.Error())
		return
	}
	if errs := validator.Validate(&req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	role, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, constants.MsgCreated, role)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	role, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, constants.MsgSuccess, role)
}

func (h *Handler) GetAll(c *gin.Context) {
	p := pagination.FromContext(c)
	f := filter.FromContext(c)
	roles, total, err := h.svc.GetAll(c.Request.Context(), p, f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, constants.MsgSuccess, roles, pagination.NewMeta(total, p))
}

func (h *Handler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constants.MsgBadRequest, err.Error())
		return
	}
	if errs := validator.Validate(&req); errs != nil {
		response.ValidationError(c, errs)
		return
	}
	role, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, constants.MsgUpdated, role)
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

func (h *Handler) AssignPermissions(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constants.MsgBadRequest, err.Error())
		return
	}
	if errs := validator.Validate(&req); errs != nil {
		response.ValidationError(c, errs)
		return
	}
	role, err := h.svc.AssignPermissions(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, constants.MsgUpdated, role)
}

func (h *Handler) RemovePermissions(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constants.MsgBadRequest, err.Error())
		return
	}
	if errs := validator.Validate(&req); errs != nil {
		response.ValidationError(c, errs)
		return
	}
	role, err := h.svc.RemovePermissions(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, constants.MsgUpdated, role)
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperrors.Wrap(apperrors.ErrBadRequest, "invalid UUID: "+param, err)
	}
	return id, nil
}