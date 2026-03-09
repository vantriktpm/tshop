package bootstrap

// NewPostgres creates a PostgreSQL connection using DB_DSN from config (.env).
// Cart-service does not use PostgreSQL; returns nil.
func NewPostgres() interface{} {
	return nil
}
