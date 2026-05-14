package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents an RBAC role in the system.
type Role struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(100);uniqueIndex;not null"         json:"name"`
	Slug        string         `gorm:"type:varchar(100);uniqueIndex;not null"         json:"slug"`
	Description string         `gorm:"type:text"                                     json:"description"`
	IsActive    bool           `gorm:"default:true;not null"                         json:"is_active"`
	Permissions []Permission   `gorm:"many2many:role_permissions;"                   json:"permissions,omitempty"`
	Users       []User         `gorm:"many2many:user_roles;"                         json:"-"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                         json:"-"`
}

func (Role) TableName() string {
	return "roles"
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}