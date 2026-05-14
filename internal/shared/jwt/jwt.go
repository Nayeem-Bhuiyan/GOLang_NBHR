package jwt

import (
	"errors"
	"fmt"
	"time"

	apperrors "nbhr/internal/domain/errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents the JWT payload.
type Claims struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	TokenType   string    `json:"token_type"`
	jwt.RegisteredClaims
}

// Manager handles JWT creation and validation.
type Manager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewManager constructs a JWT Manager.
func NewManager(accessSecret, refreshSecret string, accessExpiry, refreshExpiry time.Duration) *Manager {
	return &Manager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateAccessToken creates a signed access JWT for the given user.
func (m *Manager) GenerateAccessToken(userID uuid.UUID, email string, roles, permissions []string) (string, error) {
	claims := &Claims{
		UserID:      userID,
		Email:       email,
		Roles:       roles,
		Permissions: permissions,
		TokenType:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessExpiry)),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.accessSecret)
}

// GenerateRefreshToken creates a signed refresh JWT for the given user.
func (m *Manager) GenerateRefreshToken(userID uuid.UUID, email string) (string, error) {
	claims := &Claims{
		UserID:    userID,
		Email:     email,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshExpiry)),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.refreshSecret)
}

// ValidateAccessToken parses and validates an access token.
func (m *Manager) ValidateAccessToken(tokenStr string) (*Claims, error) {
	return m.validate(tokenStr, m.accessSecret)
}

// ValidateRefreshToken parses and validates a refresh token.
func (m *Manager) ValidateRefreshToken(tokenStr string) (*Claims, error) {
	return m.validate(tokenStr, m.refreshSecret)
}

func (m *Manager) validate(tokenStr string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.ErrTokenExpired
		}
		return nil, apperrors.ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, apperrors.ErrTokenInvalid
	}

	return claims, nil
}

// RefreshExpiry returns the configured refresh token expiry duration.
func (m *Manager) RefreshExpiry() time.Duration {
	return m.refreshExpiry
}