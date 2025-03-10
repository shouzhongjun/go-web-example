package etcd

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"

	"goWebExample/internal/configs"
	"goWebExample/pkg/infrastructure/connector"
)

// EtcdConnector ETCD连接器实现
type EtcdConnector struct {
	*connector.BaseConnector
	config   *configs.Etcd
	client   *clientv3.Client
	leaseID  clientv3.LeaseID
	services map[string]string // 服务名称到服务键的映射
}

// NewEtcdConnector 创建ETCD连接器
func NewEtcdConnector(config *configs.Etcd, logger *zap.Logger) *EtcdConnector {
	base := connector.NewBaseConnector("etcd", logger)
	return &EtcdConnector{
		BaseConnector: base,
		config:        config,
		services:      make(map[string]string),
	}
}

// Connect 连接到ETCD
func (c *EtcdConnector) Connect(ctx context.Context) error {
	if c.IsConnected() {
		return nil
	}

	// 检查配置是否有效
	if c.config == nil || c.config.EtcdAddr() == "" {
		c.Logger().Info("ETCD配置为空或地址未设置，跳过连接")
		return fmt.Errorf("ETCD配置无效")
	}

	// 检查是否主动禁用
	if !c.config.Enable {
		c.Logger().Info("ETCD服务已禁用")
		return fmt.Errorf("ETCD服务已禁用")
	}

	c.Logger().Info("正在连接ETCD服务器",
		zap.String("endpoint", c.config.EtcdAddr()),
		zap.Duration("timeout", c.config.DialTimeout()))

	// 增加连接超时时间
	dialTimeout := c.config.DialTimeout()
	if dialTimeout < 10*time.Second {
		dialTimeout = 10 * time.Second // 确保至少有10秒的连接超时
	}

	// 创建ETCD客户端
	client, err := clientv3.New(clientv3.Config{
		Endpoints:            []string{c.config.EtcdAddr()},
		DialTimeout:          dialTimeout,
		Username:             c.config.Username,
		Password:             c.config.Password,
		Logger:               zap.NewNop(),
		AutoSyncInterval:     30 * time.Second,
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: 3 * time.Second,
		PermitWithoutStream:  true,
	})

	if err != nil {
		return fmt.Errorf("创建ETCD客户端失败: %w", err)
	}

	// 检查连接状态
	statusCtx, cancel := context.WithTimeout(ctx, dialTimeout)
	defer cancel()

	_, err = client.Status(statusCtx, c.config.EtcdAddr())
	if err != nil {
		if closeErr := client.Close(); closeErr != nil {
			c.Logger().Error("关闭ETCD客户端失败", zap.Error(closeErr))
		}
		return fmt.Errorf("连接ETCD服务器失败: %w", err)
	}

	c.client = client
	c.SetConnected(true)
	c.SetClient(client)
	c.Logger().Info("成功连接到ETCD服务器",
		zap.String("endpoint", c.config.EtcdAddr()))

	return nil
}

// Disconnect 断开ETCD连接
func (c *EtcdConnector) Disconnect(ctx context.Context) error {
	if !c.IsConnected() || c.client == nil {
		return nil
	}

	// 撤销所有租约
	if c.leaseID != 0 {
		_, err := c.client.Revoke(ctx, c.leaseID)
		if err != nil {
			c.Logger().Warn("撤销租约失败", zap.Error(err))
		}
	}

	// 关闭客户端
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("关闭ETCD客户端失败: %w", err)
	}

	c.SetConnected(false)
	c.SetClient(nil)
	c.client = nil
	c.leaseID = 0
	c.Logger().Info("ETCD连接已关闭")

	return nil
}

// GetTypedClient 获取类型化的ETCD客户端
func (c *EtcdConnector) GetTypedClient() *clientv3.Client {
	return c.client
}

// HealthCheck 检查ETCD健康状态
func (c *EtcdConnector) HealthCheck(ctx context.Context) (bool, error) {
	if !c.IsConnected() || c.client == nil {
		return false, fmt.Errorf("ETCD未连接")
	}

	statusCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.client.Status(statusCtx, c.config.EtcdAddr())
	if err != nil {
		return false, fmt.Errorf("ETCD健康检查失败: %w", err)
	}

	return true, nil
}

