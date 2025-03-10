package discovery

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"

	"goWebExample/internal/configs"
	"goWebExample/internal/infrastructure/connector"
)

// EtcdConnector ETCD连接器
type EtcdConnector struct {
	connector.Connector
	config          *configs.Etcd
	client          *clientv3.Client
	leaseID         clientv3.LeaseID
	cancelKeepAlive context.CancelFunc
}

// NewEtcdConnector 创建ETCD连接器
func NewEtcdConnector(config *configs.Etcd, logger *zap.Logger) *EtcdConnector {
	return &EtcdConnector{
		Connector: *connector.NewConnector("etcd", logger),
		config:    config,
	}
}

// Connect 连接ETCD
func (c *EtcdConnector) Connect(ctx context.Context) error {
	if c.IsConnected() {
		return nil
	}

	c.Logger().Info("正在连接ETCD",
		zap.String("addr", c.config.EtcdAddr()))

	clientConfig := clientv3.Config{
		Endpoints:   []string{c.config.EtcdAddr()},
		DialTimeout: c.config.DialTimeout(),
	}

	if c.config.Username != "" && c.config.Password != "" {
		clientConfig.Username = c.config.Username
		clientConfig.Password = c.config.Password
	}

	client, err := clientv3.New(clientConfig)
	if err != nil {
		return fmt.Errorf("ETCD连接失败: %w", err)
	}

	// 验证连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Status(ctx, c.config.EtcdAddr()); err != nil {
		return fmt.Errorf("ETCD连接验证失败: %w", err)
	}

	c.client = client
	c.SetConnected(true)
	c.Logger().Info("ETCD连接成功")

	return nil
}

// Disconnect 断开ETCD连接
func (c *EtcdConnector) Disconnect(ctx context.Context) error {
	if !c.IsConnected() {
		return nil
	}

	if err := c.client.Close(); err != nil {
		return fmt.Errorf("关闭ETCD连接失败: %w", err)
	}

	c.client = nil
	c.SetConnected(false)
	c.Logger().Info("ETCD连接已关闭")

	return nil
}

// GetClient 获取ETCD客户端
func (c *EtcdConnector) GetClient() *clientv3.Client {
	return c.client
}

