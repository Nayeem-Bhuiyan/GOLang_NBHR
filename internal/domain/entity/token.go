package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RefreshToken stores issued refresh tokens for rotation and revocation.
type RefreshToken struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index"                       json:"user_id"`
	TokenHash string         `gorm:"type:varchar(255);uniqueIndex;not null"          json:"-"`
	ExpiresAt time.Time      `gorm:"not null"                                       json:"expires_at"`
	IsRevoked bool           `gorm:"default:false;not null"                         json:"is_revoked"`
	UserAgent string         `gorm:"type:varchar(255)"                              json:"user_agent"`
	IPAddress string         `gorm:"type:varchar(45)"                               json:"ip_address"`
	User      User           `gorm:"foreignKey:UserID"                              json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                          json:"-"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}

func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}