package kafkainput

import (
	"context"
	"fmt"
	"log/slog"

	"net-commander-server/internal/config"

	"github.com/segmentio/kafka-go"
)

type InputKafka struct {
	logger *slog.Logger
	config *config.KafkaConfig
}

func New(logger *slog.Logger, appCfg *config.KafkaConfig) *InputKafka {
	return &InputKafka{
		logger: logger,
		config: appCfg,
	}
}

func (k *InputKafka) StartConsumer() (chan []byte, error) {
	k.logger.Info("Starting kafka consumer")

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{k.config.Brokers},
		Topic:    k.config.Topic,
		GroupID:  k.config.ConsumerGroup,
		MaxBytes: 1e6, // 1MB

	})

	data := make(chan []byte, 500)
	go func() {
		for {
			m, err := r.ReadMessage(context.Background())
			if err != nil {
				k.logger.Error("Kafka consumer failed", slog.String("error", err.Error()))
				return
			}

			k.logger.Info(fmt.Sprintf("message at offset %d partirion %d: %s = %s", m.Offset, m.Partition, string(m.Key), string(m.Value)))
			data <- m.Value
		}
	}()

	k.logger.Info("Kafka consumer running")
	return data, nil
}
