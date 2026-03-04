package kafka

import (
	"context"

	"github.com/IBM/sarama"
)

type Publisher struct {
	producer sarama.SyncProducer
}

func NewPublisher(brokers []string) (*Publisher, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return &Publisher{producer: producer}, nil
}

func (p *Publisher) Publish(ctx context.Context, topic, key string, payload []byte) error {
	_, _, err := p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(payload),
	})
	return err
}

func (p *Publisher) Close() error {
	return p.producer.Close()
}
