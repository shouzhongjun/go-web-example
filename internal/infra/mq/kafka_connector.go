package mq

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"goWebExample/internal/configs"
	"goWebExample/internal/infra/connector"
)

// KafkaConnector Kafka连接器
type KafkaConnector struct {
	connector.Connector
	config      *configs.KafkaConfig
	producer    *kafka.Writer
	consumer    *kafka.Reader
	adminClient *kafka.Client
	dialer      *kafka.Dialer
}

// NewKafkaConnector 创建Kafka连接器
func NewKafkaConnector(config *configs.KafkaConfig, logger *zap.Logger) *KafkaConnector {
	return &KafkaConnector{
		Connector: *connector.NewConnector("kafka", logger),
		config:    config,
	}
}

// Connect 连接Kafka
func (c *KafkaConnector) Connect(ctx context.Context) error {
	if c.IsConnected() {
		return nil
	}

	c.Logger().Info("正在连接Kafka",
		zap.Strings("brokers", c.config.KafkaBrokers()),
		zap.String("topic", c.config.Topic))

	// 创建生产者
	c.producer = &kafka.Writer{
		Addr:         kafka.TCP(c.config.KafkaBrokers()...),
		Topic:        c.config.Topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    c.config.BatchSize,
		Async:        c.config.Async,
		RequiredAcks: kafka.RequireAll,
	}

	// 创建消费者
	c.consumer = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        c.config.KafkaBrokers(),
		Topic:          c.config.Topic,
		GroupID:        c.config.GroupID,
		MinBytes:       10e3,        // 10KB
		MaxBytes:       10e6,        // 10MB
		CommitInterval: time.Second, // 自动提交的间隔
	})

	// 创建管理客户端
	c.dialer = &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}
	c.adminClient = &kafka.Client{
		Addr: kafka.TCP(c.config.KafkaBrokers()...),
		Transport: &kafka.Transport{
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return c.dialer.DialContext(ctx, network, address)
			},
		},
	}

	// 验证连接
	if err := c.checkConnection(ctx); err != nil {
		c.closeAll()
		return fmt.Errorf("kafka连接验证失败: %w", err)
	}

	c.SetConnected(true)
	c.Logger().Info("Kafka连接成功")

	return nil
}

// Disconnect 断开Kafka连接
func (c *KafkaConnector) Disconnect(ctx context.Context) error {
	if !c.IsConnected() {
		return nil
	}

	c.closeAll()
	c.SetConnected(false)
	c.Logger().Info("Kafka连接已关闭")

	return nil
}

// closeAll 关闭所有客户端
func (c *KafkaConnector) closeAll() {
	if c.producer != nil {
		if err := c.producer.Close(); err != nil {
			c.Logger().Error("关闭Kafka生产者失败", zap.Error(err))
		}
		c.producer = nil
	}

	if c.consumer != nil {
		if err := c.consumer.Close(); err != nil {
			c.Logger().Error("关闭Kafka消费者失败", zap.Error(err))
		}
		c.consumer = nil
	}

	c.dialer = nil
	c.adminClient = nil
}

// GetProducer 获取生产者
func (c *KafkaConnector) GetProducer() *kafka.Writer {
	return c.producer
}

// GetConsumer 获取消费者
func (c *KafkaConnector) GetConsumer() *kafka.Reader {
	return c.consumer
}

// GetAdminClient 获取管理客户端
func (c *KafkaConnector) GetAdminClient() *kafka.Client {
	return c.adminClient
}

// checkConnection 检查连接状态
func (c *KafkaConnector) checkConnection(ctx context.Context) error {
	// 检查管理客户端
	_, err := c.adminClient.Metadata(ctx, &kafka.MetadataRequest{
		Topics: []string{c.config.Topic},
	})
	if err != nil {
		return fmt.Errorf("无法获取主题元数据: %w", err)
	}

	return nil
}

// HealthCheck 健康检查
func (c *KafkaConnector) HealthCheck(ctx context.Context) (bool, error) {
	if !c.IsConnected() {
		return false, fmt.Errorf("Kafka未连接")
	}

	err := c.checkConnection(ctx)
	return err == nil, err
}

// CreateTopic 创建主题
func (c *KafkaConnector) CreateTopic(ctx context.Context, topic string, partitions int, replicationFactor int) error {
	if !c.IsConnected() {
		return fmt.Errorf("Kafka未连接")
	}

	_, err := c.adminClient.CreateTopics(ctx, &kafka.CreateTopicsRequest{
		Topics: []kafka.TopicConfig{
			{
				Topic:             topic,
				NumPartitions:     partitions,
				ReplicationFactor: replicationFactor,
			},
		},
	})

	if err != nil {
		return fmt.Errorf("创建主题失败: %w", err)
	}

	c.Logger().Info("主题创建成功",
		zap.String("topic", topic),
		zap.Int("partitions", partitions),
		zap.Int("replicationFactor", replicationFactor))

	return nil
}

// DeleteTopic 删除主题
func (c *KafkaConnector) DeleteTopic(ctx context.Context, topic string) error {
	if !c.IsConnected() {
		return fmt.Errorf("Kafka未连接")
	}

	_, err := c.adminClient.DeleteTopics(ctx, &kafka.DeleteTopicsRequest{
		Topics: []string{topic},
	})

	if err != nil {
		return fmt.Errorf("删除主题失败: %w", err)
	}

	c.Logger().Info("主题删除成功", zap.String("topic", topic))
	return nil
}

// SendMessage 发送消息
func (c *KafkaConnector) SendMessage(ctx context.Context, topic string, key string, value []byte) error {
	if !c.IsConnected() || c.producer == nil {
		return fmt.Errorf("Kafka未连接")
	}

	err := c.producer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
	})

	if err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	c.Logger().Debug("消息发送成功",
		zap.String("topic", topic),
		zap.String("key", key),
		zap.Int("valueSize", len(value)))

	return nil
}

// ConsumeMessages 消费消息
func (c *KafkaConnector) ConsumeMessages(ctx context.Context, handler func([]byte) error) error {
	if !c.IsConnected() || c.consumer == nil {
		return fmt.Errorf("Kafka未连接")
	}

	c.Logger().Info("开始消费消息",
		zap.String("topic", c.config.Topic),
		zap.String("groupID", c.config.GroupID))

	for {
		select {
		case <-ctx.Done():
			c.Logger().Info("停止消费消息")
			return nil
		default:
			msg, err := c.consumer.ReadMessage(ctx)
			if err != nil {
				c.Logger().Error("读取消息失败", zap.Error(err))
				continue
			}

			if err := handler(msg.Value); err != nil {
				c.Logger().Error("处理消息失败",
					zap.Error(err),
					zap.String("topic", msg.Topic),
					zap.Int64("offset", msg.Offset))
				continue
			}

			c.Logger().Debug("消息处理成功",
				zap.String("topic", msg.Topic),
				zap.Int64("offset", msg.Offset))
		}
	}
}
