package bootstrap

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/tshop/backend/pkg/dbutil"
	"github.com/tshop/backend/services/product-service/internal/repository"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewPostgres creates a PostgreSQL connection using DB_DSN from config (.env) and runs migrations.
func NewPostgres() (*gorm.DB, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=1 dbname=product_db port=5432 sslmode=disable"
	}
	if err := dbutil.EnsureDatabase(dsn, "product_db"); err != nil {
		return nil, fmt.Errorf("ensure database: %w", err)
	}
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(gormpostgres.New(gormpostgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := repository.Migrate(db); err != nil {
		return nil, err
	}
	return db, nil
}
