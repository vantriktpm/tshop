package bootstrap

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// loadEnv loads key=value from .env file into the process environment.
// Tries: cmd/.env, .env, then executable directory.
func loadEnv() error {
	for _, path := range envPaths() {
		if err := loadEnvFile(path); err == nil {
			return nil
		}
	}
	return os.ErrNotExist
}

// loadEnvFile reads a .env file and sets env vars. Returns nil if file was read.
func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		i := strings.Index(line, "=")
		if i <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:i])
		value := strings.TrimSpace(line[i+1:])
		if key != "" {
			_ = os.Setenv(key, value)
		}
	}
	return s.Err()
}

func envPaths() []string {
	base := "."
	if exec, err := os.Executable(); err == nil {
		base = filepath.Dir(exec)
	}
	return []string{
		"cmd/.env",
		".env",
		filepath.Join(base, "cmd", ".env"),
		filepath.Join(base, ".env"),
	}
}
