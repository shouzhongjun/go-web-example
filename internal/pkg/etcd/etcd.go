package etcd

import (
	"context"
	"fmt"
	"goWebExample/internal/configs"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// ServiceRegistry 定义服务注册接口
type ServiceRegistry interface {
	Register(ctx context.Context) error
	Deregister(ctx context.Context) error
}

// etcdRegistry 实现 ServiceRegistry 接口
type etcdRegistry struct {
	client     *clientv3.Client
	config     *configs.AllConfig
	logger     *zap.Logger
	serviceKey string
	leaseID    clientv3.LeaseID
}

func (e *etcdRegistry) Register(ctx context.Context) error {
	// 创建一个租约，TTL为30秒
	ttl := int64(30)
	if e.config.Etcd.LeaseTTL > 0 {
		ttl = e.config.Etcd.LeaseTTL
	}

	// 增加超时时间到15秒
	grantCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// 创建租约
	lease, err := e.client.Grant(grantCtx, ttl)
	if err != nil {
		e.logger.Error("创建租约失败",
			zap.String("endpoint", e.config.Etcd.GetAddr()),
			zap.Error(err))
		return fmt.Errorf("创建租约失败: %w", err)
	}
	e.leaseID = lease.ID

	// 修正服务信息构建
	serviceValue := fmt.Sprintf(`{"name":"%s","address":"%s","version":"%s"}`,
		e.config.Server.ServerName)

	// 增加Put操作超时时间
	putCtx, putCancel := context.WithTimeout(ctx, 15*time.Second)
	defer putCancel()

	// 将服务信息写入etcd
	_, err = e.client.Put(putCtx, e.serviceKey, serviceValue, clientv3.WithLease(lease.ID))
	if err != nil {
		e.logger.Error("注册服务失败",
			zap.String("endpoint", e.config.Etcd.GetAddr()),
			zap.String("key", e.serviceKey),
			zap.Error(err))
		return fmt.Errorf("注册服务失败: %w", err)
	}

	// 创建keepalive通道前先检查客户端状态
	if e.client == nil {
		return fmt.Errorf("etcd客户端未初始化")
	}

	keepAliveCh, err := e.client.KeepAlive(ctx, lease.ID)
	if err != nil {
		e.logger.Error("保持租约失败", zap.Error(err))
		return fmt.Errorf("保持租约失败: %w", err)
	}

	// 启动一个goroutine来处理keepalive响应
	go func() {
		for {
			select {
			case resp, ok := <-keepAliveCh:
				if !ok {
					e.logger.Warn("租约keepalive通道已关闭")
					return
				}
				if resp == nil {
					e.logger.Warn("收到空的keepalive响应")
					return
				}
				e.logger.Debug("续约成功",
					zap.Int64("leaseID", int64(resp.ID)),
					zap.Int64("TTL", resp.TTL))
			case <-ctx.Done():
				e.logger.Warn("服务注册上下文已取消")
				return
			}
		}
	}()

	e.logger.Info("服务已成功注册到Etcd")

	return nil
}

// NewServiceRegistry 创建服务注册器，失败时返回空实现
func NewServiceRegistry(config *configs.AllConfig, logger *zap.Logger) ServiceRegistry {
	// 检查 Etcd 配置是否启用
	if config.Etcd == nil || config.Etcd.GetAddr() == "" {
		logger.Info("Etcd配置为空或地址未设置，跳过服务注册")
		return &failedRegistry{logger: logger, reason: "config_missing"}
	}

	// 检查是否主动禁用
	if !config.Etcd.Enable {
		logger.Info("Etcd服务注册功能已禁用")
		return &failedRegistry{logger: logger, reason: "disabled"}
	}

	// 记录尝试连接的信息
	logger.Info("尝试连接etcd服务器",
		zap.String("endpoint", config.Etcd.GetAddr()),
		zap.Duration("timeout", config.Etcd.DialTimeout()))

	// 增加连接超时时间
	dialTimeout := config.Etcd.DialTimeout()
	if dialTimeout < 10*time.Second {
		dialTimeout = 10 * time.Second // 确保至少有10秒的连接超时
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:            []string{config.Etcd.GetAddr()},
		DialTimeout:          dialTimeout,
		Username:             config.Etcd.Username,
		Password:             config.Etcd.Password,
		Logger:               zap.NewNop(),
		AutoSyncInterval:     30 * time.Second,
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: 3 * time.Second,
		PermitWithoutStream:  true,
	})

	if err != nil {
		logger.Error("创建etcd客户端失败，将使用空实现",
			zap.String("endpoint", config.Etcd.GetAddr()),
			zap.Error(err))
		return &failedRegistry{err: err, logger: logger, reason: "connection_failed"}
	}

	// 增加状态检查超时时间
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	// 尝试获取状态前记录日志
	logger.Debug("正在检查etcd服务器状态", zap.String("endpoint", config.Etcd.GetAddr()))

	_, err = client.Status(ctx, config.Etcd.GetAddr())
	if err != nil {
		err := client.Close()
		if err != nil {
			return nil
		}
		logger.Error("连接etcd服务器失败，将使用空实现",
			zap.String("endpoint", config.Etcd.GetAddr()),
			zap.Error(err))

		// 提供更详细的错误信息和建议
		logger.Warn("请确保etcd服务器已启动并且可以访问。您可以使用以下命令检查etcd状态：",
			zap.String("check_command", "curl -L http://"+config.Etcd.GetAddr()+"/health"))

		return &failedRegistry{err: err, logger: logger, reason: "connection_failed"}
	}

	logger.Info("成功连接到etcd服务器",
		zap.String("endpoint", config.Etcd.GetAddr()))

	serviceKey := fmt.Sprintf("/services/%s", config.Server.ServerName)

	registry := &etcdRegistry{
		client:     client,
		config:     config,
		logger:     logger,
		serviceKey: serviceKey,
	}
	return registry
}

// Deregister 从Etcd注销服务
func (e *etcdRegistry) Deregister(ctx context.Context) error {
	if e.leaseID != 0 {
		_, err := e.client.Revoke(ctx, e.leaseID)
		if err != nil {
			return fmt.Errorf("撤销租约失败: %w", err)
		}
		e.logger.Info("服务已从Etcd注销", zap.String("serviceKey", e.serviceKey))
	}

	if e.client != nil {
		if err := e.client.Close(); err != nil {
			e.logger.Warn("关闭etcd客户端失败", zap.Error(err))
		}
	}

	return nil
}

// failedRegistry 表示连接失败的注册器
type failedRegistry struct {
	err    error
	logger *zap.Logger
	reason string // 添加一个原因字段
}

func (f *failedRegistry) Register(ctx context.Context) error {
	// 如果是主动不启用，则只记录信息级别日志
	if f.reason == "disabled" {
		f.logger.Info("ETCD服务注册已禁用，跳过注册")
		return nil
	}

	// 否则记录错误日志
	f.logger.Error("无法注册服务，ETCD连接失败", zap.Error(f.err))
	return f.err
}

func (f *failedRegistry) Deregister(ctx context.Context) error {
	return nil
}