// HealthCheck 健康检查
func (c *EtcdConnector) HealthCheck(ctx context.Context) (bool, error) {
	if !c.IsConnected() || c.client == nil {
		return false, fmt.Errorf("ETCD未连接")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.client.Status(ctx, c.config.EtcdAddr())
	return err == nil, err
}

// RegisterService 注册服务
func (c *EtcdConnector) RegisterService(ctx context.Context, serviceName, value string) error {
	if !c.IsConnected() || c.client == nil {
		return fmt.Errorf("ETCD未连接")
	}

	// 如果已存在租约，先清理
	if c.leaseID != 0 {
		if c.cancelKeepAlive != nil {
			c.cancelKeepAlive()
		}
		if _, err := c.client.Revoke(ctx, c.leaseID); err != nil {
			c.Logger().Warn("清理旧租约失败", zap.Error(err))
		}
		c.leaseID = 0
	}

	c.Logger().Info("开始注册服务",
		zap.String("serviceName", serviceName),
		zap.String("value", value),
		zap.Int64("leaseTTL", c.config.GetLeaseTTL()))

	// 创建租约，设置TTL为配置值的2倍，确保有足够的续租时间
	lease := clientv3.NewLease(c.client)
	leaseCtx, leaseCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer leaseCancel()

	ttl := c.config.GetLeaseTTL() * 2 // 将TTL设置为配置值的2倍
	leaseResp, err := lease.Grant(leaseCtx, ttl)
	if err != nil {
		c.Logger().Error("创建租约失败",
			zap.Error(err),
			zap.Int64("leaseTTL", ttl))
		return fmt.Errorf("创建租约失败: %w", err)
	}

	c.Logger().Info("租约创建成功",
		zap.Int64("leaseID", int64(leaseResp.ID)),
		zap.Int64("TTL", leaseResp.TTL))

	// 保存租约ID
	c.leaseID = leaseResp.ID

	// 设置服务键值
	key := fmt.Sprintf("/services/%s", serviceName)
	putCtx, putCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer putCancel()

	_, err = c.client.Put(putCtx, key, value, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		c.Logger().Error("注册服务失败",
			zap.Error(err),
			zap.String("key", key),
			zap.String("value", value))
		return fmt.Errorf("注册服务失败: %w", err)
	}

	// 创建一个独立的 context 用于租约续租
	keepAliveCtx, cancel := context.WithCancel(context.Background())
	c.cancelKeepAlive = cancel

	// 自动续租
	keepAliveChan, err := lease.KeepAlive(keepAliveCtx, leaseResp.ID)
	if err != nil {
		cancel()
		c.Logger().Error("设置自动续租失败",
			zap.Error(err),
			zap.Int64("leaseID", int64(leaseResp.ID)))
		return fmt.Errorf("设置自动续租失败: %w", err)
	}

	// 启动续租监控
	go func() {
		defer cancel()
		failCount := 0
		ticker := time.NewTicker(time.Second) // 每秒检查一次租约状态
		defer ticker.Stop()

		checkLease := func() bool {
			timeToLive, err := lease.TimeToLive(context.Background(), c.leaseID)
			if err != nil {
				c.Logger().Error("获取租约状态失败",
					zap.Error(err),
					zap.Int64("leaseID", int64(c.leaseID)))
				return false
			}

			if timeToLive.TTL <= ttl/2 {
				c.Logger().Warn("租约TTL过低，执行立即续租",
					zap.Int64("leaseID", int64(c.leaseID)),
					zap.Int64("currentTTL", timeToLive.TTL),
					zap.Int64("expectedTTL", ttl))

				// 执行立即续租
				if _, err := lease.KeepAliveOnce(context.Background(), c.leaseID); err != nil {
					c.Logger().Error("立即续租失败",
						zap.Error(err),
						zap.Int64("leaseID", int64(c.leaseID)))
					return false
				}
			}
			return true
		}

		for {
			select {
			case resp := <-keepAliveChan:
				if resp == nil {
					failCount++
					c.Logger().Error("租约续租失败",
						zap.Int64("leaseID", int64(c.leaseID)),
						zap.Int("failCount", failCount))

					if !checkLease() {
						// 尝试重新注册
						if err := c.RegisterService(context.Background(), serviceName, value); err != nil {
							c.Logger().Error("服务重新注册失败", zap.Error(err))
						}
						return
					}
					continue
				}

				// 重置失败计数
				failCount = 0
				c.Logger().Debug("租约续租成功",
					zap.Int64("leaseID", int64(resp.ID)),
					zap.Int64("TTL", resp.TTL))

			case <-ticker.C:
				if !checkLease() {
					// 尝试重新注册
					if err := c.RegisterService(context.Background(), serviceName, value); err != nil {
						c.Logger().Error("服务重新注册失败", zap.Error(err))
					}
					return
				}

			case <-keepAliveCtx.Done():
				c.Logger().Info("续租context已取消",
					zap.Int64("leaseID", int64(c.leaseID)))
				// 不返回，继续尝试重新注册
				if err := c.RegisterService(context.Background(), serviceName, value); err != nil {
					c.Logger().Error("服务重新注册失败", zap.Error(err))
				}
			}
		}
	}()

	c.Logger().Info("服务注册成功",
		zap.String("key", key),
		zap.String("value", value),
		zap.Int64("leaseID", int64(leaseResp.ID)))

	return nil
}

// DeregisterService 注销服务
func (c *EtcdConnector) DeregisterService(ctx context.Context, serviceName string) error {
	if !c.IsConnected() || c.client == nil {
		return fmt.Errorf("ETCD未连接")
	}

	// 删除服务键值
	key := fmt.Sprintf("/services/%s", serviceName)
	_, err := c.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("删除服务键值失败: %w", err)
	}

	c.Logger().Info("服务注销成功", zap.String("key", key))
	return nil
}
