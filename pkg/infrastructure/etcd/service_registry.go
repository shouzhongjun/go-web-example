package etcd

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"goWebExample/internal/configs"
)

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Version string `json:"version"`
}

// ServiceRegistry 定义服务注册接口
type ServiceRegistry interface {
	Register(ctx context.Context) error
	Deregister(ctx context.Context) error
}

// etcdRegistry 实现 ServiceRegistry 接口
type etcdRegistry struct {
	connector  *EtcdConnector
	config     *configs.AllConfig
	logger     *zap.Logger
	serviceKey string
}

// NewServiceRegistry 创建服务注册器
func NewServiceRegistry(config *configs.AllConfig, logger *zap.Logger, connector *EtcdConnector) ServiceRegistry {
	// 检查 Etcd 配置是否启用
	if config.Etcd == nil || config.Etcd.EtcdAddr() == "" {
		logger.Info("Etcd配置为空或地址未设置，跳过服务注册")
		return &failedRegistry{logger: logger, reason: "config_missing"}
	}

	// 检查是否主动禁用
	if !config.Etcd.Enable {
		logger.Info("Etcd服务注册功能已禁用")
		return &failedRegistry{logger: logger, reason: "disabled"}
	}

	serviceKey := fmt.Sprintf("/services/%s", config.Server.ServerName)

	registry := &etcdRegistry{
		connector:  connector,
		config:     config,
		logger:     logger,
		serviceKey: serviceKey,
	}
	return registry
}

func (e *etcdRegistry) Register(ctx context.Context) error {
	// 构建服务信息
	serviceValue := fmt.Sprintf(`{"name":"%s","address":"%s","version":"%s"}`,
		e.config.Server.ServerName, e.config.Server.Host, e.config.Server.Version)

	// 使用connector注册服务
	err := e.connector.RegisterService(ctx, e.config.Server.ServerName, serviceValue)
	if err != nil {
		e.logger.Error("注册服务失败",
			zap.String("endpoint", e.config.Etcd.EtcdAddr()),
			zap.String("key", e.serviceKey),
			zap.Error(err))
		return fmt.Errorf("注册服务失败: %w", err)
	}

	e.logger.Info("服务已成功注册到Etcd")
	return nil
}

func (e *etcdRegistry) Deregister(ctx context.Context) error {
	// 使用connector注销服务
	err := e.connector.DeregisterService(ctx, e.config.Server.ServerName)
	if err != nil {
		e.logger.Error("注销服务失败",
			zap.String("endpoint", e.config.Etcd.EtcdAddr()),
			zap.String("key", e.serviceKey),
			zap.Error(err))
		return fmt.Errorf("注销服务失败: %w", err)
	}

	e.logger.Info("服务已从Etcd注销", zap.String("serviceKey", e.serviceKey))
	return nil
}

// failedRegistry 表示连接失败的注册器
type failedRegistry struct {
	err    error
	logger *zap.Logger
	reason string // 原因
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
