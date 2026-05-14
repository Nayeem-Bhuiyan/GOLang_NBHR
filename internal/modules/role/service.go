package role

import (
	"context"

	"nbhr/internal/domain/entity"
	apperrors "nbhr/internal/domain/errors"
	"nbhr/internal/dto"
	"nbhr/internal/modules/permission"
	"nbhr/internal/shared/filter"
	"nbhr/internal/shared/pagination"

	"github.com/google/uuid"
)

// Service defines the role business logic contract.
type Service interface {
	Create(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.RoleResponse, error)
	GetAll(ctx context.Context, p *pagination.Params, f *filter.Params) ([]dto.RoleResponse, int64, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AssignPermissions(ctx context.Context, roleID uuid.UUID, req *dto.AssignPermissionsRequest) (*dto.RoleResponse, error)
	RemovePermissions(ctx context.Context, roleID uuid.UUID, req *dto.AssignPermissionsRequest) (*dto.RoleResponse, error)
}

type service struct {
	repo     Repository
	permRepo permission.Repository
}

// NewService constructs a role service.
func NewService(repo Repository, permRepo permission.Repository) Service {
	return &service{repo: repo, permRepo: permRepo}
}

func (s *service) Create(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	exists, err := s.repo.ExistsBySlug(ctx, req.Slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrConflict
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	role := &entity.Role{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		IsActive:    isActive,
	}

	if len(req.PermIDs) > 0 {
		perms, err := s.permRepo.FindByIDs(ctx, req.PermIDs)
		if err != nil {
			return nil, err
		}
		role.Permissions = perms
	}

	if err := s.repo.Create(ctx, role); err != nil {
		return nil, err
	}

	return toRoleResponse(role), nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*dto.RoleResponse, error) {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toRoleResponse(role), nil
}

func (s *service) GetAll(ctx context.Context, p *pagination.Params, f *filter.Params) ([]dto.RoleResponse, int64, error) {
	roles, total, err := s.repo.FindAll(ctx, p, f)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.RoleResponse, 0, len(roles))
	for i := range roles {
		responses = append(responses, *toRoleResponse(&roles[i]))
	}

	return responses, total, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.IsActive != nil {
		role.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, role); err != nil {
		return nil, err
	}

	return toRoleResponse(role), nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *service) AssignPermissions(ctx context.Context, roleID uuid.UUID, req *dto.AssignPermissionsRequest) (*dto.RoleResponse, error) {
	role, err := s.repo.FindByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	perms, err := s.permRepo.FindByIDs(ctx, req.PermissionIDs)
	if err != nil {
		return nil, err
	}

	if err := s.repo.AssignPermissions(ctx, role.ID, perms); err != nil {
		return nil, err
	}

	// Refresh role with new permissions
	updated, err := s.repo.FindByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	return toRoleResponse(updated), nil
}

func (s *service) RemovePermissions(ctx context.Context, roleID uuid.UUID, req *dto.AssignPermissionsRequest) (*dto.RoleResponse, error) {
	if _, err := s.repo.FindByID(ctx, roleID); err != nil {
		return nil, err
	}

	if err := s.repo.RemovePermissions(ctx, roleID, req.PermissionIDs); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	return toRoleResponse(updated), nil
}

func toRoleResponse(r *entity.Role) *dto.RoleResponse {
	resp := &dto.RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		Slug:        r.Slug,
		Description: r.Description,
		IsActive:    r.IsActive,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}

	if len(r.Permissions) > 0 {
		resp.Permissions = make([]dto.PermissionResponse, 0, len(r.Permissions))
		for i := range r.Permissions {
			p := &r.Permissions[i]
			resp.Permissions = append(resp.Permissions, dto.PermissionResponse{
				ID:          p.ID,
				Name:        p.Name,
				Slug:        p.Slug,
				Resource:    p.Resource,
				Action:      p.Action,
				Description: p.Description,
				IsActive:    p.IsActive,
				CreatedAt:   p.CreatedAt,
				UpdatedAt:   p.UpdatedAt,
			})
		}
	}

	return resp
}