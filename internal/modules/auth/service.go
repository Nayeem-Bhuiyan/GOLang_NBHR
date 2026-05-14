package auth

import (
	"context"
	"time"

	"nbhr/internal/domain/entity"
	apperrors "nbhr/internal/domain/errors"
	"nbhr/internal/dto"
	"nbhr/internal/modules/user"
	"nbhr/internal/shared/crypto"
	"nbhr/internal/shared/jwt"
)

// Service defines the auth business logic contract.
type Service interface {
	Login(ctx context.Context, req *dto.LoginRequest, userAgent, ip string) (*dto.TokenResponse, error)
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error)
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest, userAgent, ip string) (*dto.TokenResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, refreshToken string) error
	ChangePassword(ctx context.Context, userID string, req *dto.ChangePasswordRequest) error
}

type service struct {
	userRepo   user.Repository
	tokenRepo  TokenRepository
	jwtManager *jwt.Manager
}

// NewService constructs an auth service.
func NewService(
	userRepo user.Repository,
	tokenRepo TokenRepository,
	jwtManager *jwt.Manager,
) Service {
	return &service{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtManager: jwtManager,
	}
}

func (s *service) Login(ctx context.Context, req *dto.LoginRequest, userAgent, ip string) (*dto.TokenResponse, error) {
	u, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// Return generic error to prevent user enumeration
		return nil, apperrors.ErrInvalidCredentials
	}

	if !u.IsActive {
		return nil, apperrors.ErrAccountInactive
	}

	if !crypto.CheckPassword(req.Password, u.Password) {
		return nil, apperrors.ErrInvalidCredentials
	}

	tokens, err := s.issueTokenPair(ctx, u, userAgent, ip)
	if err != nil {
		return nil, err
	}

	// Fire-and-forget last login update (non-critical)
	_ = s.userRepo.UpdateLastLogin(ctx, u.ID)

	return tokens, nil
}

func (s *service) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error) {
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
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

	u := &entity.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  hashed,
		Phone:     req.Phone,
		IsActive:  true,
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	return toUserResponse(u), nil
}

func (s *service) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest, userAgent, ip string) (*dto.TokenResponse, error) {
	// Validate the refresh JWT signature and expiry
	claims, err := s.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Check the token exists and is not revoked in DB
	tokenHash := crypto.HashToken(req.RefreshToken)
	storedToken, err := s.tokenRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil, apperrors.ErrTokenInvalid
	}

	if storedToken.IsRevoked || storedToken.IsExpired() {
		return nil, apperrors.ErrTokenRevoked
	}

	// Revoke old token (rotation)
	if err := s.tokenRepo.RevokeByHash(ctx, tokenHash); err != nil {
		return nil, err
	}

	// Reload user with fresh roles/permissions
	u, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	if !u.IsActive {
		return nil, apperrors.ErrAccountInactive
	}

	return s.issueTokenPair(ctx, u, userAgent, ip)
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := crypto.HashToken(refreshToken)
	return s.tokenRepo.RevokeByHash(ctx, tokenHash)
}

func (s *service) LogoutAll(ctx context.Context, refreshToken string) error {
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return err
	}
	return s.tokenRepo.RevokeAllForUser(ctx, claims.UserID)
}

func (s *service) ChangePassword(ctx context.Context, userIDStr string, req *dto.ChangePasswordRequest) error {
	u, err := s.userRepo.FindByEmail(ctx, userIDStr)
	if err != nil {
		return err
	}

	if !crypto.CheckPassword(req.CurrentPassword, u.Password) {
		return apperrors.ErrInvalidCredentials
	}

	hashed, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		return apperrors.New(500, "failed to hash new password", err)
	}

	u.Password = hashed
	return s.userRepo.Update(ctx, u)
}

// issueTokenPair generates access + refresh tokens and persists the refresh token.
func (s *service) issueTokenPair(ctx context.Context, u *entity.User, userAgent, ip string) (*dto.TokenResponse, error) {
	roleSlugs := make([]string, 0, len(u.Roles))
	for _, r := range u.Roles {
		if r.IsActive {
			roleSlugs = append(roleSlugs, r.Slug)
		}
	}

	permSlugs := u.GetPermissionSlugs()

	accessToken, err := s.jwtManager.GenerateAccessToken(u.ID, u.Email, roleSlugs, permSlugs)
	if err != nil {
		return nil, apperrors.New(500, "failed to generate access token", err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(u.ID, u.Email)
	if err != nil {
		return nil, apperrors.New(500, "failed to generate refresh token", err)
	}

	// Store hashed refresh token
	tokenHash := crypto.HashToken(refreshToken)
	storedToken := &entity.RefreshToken{
		UserID:    u.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.jwtManager.RefreshExpiry()),
		UserAgent: userAgent,
		IPAddress: ip,
	}

	if err := s.tokenRepo.Create(ctx, storedToken); err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtManager.RefreshExpiry().Seconds()),
	}, nil
}

func toUserResponse(u *entity.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:         u.ID,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Email:      u.Email,
		Phone:      u.Phone,
		IsActive:   u.IsActive,
		IsVerified: u.IsVerified,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}