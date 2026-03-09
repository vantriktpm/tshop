package postgres

import (
	"fmt"

	"github.com/tshop/backend/pkg/dbutil"
	"gorm.io/gorm"
)

// EnsureSchema creates the database (if missing) and auto-migrates all tables
// for shipping-service. dsn must point to the shipping_db database.
func EnsureSchema(dsn string, db *gorm.DB) error {
	if err := dbutil.EnsureDatabase(dsn, "shipping_db"); err != nil {
		return fmt.Errorf("shipping-service: ensure database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("shipping-service: get sql.DB: %w", err)
	}
	if err := dbutil.EnsureSchema(sqlDB, "public"); err != nil {
		return fmt.Errorf("shipping-service: ensure schema: %w", err)
	}

	if err := db.AutoMigrate(&ShipmentModel{}); err != nil {
		return fmt.Errorf("shipping-service: auto-migrate: %w", err)
	}
	return nil
}
