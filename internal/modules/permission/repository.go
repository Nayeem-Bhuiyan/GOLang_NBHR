package permission

import (
	"context"

	"nbhr/internal/domain/entity"
	apperrors "nbhr/internal/domain/errors"
	"nbhr/internal/shared/filter"
	"nbhr/internal/shared/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

//go:generate mockgen -source=repository.go -destination=mocks/repository_mock.go

// Repository defines the permission data access contract.
type Repository interface {
	Create(ctx context.Context, perm *entity.Permission) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Permission, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Permission, error)
	FindAll(ctx context.Context, p *pagination.Params, f *filter.Params) ([]entity.Permission, int64, error)
	Update(ctx context.Context, perm *entity.Permission) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]entity.Permission, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository constructs a permission repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, perm *entity.Permission) error {
	if err := r.db.WithContext(ctx).Create(perm).Error; err != nil {
		return apperrors.New(500, "failed to create permission", err)
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Permission, error) {
	var perm entity.Permission
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&perm).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.New(500, "failed to fetch permission", err)
	}
	return &perm, nil
}

func (r *repository) FindBySlug(ctx context.Context, slug string) (*entity.Permission, error) {
	var perm entity.Permission
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&perm).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.New(500, "failed to fetch permission", err)
	}
	return &perm, nil
}

func (r *repository) FindAll(ctx context.Context, p *pagination.Params, f *filter.Params) ([]entity.Permission, int64, error) {
	var perms []entity.Permission
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Permission{})
	query = filter.Apply(query, f, "name", "slug", "resource", "action")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.New(500, "failed to count permissions", err)
	}

	err := query.
		Order(p.OrderClause()).
		Limit(p.PageSize).
		Offset(p.Offset()).
		Find(&perms).Error
	if err != nil {
		return nil, 0, apperrors.New(500, "failed to fetch permissions", err)
	}

	return perms, total, nil
}

func (r *repository) Update(ctx context.Context, perm *entity.Permission) error {
	if err := r.db.WithContext(ctx).Save(perm).Error; err != nil {
		return apperrors.New(500, "failed to update permission", err)
	}
	return nil
}

func (r *repository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Permission{}).Error
	if err != nil {
		return apperrors.New(500, "failed to delete permission", err)
	}
	return nil
}

func (r *repository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Permission{}).Where("slug = ?", slug).Count(&count).Error
	if err != nil {
		return false, apperrors.New(500, "failed to check permission slug", err)
	}
	return count > 0, nil
}

func (r *repository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]entity.Permission, error) {
	var perms []entity.Permission
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&perms).Error
	if err != nil {
		return nil, apperrors.New(500, "failed to fetch permissions by IDs", err)
	}
	return perms, nil
}