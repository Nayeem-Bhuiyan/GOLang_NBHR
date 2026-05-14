package migrations

import (
	"fmt"

	"nbhr/internal/domain/entity"

	"gorm.io/gorm"
)

// Run executes auto-migrations for all domain entities.
// In production, prefer versioned migration files (golang-migrate).
func Run(db *gorm.DB) error {
	// Enable uuid-ossp extension for gen_random_uuid()
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`).Error; err != nil {
		return fmt.Errorf("failed to enable pgcrypto extension: %w", err)
	}

	models := []interface{}{
		&entity.Permission{},
		&entity.Role{},
		&entity.User{},
		&entity.RefreshToken{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("migration failed for %T: %w", model, err)
		}
	}

	return nil
}