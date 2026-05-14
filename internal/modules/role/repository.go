package role

import (
	"context"

	"nbhr/internal/domain/entity"
	apperrors "nbhr/internal/domain/errors"
	"nbhr/internal/shared/filter"
	"nbhr/internal/shared/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository defines the role data access contract.
type Repository interface {
	Create(ctx context.Context, role *entity.Role) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Role, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Role, error)
	FindAll(ctx context.Context, p *pagination.Params, f *filter.Params) ([]entity.Role, int64, error)
	Update(ctx context.Context, role *entity.Role) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	AssignPermissions(ctx context.Context, roleID uuid.UUID, permissions []entity.Permission) error
	RemovePermissions(ctx context.Context, roleID uuid.UUID, permIDs []uuid.UUID) error
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]entity.Role, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository constructs a role repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, role *entity.Role) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(role).Error; err != nil {
			return err
		}
		if len(role.Permissions) > 0 {
			if err := tx.Model(role).Association("Permissions").Replace(role.Permissions); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return apperrors.New(500, "failed to create role", err)
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions", "deleted_at IS NULL").
		Where("id = ?", id).
		First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.New(500, "failed to fetch role", err)
	}
	return &role, nil
}

func (r *repository) FindBySlug(ctx context.Context, slug string) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions", "deleted_at IS NULL").
		Where("slug = ?", slug).
		First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.New(500, "failed to fetch role", err)
	}
	return &role, nil
}

func (r *repository) FindAll(ctx context.Context, p *pagination.Params, f *filter.Params) ([]entity.Role, int64, error) {
	var roles []entity.Role
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Role{})
	query = filter.Apply(query, f, "name", "slug")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.New(500, "failed to count roles", err)
	}

	err := query.
		Preload("Permissions", "deleted_at IS NULL").
		Order(p.OrderClause()).
		Limit(p.PageSize).
		Offset(p.Offset()).
		Find(&roles).Error
	if err != nil {
		return nil, 0, apperrors.New(500, "failed to fetch roles", err)
	}

	return roles, total, nil
}

func (r *repository) Update(ctx context.Context, role *entity.Role) error {
	if err := r.db.WithContext(ctx).Save(role).Error; err != nil {
		return apperrors.New(500, "failed to update role", err)
	}
	return nil
}

func (r *repository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Role{}).Error
	if err != nil {
		return apperrors.New(500, "failed to delete role", err)
	}
	return nil
}

func (r *repository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Role{}).Where("slug = ?", slug).Count(&count).Error
	if err != nil {
		return false, apperrors.New(500, "failed to check role slug", err)
	}
	return count > 0, nil
}

func (r *repository) AssignPermissions(ctx context.Context, roleID uuid.UUID, permissions []entity.Permission) error {
	var role entity.Role
	role.ID = roleID
	err := r.db.WithContext(ctx).Model(&role).Association("Permissions").Append(permissions)
	if err != nil {
		return apperrors.New(500, "failed to assign permissions to role", err)
	}
	return nil
}

func (r *repository) RemovePermissions(ctx context.Context, roleID uuid.UUID, permIDs []uuid.UUID) error {
	var role entity.Role
	role.ID = roleID
	perms := make([]entity.Permission, 0, len(permIDs))
	for _, id := range permIDs {
		perms = append(perms, entity.Permission{ID: id})
	}
	err := r.db.WithContext(ctx).Model(&role).Association("Permissions").Delete(perms)
	if err != nil {
		return apperrors.New(500, "failed to remove permissions from role", err)
	}
	return nil
}

func (r *repository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]entity.Role, error) {
	var roles []entity.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions", "deleted_at IS NULL").
		Where("id IN ?", ids).
		Find(&roles).Error
	if err != nil {
		return nil, apperrors.New(500, "failed to fetch roles by IDs", err)
	}
	return roles, nil
}