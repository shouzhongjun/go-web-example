package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"

	"goWebExample/internal/configs"
	"goWebExample/internal/infra/connector"
)

// MongoDBConnector MongoDB连接器
type MongoDBConnector struct {
	connector.Connector
	config *configs.MongoDB
	client *mongo.Client
}

// NewMongoDBConnector 创建MongoDB连接器
func NewMongoDBConnector(config *configs.MongoDB, logger *zap.Logger) *MongoDBConnector {
	return &MongoDBConnector{
		Connector: *connector.NewConnector("mongodb", logger),
		config:    config,
	}
}

// Connect 连接MongoDB
func (c *MongoDBConnector) Connect(ctx context.Context) error {
	if c.IsConnected() {
		return nil
	}

	c.Logger().Info("正在连接MongoDB",
		zap.String("uri", c.config.URI))

	clientOptions := options.Client().
		ApplyURI(c.config.URI).
		SetMaxPoolSize(uint64(c.config.MaxPoolSize)).
		SetMinPoolSize(uint64(c.config.MinPoolSize)).
		SetMaxConnIdleTime(time.Duration(c.config.MaxConnIdleTime) * time.Second)

	if c.config.Username != "" && c.config.Password != "" {
		credential := options.Credential{
			Username: c.config.Username,
			Password: c.config.Password,
		}
		clientOptions.SetAuth(credential)
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("MongoDB连接失败: %w", err)
	}

	// 验证连接
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("MongoDB连接验证失败: %w", err)
	}

	c.client = client
	c.SetConnected(true)
	c.Logger().Info("MongoDB连接成功",
		zap.Uint64("最大连接数", *clientOptions.MaxPoolSize),
		zap.Uint64("最小连接数", *clientOptions.MinPoolSize))

	return nil
}

// Disconnect 断开MongoDB连接
func (c *MongoDBConnector) Disconnect(ctx context.Context) error {
	if !c.IsConnected() {
		return nil
	}

	if err := c.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("关闭MongoDB连接失败: %w", err)
	}

	c.client = nil
	c.SetConnected(false)
	c.Logger().Info("MongoDB连接已关闭")

	return nil
}

// GetClient 获取MongoDB客户端
func (c *MongoDBConnector) GetClient() *mongo.Client {
	return c.client
}

// GetDatabase 获取指定数据库
func (c *MongoDBConnector) GetDatabase(name string) *mongo.Database {
	return c.client.Database(name)
}

// GetCollection 获取指定集合
func (c *MongoDBConnector) GetCollection(database, collection string) *mongo.Collection {
	return c.client.Database(database).Collection(collection)
}

// HealthCheck 健康检查
func (c *MongoDBConnector) HealthCheck(ctx context.Context) (bool, error) {
	if !c.IsConnected() || c.client == nil {
		return false, fmt.Errorf("MongoDB未连接")
	}

	err := c.client.Ping(ctx, readpref.Primary())
	return err == nil, err
}
