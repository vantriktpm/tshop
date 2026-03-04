package kafka

import (
	"context"

	"github.com/IBM/sarama"
)

// ConsumerGroup wraps sarama ConsumerGroup for graceful shutdown.
type ConsumerGroup struct {
	client sarama.ConsumerGroup
	group  string
}

func NewConsumerGroup(brokers []string, groupID string, cfg *sarama.Config) (*ConsumerGroup, error) {
	if cfg == nil {
		cfg = sarama.NewConfig()
		cfg.Version = sarama.V2_8_0_0
		cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
		cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	}
	client, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		return nil, err
	}
	return &ConsumerGroup{client: client, group: groupID}, nil
}

// Consume runs the handler for each topic. Cancelling ctx stops consuming gracefully.
func (g *ConsumerGroup) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	done := make(chan error, 1)
	go func() {
		for {
			if err := g.client.Consume(ctx, topics, handler); err != nil {
				done <- err
				return
			}
			if ctx.Err() != nil {
				done <- ctx.Err()
				return
			}
		}
	}()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (g *ConsumerGroup) Close() error {
	return g.client.Close()
}
