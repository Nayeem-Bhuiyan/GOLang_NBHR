package permission

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

// Handler handles HTTP requests for the permission module.
type Handler struct {
	svc Service
}

// NewHandler constructs a permission handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Create godoc
// POST /api/v1/permissions
func (h *Handler) Create(c *gin.Context) {
	var req dto.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constants.MsgBadRequest, err.Error())
		return
	}

	if errs := validator.Validate(&req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	perm, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, constants.MsgCreated, perm)
}

// GetByID godoc
// GET /api/v1/permissions/:id
func (h *Handler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	perm, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, constants.MsgSuccess, perm)
}

// GetAll godoc
// GET /api/v1/permissions
func (h *Handler) GetAll(c *gin.Context) {
	p := pagination.FromContext(c)
	f := filter.FromContext(c)

	perms, total, err := h.svc.GetAll(c.Request.Context(), p, f)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Paginated(c, constants.MsgSuccess, perms, pagination.NewMeta(total, p))
}

// Update godoc
// PUT /api/v1/permissions/:id
func (h *Handler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	var req dto.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, constants.MsgBadRequest, err.Error())
		return
	}

	if errs := validator.Validate(&req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	perm, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, constants.MsgUpdated, perm)
}

// Delete godoc
// DELETE /api/v1/permissions/:id
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

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperrors.Wrap(apperrors.ErrBadRequest, "invalid UUID: "+param, err)
	}
	return id, nil
}