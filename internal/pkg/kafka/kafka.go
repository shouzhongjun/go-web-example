package kafka

import (
	"github.com/segmentio/kafka-go"
)

func ProvideKafkaWriter(config *KafkaConfig) *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:  config.Brokers,
		Topic:    "example-topic",
		Balancer: &kafka.LeastBytes{},
	})
}
