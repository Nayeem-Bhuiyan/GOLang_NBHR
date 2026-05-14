package permission

import (
	"context"

	"nbhr/internal/domain/entity"
	apperrors "nbhr/internal/domain/errors"
	"nbhr/internal/dto"
	"nbhr/internal/shared/filter"
	"nbhr/internal/shared/pagination"

	"github.com/google/uuid"
)

// Service defines the permission business logic contract.
type Service interface {
	Create(ctx context.Context, req *dto.CreatePermissionRequest) (*dto.PermissionResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.PermissionResponse, error)
	GetAll(ctx context.Context, p *pagination.Params, f *filter.Params) ([]dto.PermissionResponse, int64, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdatePermissionRequest) (*dto.PermissionResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
}

// NewService constructs a permission service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *dto.CreatePermissionRequest) (*dto.PermissionResponse, error) {
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

	perm := &entity.Permission{
		Name:        req.Name,
		Slug:        req.Slug,
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
		IsActive:    isActive,
	}

	if err := s.repo.Create(ctx, perm); err != nil {
		return nil, err
	}

	return toPermissionResponse(perm), nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*dto.PermissionResponse, error) {
	perm, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toPermissionResponse(perm), nil
}

func (s *service) GetAll(ctx context.Context, p *pagination.Params, f *filter.Params) ([]dto.PermissionResponse, int64, error) {
	perms, total, err := s.repo.FindAll(ctx, p, f)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.PermissionResponse, 0, len(perms))
	for i := range perms {
		responses = append(responses, *toPermissionResponse(&perms[i]))
	}

	return responses, total, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req *dto.UpdatePermissionRequest) (*dto.PermissionResponse, error) {
	perm, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		perm.Name = req.Name
	}
	if req.Description != "" {
		perm.Description = req.Description
	}
	if req.IsActive != nil {
		perm.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, perm); err != nil {
		return nil, err
	}

	return toPermissionResponse(perm), nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.SoftDelete(ctx, id)
}

func toPermissionResponse(p *entity.Permission) *dto.PermissionResponse {
	return &dto.PermissionResponse{
		ID:          p.ID,
		Name:        p.Name,
		Slug:        p.Slug,
		Resource:    p.Resource,
		Action:      p.Action,
		Description: p.Description,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}