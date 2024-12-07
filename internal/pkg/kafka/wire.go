package kafka

import "github.com/google/wire"

var KafkaSet = wire.NewSet(
	ProvideKafkaConfig,
	ProvideKafkaWriter,
)
