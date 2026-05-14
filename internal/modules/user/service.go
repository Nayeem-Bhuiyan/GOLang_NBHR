package user

import (
	"context"

	"nbhr/internal/domain/entity"
	apperrors "nbhr/internal/domain/errors"
	"nbhr/internal/dto"
	"nbhr/internal/modules/role"
	"nbhr/internal/shared/crypto"
	"nbhr/internal/shared/filter"
	"nbhr/internal/shared/pagination"

	"github.com/google/uuid"
)

// Service defines the user business logic contract.
type Service interface {
	Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error)
	GetAll(ctx context.Context, p *pagination.Params, f *filter.Params) ([]dto.UserResponse, int64, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AssignRoles(ctx context.Context, userID uuid.UUID, req *dto.AssignRolesRequest) (*dto.UserResponse, error)
	ToggleActive(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error)
}

type service struct {
	repo     Repository
	roleRepo role.Repository
}

// NewService constructs a user service.
func NewService(repo Repository, roleRepo role.Repository) Service {
	return &service{repo: repo, roleRepo: roleRepo}
}

func (s *service) Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	exists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrConflict
	}

	hashed, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, apperrors.New(500, "failed to hash password", err)
	}

	user := &entity.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  hashed,
		Phone:     req.Phone,
		IsActive:  true,
	}

	if len(req.RoleIDs) > 0 {
		roles, err := s.roleRepo.FindByIDs(ctx, req.RoleIDs)
		if err != nil {
			return nil, err
		}
		user.Roles = roles
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

func (s *service) GetAll(ctx context.Context, p *pagination.Params, f *filter.Params) ([]dto.UserResponse, int64, error) {
	users, total, err := s.repo.FindAll(ctx, p, f)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.UserResponse, 0, len(users))
	for i := range users {
		responses = append(responses, *toUserResponse(&users[i]))
	}
	return responses, total, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *service) AssignRoles(ctx context.Context, userID uuid.UUID, req *dto.AssignRolesRequest) (*dto.UserResponse, error) {
	if _, err := s.repo.FindByID(ctx, userID); err != nil {
		return nil, err
	}

	roles, err := s.roleRepo.FindByIDs(ctx, req.RoleIDs)
	if err != nil {
		return nil, err
	}

	if err := s.repo.AssignRoles(ctx, userID, roles); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return toUserResponse(updated), nil
}

func (s *service) ToggleActive(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.IsActive = !user.IsActive
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

func toUserResponse(u *entity.User) *dto.UserResponse {
	resp := &dto.UserResponse{
		ID:          u.ID,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Email:       u.Email,
		Phone:       u.Phone,
		IsActive:    u.IsActive,
		IsVerified:  u.IsVerified,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}

	if len(u.Roles) > 0 {
		resp.Roles = make([]dto.RoleResponse, 0, len(u.Roles))
		for i := range u.Roles {
			r := &u.Roles[i]
			resp.Roles = append(resp.Roles, dto.RoleResponse{
				ID:          r.ID,
				Name:        r.Name,
				Slug:        r.Slug,
				Description: r.Description,
				IsActive:    r.IsActive,
				CreatedAt:   r.CreatedAt,
				UpdatedAt:   r.UpdatedAt,
			})
		}
	}

	return resp
}