package bootstrap

import (
	"log"

	"github.com/tshop/backend/services/product-service/internal/container"
	"gorm.io/gorm"
)

func New() *container.Container {
	if err := loadEnv(); err != nil {
		log.Printf("bootstrap: load .env: %v (using OS env)", err)
	}

	db, err := NewPostgres()
	if err != nil {
		log.Fatal("bootstrap: postgres: ", err)
	}

	return container.New(db)
}

// NewDB is used by tests or when only DB is needed.
func NewDB() *gorm.DB {
	_ = loadEnv()
	db, err := NewPostgres()
	if err != nil {
		return nil
	}
	return db
}
