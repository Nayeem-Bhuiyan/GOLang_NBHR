package user

import (
	"context"

	"nbhr/internal/domain/entity"
	apperrors "nbhr/internal/domain/errors"
	"nbhr/internal/shared/filter"
	"nbhr/internal/shared/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository defines the user data access contract.
type Repository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindAll(ctx context.Context, p *pagination.Params, f *filter.Params) ([]entity.User, int64, error)
	Update(ctx context.Context, user *entity.User) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	AssignRoles(ctx context.Context, userID uuid.UUID, roles []entity.Role) error
	RemoveRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository constructs a user repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, user *entity.User) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		if len(user.Roles) > 0 {
			return tx.Model(user).Association("Roles").Replace(user.Roles)
		}
		return nil
	})
	if err != nil {
		return apperrors.New(500, "failed to create user", err)
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Roles", "deleted_at IS NULL").
		Preload("Roles.Permissions", "deleted_at IS NULL").
		Where("id = ?", id).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.New(500, "failed to fetch user", err)
	}
	return &user, nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Roles", "deleted_at IS NULL AND is_active = true").
		Preload("Roles.Permissions", "deleted_at IS NULL AND is_active = true").
		Where("email = ?", email).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.New(500, "failed to fetch user", err)
	}
	return &user, nil
}

func (r *repository) FindAll(ctx context.Context, p *pagination.Params, f *filter.Params) ([]entity.User, int64, error) {
	var users []entity.User
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.User{})
	query = filter.Apply(query, f, "first_name", "last_name", "email", "phone")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.New(500, "failed to count users", err)
	}

	err := query.
		Preload("Roles", "deleted_at IS NULL").
		Order(p.OrderClause()).
		Limit(p.PageSize).
		Offset(p.Offset()).
		Find(&users).Error
	if err != nil {
		return nil, 0, apperrors.New(500, "failed to fetch users", err)
	}

	return users, total, nil
}

func (r *repository) Update(ctx context.Context, user *entity.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return apperrors.New(500, "failed to update user", err)
	}
	return nil
}

func (r *repository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.User{}).Error
	if err != nil {
		return apperrors.New(500, "failed to delete user", err)
	}
	return nil
}

func (r *repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, apperrors.New(500, "failed to check user email", err)
	}
	return count > 0, nil
}

func (r *repository) AssignRoles(ctx context.Context, userID uuid.UUID, roles []entity.Role) error {
	var user entity.User
	user.ID = userID
	err := r.db.WithContext(ctx).Model(&user).Association("Roles").Append(roles)
	if err != nil {
		return apperrors.New(500, "failed to assign roles to user", err)
	}
	return nil
}

func (r *repository) RemoveRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	var user entity.User
	user.ID = userID
	roles := make([]entity.Role, 0, len(roleIDs))
	for _, id := range roleIDs {
		roles = append(roles, entity.Role{ID: id})
	}
	err := r.db.WithContext(ctx).Model(&user).Association("Roles").Delete(roles)
	if err != nil {
		return apperrors.New(500, "failed to remove roles from user", err)
	}
	return nil
}

func (r *repository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).
		Update("last_login_at", gorm.Expr("NOW()")).Error
	if err != nil {
		return apperrors.New(500, "failed to update last login", err)
	}
	return nil
}