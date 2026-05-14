package auth

import (
	"context"
	"time"

	"nbhr/internal/domain/entity"
	apperrors "nbhr/internal/domain/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TokenRepository defines the refresh token data access contract.
type TokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*entity.RefreshToken, error)
	RevokeByHash(ctx context.Context, hash string) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]entity.RefreshToken, error)
}

type tokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository constructs a refresh token repository.
func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return apperrors.New(500, "failed to store refresh token", err)
	}
	return nil
}

func (r *tokenRepository) FindByHash(ctx context.Context, hash string) (*entity.RefreshToken, error) {
	var token entity.RefreshToken
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("token_hash = ? AND is_revoked = false", hash).
		First(&token).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrTokenInvalid
		}
		return nil, apperrors.New(500, "failed to fetch refresh token", err)
	}
	return &token, nil
}

func (r *tokenRepository) RevokeByHash(ctx context.Context, hash string) error {
	err := r.db.WithContext(ctx).
		Model(&entity.RefreshToken{}).
		Where("token_hash = ?", hash).
		Update("is_revoked", true).Error
	if err != nil {
		return apperrors.New(500, "failed to revoke refresh token", err)
	}
	return nil
}

func (r *tokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	err := r.db.WithContext(ctx).
		Model(&entity.RefreshToken{}).
		Where("user_id = ? AND is_revoked = false", userID).
		Update("is_revoked", true).Error
	if err != nil {
		return apperrors.New(500, "failed to revoke user tokens", err)
	}
	return nil
}

func (r *tokenRepository) DeleteExpired(ctx context.Context) error {
	err := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&entity.RefreshToken{}).Error
	if err != nil {
		return apperrors.New(500, "failed to delete expired tokens", err)
	}
	return nil
}

func (r *tokenRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]entity.RefreshToken, error) {
	var tokens []entity.RefreshToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_revoked = false AND expires_at > ?", userID, time.Now()).
		Find(&tokens).Error
	if err != nil {
		return nil, apperrors.New(500, "failed to fetch active tokens", err)
	}
	return tokens, nil
}