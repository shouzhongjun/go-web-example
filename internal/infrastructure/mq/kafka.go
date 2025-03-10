package kafka

import (
	"context"
	"time"

	"goWebExample/internal/configs"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// KafkaClient 封装Kafka的读写操作
type KafkaClient struct {
	Writer *kafka.Writer
	Reader *kafka.Reader
	Logger *zap.Logger
	Config *configs.KafkaConfig
}

// NewKafka 创建Kafka客户端
func NewKafka(config *configs.AllConfig, logger *zap.Logger) *KafkaClient {
	return &KafkaClient{
		Writer: createWriter(&config.Kafka),
		Reader: createReader(&config.Kafka),
		Logger: logger,
		Config: &config.Kafka,
	}
}

// createWriter 创建Kafka写入器
func createWriter(config *configs.KafkaConfig) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(config.KafkaBrokers()...),
		Topic:        config.Topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    config.BatchSize,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
		Async:        config.Async,
		Compression:  kafka.Snappy,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

// createReader 创建Kafka读取器
func createReader(config *configs.KafkaConfig) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        config.KafkaBrokers(),
		Topic:          config.Topic,
		GroupID:        config.GroupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        1 * time.Second,
		StartOffset:    kafka.LastOffset,
		CommitInterval: time.Second,
	})
}

// WriteMessage 发送消息到Kafka
func (k *KafkaClient) WriteMessage(ctx context.Context, key, value []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}

	err := k.Writer.WriteMessages(ctx, msg)
	if err != nil {
		k.Logger.Error("发送Kafka消息失败",
			zap.Error(err),
			zap.ByteString("key", key))
		return err
	}

	return nil
}

// ReadMessage 从Kafka读取消息
func (k *KafkaClient) ReadMessage(ctx context.Context) (kafka.Message, error) {
	msg, err := k.Reader.ReadMessage(ctx)
	if err != nil {
		k.Logger.Error("读取Kafka消息失败", zap.Error(err))
		return kafka.Message{}, err
	}

	return msg, nil
}

// Close 关闭Kafka连接
func (k *KafkaClient) Close() error {
	if err := k.Writer.Close(); err != nil {
		k.Logger.Error("关闭Kafka写入器失败", zap.Error(err))
		return err
	}

	if err := k.Reader.Close(); err != nil {
		k.Logger.Error("关闭Kafka读取器失败", zap.Error(err))
		return err
	}

	return nil
}
