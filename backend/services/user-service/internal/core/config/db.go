package config

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/tshop/backend/pkg/dbutil"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewPostgresDB creates a *gorm.DB connection for user-service using config or env.
// It ensures the target database and schema exist before returning the connection.
func NewPostgresDB() (*gorm.DB, error) {
	dsn := resolveDSN()

	// Ensure the database exists (creates it if missing).
	if err := dbutil.EnsureDatabase(dsn, "user_db"); err != nil {
		return nil, fmt.Errorf("ensure database: %w", err)
	}

	db, err := openWithPGX(dsn)
	if err != nil {
		return nil, err
	}

	// Ensure the "service" schema exists (user-service uses service.users).
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	if err := dbutil.EnsureSchema(sqlDB, "service"); err != nil {
		return nil, fmt.Errorf("ensure schema: %w", err)
	}

	// Ensure the users table exists inside the "service" schema.
	if err := ensureUsersTable(sqlDB); err != nil {
		return nil, fmt.Errorf("ensure users table: %w", err)
	}

	return db, nil
}

// ensureUsersTable creates service.users if it does not already exist.
func ensureUsersTable(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS service.users (
    id                  UUID PRIMARY KEY,
    user_name           VARCHAR(255),
    full_name           VARCHAR(255),
    phone               VARCHAR(255),
    password_hash       TEXT,
    salt                VARCHAR(255),
    status              SMALLINT,
    is_verified         BOOLEAN,
    user_id             VARCHAR(255),
    provider            VARCHAR(255),
    provider_user_id    VARCHAR(255),
    access_token        TEXT,
    password_changed_at TIMESTAMPTZ,
    refresh_token       TEXT,
    token_version       INTEGER DEFAULT 1,
    created_by          VARCHAR(50),
    updated_by          VARCHAR(50),
    created_at          TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ,
    expires_at          TIMESTAMPTZ
)`)
	return err
}

// openWithPGX opens a *sql.DB using pgx stdlib and wraps it with GORM.
func openWithPGX(dsn string) (*gorm.DB, error) {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return gorm.Open(gormpostgres.New(gormpostgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
}

// resolveDSN resolves the DSN with the same priority chain as before.
func resolveDSN() string {
	// Priority: USER_SERVICE_CONFIG_PATH file > .env/user.config in CWD >
	// known service paths > executable directory > USER_SERVICE_DB_DSN env > default.
	if fileDSN, ok := loadFromUserConfig("DB_DSN"); ok && fileDSN != "" {
		return fileDSN
	}
	if dsn := os.Getenv("USER_SERVICE_DB_DSN"); dsn != "" {
		return dsn
	}
	return "host=localhost user=postgres password=1 dbname=user_db port=5432 sslmode=disable"
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