// RegisterService 注册服务到ETCD
func (c *EtcdConnector) RegisterService(ctx context.Context, serviceName, serviceValue string) error {
	if !c.IsConnected() || c.client == nil {
		return fmt.Errorf("ETCD未连接")
	}

	// 创建一个租约，TTL为30秒
	ttl := int64(30)
	if c.config.LeaseTTL > 0 {
		ttl = c.config.LeaseTTL
	}

	// 增加超时时间到15秒
	grantCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// 创建租约
	lease, err := c.client.Grant(grantCtx, ttl)
	if err != nil {
		return fmt.Errorf("创建租约失败: %w", err)
	}
	c.leaseID = lease.ID

	// 服务键
	serviceKey := fmt.Sprintf("/services/%s", serviceName)
	c.services[serviceName] = serviceKey

	// 增加Put操作超时时间
	putCtx, putCancel := context.WithTimeout(ctx, 15*time.Second)
	defer putCancel()

	// 将服务信息写入ETCD
	_, err = c.client.Put(putCtx, serviceKey, serviceValue, clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("注册服务失败: %w", err)
	}

	// 创建keepalive通道
	keepAliveCh, err := c.client.KeepAlive(ctx, lease.ID)
	if err != nil {
		return fmt.Errorf("保持租约失败: %w", err)
	}

	// 启动一个goroutine来处理keepalive响应
	go func() {
		for {
			select {
			case resp, ok := <-keepAliveCh:
				if !ok {
					c.Logger().Warn("租约keepalive通道已关闭")
					return
				}
				if resp == nil {
					c.Logger().Warn("收到空的keepalive响应")
					return
				}
				c.Logger().Debug("续约成功",
					zap.Int64("leaseID", int64(resp.ID)),
					zap.Int64("TTL", resp.TTL))
			case <-ctx.Done():
				c.Logger().Warn("服务注册上下文已取消")
				return
			}
		}
	}()

	c.Logger().Info("服务已成功注册到ETCD",
		zap.String("serviceName", serviceName),
		zap.String("serviceKey", serviceKey))

	return nil
}

// DeregisterService 从ETCD注销服务
func (c *EtcdConnector) DeregisterService(ctx context.Context, serviceName string) error {
	if !c.IsConnected() || c.client == nil {
		return fmt.Errorf("ETCD未连接")
	}

	serviceKey, ok := c.services[serviceName]
	if !ok {
		return fmt.Errorf("服务 %s 未注册", serviceName)
	}

	// 删除服务键
	deleteCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.client.Delete(deleteCtx, serviceKey)
	if err != nil {
		return fmt.Errorf("从ETCD注销服务失败: %w", err)
	}

	delete(c.services, serviceName)
	c.Logger().Info("服务已从ETCD注销",
		zap.String("serviceName", serviceName),
		zap.String("serviceKey", serviceKey))

	return nil
}

// GetService 从ETCD获取服务信息
func (c *EtcdConnector) GetService(ctx context.Context, serviceName string) (string, error) {
	if !c.IsConnected() || c.client == nil {
		return "", fmt.Errorf("ETCD未连接")
	}

	serviceKey := fmt.Sprintf("/services/%s", serviceName)

	// 获取服务信息
	getCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.Get(getCtx, serviceKey)
	if err != nil {
		return "", fmt.Errorf("从ETCD获取服务信息失败: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("服务 %s 不存在", serviceName)
	}

	return string(resp.Kvs[0].Value), nil
}

// WatchService 监视服务变化
func (c *EtcdConnector) WatchService(ctx context.Context, serviceName string) (<-chan string, error) {
	if !c.IsConnected() || c.client == nil {
		return nil, fmt.Errorf("ETCD未连接")
	}

	serviceKey := fmt.Sprintf("/services/%s", serviceName)
	watchCh := make(chan string, 10)

	// 启动监视
	go func() {
		defer close(watchCh)

		watchChan := c.client.Watch(ctx, serviceKey)
		for {
			select {
			case resp := <-watchChan:
				for _, event := range resp.Events {
					if event.Type == clientv3.EventTypePut {
						watchCh <- string(event.Kv.Value)
					}
				}
			case <-ctx.Done():
				c.Logger().Info("停止监视服务",
					zap.String("serviceName", serviceName))
				return
			}
		}
	}()

	return watchCh, nil
}
