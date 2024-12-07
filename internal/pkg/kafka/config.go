package kafka

type KafkaConfig struct {
	Brokers []string
}

func ProvideKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		Brokers: []string{"localhost:9092"},
	}
}
