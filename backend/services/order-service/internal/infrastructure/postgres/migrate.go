package postgres

import (
	"fmt"

	"github.com/tshop/backend/pkg/dbutil"
	"gorm.io/gorm"
)

// EnsureSchema creates the database (if missing) and auto-migrates all tables
// for order-service. dsn must point to the order_db database.
func EnsureSchema(dsn string, db *gorm.DB) error {
	if err := dbutil.EnsureDatabase(dsn, "order_db"); err != nil {
		return fmt.Errorf("order-service: ensure database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("order-service: get sql.DB: %w", err)
	}
	if err := dbutil.EnsureSchema(sqlDB, "public"); err != nil {
		return fmt.Errorf("order-service: ensure schema: %w", err)
	}

	if err := db.AutoMigrate(&OrderModel{}, &OrderItemModel{}); err != nil {
		return fmt.Errorf("order-service: auto-migrate: %w", err)
	}
	return nil
}
