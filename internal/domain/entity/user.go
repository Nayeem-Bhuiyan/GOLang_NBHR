package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a system user.
type User struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FirstName   string         `gorm:"type:varchar(100);not null"                     json:"first_name"`
	LastName    string         `gorm:"type:varchar(100);not null"                     json:"last_name"`
	Email       string         `gorm:"type:varchar(255);uniqueIndex;not null"          json:"email"`
	Password    string         `gorm:"type:varchar(255);not null"                     json:"-"`
	Phone       string         `gorm:"type:varchar(20)"                               json:"phone"`
	IsActive    bool           `gorm:"default:true;not null"                          json:"is_active"`
	IsVerified  bool           `gorm:"default:false;not null"                         json:"is_verified"`
	LastLoginAt *time.Time     `json:"last_login_at"`
	Roles       []Role         `gorm:"many2many:user_roles;"                          json:"roles,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                          json:"-"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// HasRole checks if user has a specific role by slug.
func (u *User) HasRole(slug string) bool {
	for _, r := range u.Roles {
		if r.Slug == slug && r.IsActive {
			return true
		}
	}
	return false
}

// GetPermissionSlugs returns all permission slugs across all roles.
func (u *User) GetPermissionSlugs() []string {
	seen := make(map[string]struct{})
	slugs := make([]string, 0)
	for _, r := range u.Roles {
		if !r.IsActive {
			continue
		}
		for _, p := range r.Permissions {
			if p.IsActive {
				if _, exists := seen[p.Slug]; !exists {
					seen[p.Slug] = struct{}{}
					slugs = append(slugs, p.Slug)
				}
			}
		}
	}
	return slugs
}