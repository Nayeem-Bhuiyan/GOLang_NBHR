package filter

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Params holds generic filtering and search parameters.
type Params struct {
	Search   string
	Status   string
	IsActive *bool
}

// FromContext extracts filter params from gin context.
func FromContext(c *gin.Context) *Params {
	p := &Params{
		Search: c.Query("search"),
		Status: c.Query("status"),
	}

	if activeStr := c.Query("is_active"); activeStr != "" {
		active := activeStr == "true" || activeStr == "1"
		p.IsActive = &active
	}

	return p
}

// Apply applies filter conditions to a GORM query.
func Apply(db *gorm.DB, p *Params, searchFields ...string) *gorm.DB {
	if p == nil {
		return db
	}

	if p.IsActive != nil {
		db = db.Where("is_active = ?", *p.IsActive)
	}

	if p.Search != "" && len(searchFields) > 0 {
		query := ""
		args := make([]interface{}, 0, len(searchFields))
		for i, field := range searchFields {
			if i > 0 {
				query += " OR "
			}
			query += field + " ILIKE ?"
			args = append(args, "%"+p.Search+"%")
		}
		db = db.Where("("+query+")", args...)
	}

	return db
}