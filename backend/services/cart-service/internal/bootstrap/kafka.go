package bootstrap

// NewKafka creates a Kafka producer using KAFKA_BROKER from config (.env).
// Cart-service does not use Kafka; returns nil.
func NewKafka() interface{} {
	return nil
}
