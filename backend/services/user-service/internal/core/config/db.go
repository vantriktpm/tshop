package config

import (
	"bufio"
	"database/sql"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewPostgresDB creates a *gorm.DB connection for user-service using config or env.
func NewPostgresDB() (*gorm.DB, error) {
	// Priority: user.config (DB_DSN=...) > env USER_SERVICE_DB_DSN > hard-coded default
	if fileDSN, ok := loadFromUserConfig("DB_DSN"); ok && fileDSN != "" {
		return openWithPGX(fileDSN)
	}

	if dsn := os.Getenv("USER_SERVICE_DB_DSN"); dsn != "" {
		return openWithPGX(dsn)
	}

	dsn := "host=localhost user=postgres password=1 dbname=tshop port=5432 sslmode=disable"
	return openWithPGX(dsn)
}

// openWithPGX opens a *sql.DB using pgx stdlib and wraps it with GORM.
func openWithPGX(dsn string) (*gorm.DB, error) {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	// Optional: tune pool here (SetMaxOpenConns, etc.) if needed.
	return gorm.Open(gormpostgres.New(gormpostgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
}

// loadFromUserConfig reads config (key=value per line) and returns value for the given key.
// Supports both user.config and .env style files.
// Search order:
// 1) USER_SERVICE_CONFIG_PATH (if set)
// 2) .env or user.config in current working directory
// 3) known service paths when running from workspace/backend root
// 4) .env or user.config next to executable (and its parent dir)
func loadFromUserConfig(key string) (string, bool) {
	// 1) Explicit path from env
	if p := os.Getenv("USER_SERVICE_CONFIG_PATH"); p != "" {
		if v, ok := readConfigKey(p, key); ok {
			return v, true
		}
	}

	// 2) .env / user.config in current working directory
	for _, name := range []string{".env", "user.config"} {
		if v, ok := readConfigKey(name, key); ok {
			return v, true
		}
	}

	// 3) known service paths when running from workspace or backend root
	for _, p := range []string{
		"backend/services/user-service/cmd/.env",
	} {
		if v, ok := readConfigKey(p, key); ok {
			return v, true
		}
	}

	// 4) .env / user.config next to executable (useful when service is started from another cwd)
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		// same directory as executable
		for _, name := range []string{".env", "user.config"} {
			if v, ok := readConfigKey(filepath.Join(exeDir, name), key); ok {
				return v, true
			}
		}
		// parent directory (e.g. exe in cmd/, config in service root)
		parent := filepath.Dir(exeDir)
		for _, name := range []string{".env", "user.config"} {
			if v, ok := readConfigKey(filepath.Join(parent, name), key); ok {
				return v, true
			}
		}
	}

	return "", false
}

// readConfigKey scans a config file (key=value per line) and returns the value for key.
func readConfigKey(path, key string) (string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		prefix := key + "="
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix)), true
		}
	}
	return "", false
}
