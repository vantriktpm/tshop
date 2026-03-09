package postgres

import (
	"fmt"

	"github.com/tshop/backend/pkg/dbutil"
	"gorm.io/gorm"
)

// EnsureSchema creates the database (if missing) and auto-migrates all tables
// for promotion-service. dsn must point to the promotion_db database.
func EnsureSchema(dsn string, db *gorm.DB) error {
	if err := dbutil.EnsureDatabase(dsn, "promotion_db"); err != nil {
		return fmt.Errorf("promotion-service: ensure database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("promotion-service: get sql.DB: %w", err)
	}
	if err := dbutil.EnsureSchema(sqlDB, "public"); err != nil {
		return fmt.Errorf("promotion-service: ensure schema: %w", err)
	}

	if err := db.AutoMigrate(&PromotionModel{}); err != nil {
		return fmt.Errorf("promotion-service: auto-migrate: %w", err)
	}
	return nil
}
