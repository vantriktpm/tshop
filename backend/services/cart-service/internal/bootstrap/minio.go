package bootstrap

// NewMinio creates a MinIO client using MINIO_* from config (.env).
// Cart-service does not use MinIO; returns nil.
func NewMinio() interface{} {
	return nil
}
