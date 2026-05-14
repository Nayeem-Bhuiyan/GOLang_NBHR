package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Permission represents a granular action on a resource.
type Permission struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(150);uniqueIndex;not null"         json:"name"`
	Slug        string         `gorm:"type:varchar(150);uniqueIndex;not null"         json:"slug"`
	Resource    string         `gorm:"type:varchar(100);not null;index"               json:"resource"`
	Action      string         `gorm:"type:varchar(50);not null"                      json:"action"`
	Description string         `gorm:"type:text"                                      json:"description"`
	IsActive    bool           `gorm:"default:true;not null"                          json:"is_active"`
	Roles       []Role         `gorm:"many2many:role_permissions;"                    json:"-"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                          json:"-"`
}

func (Permission) TableName() string {
	return "permissions"
}

func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}