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

// NewServiceRegistry 创建服务注册器，失败时返回空实现
func NewServiceRegistry(config *configs.AllConfig, logger *zap.Logger) ServiceRegistry {
	if config.Etcd == nil || config.Etcd.GetAddr() == "" {
		logger.Info("Etcd配置为空或地址未设置，跳过服务注册")
		return &failedRegistry{logger: logger}
	}

	// 创建一个禁用内部日志的logger
	errorOnlyLogger := zap.NewNop()

	// 调整客户端配置
	client, err := clientv3.New(clientv3.Config{
		Endpoints:            []string{config.Etcd.GetAddr()},
		DialTimeout:          config.Etcd.DialTimeout(),
		Username:             config.Etcd.Username,
		Password:             config.Etcd.Password,
		Logger:               errorOnlyLogger,
		AutoSyncInterval:     30 * time.Second, // 定期同步endpoints
		DialKeepAliveTime:    10 * time.Second, // keepalive探活间隔
		DialKeepAliveTimeout: 3 * time.Second,  // keepalive超时时间
		PermitWithoutStream:  true,             // 允许无流连接
	})

	if err != nil {
		// 使用应用自己的logger记录错误
		logger.Error("连接etcd服务器失败，将使用空实现",
			zap.String("endpoint", config.Etcd.GetAddr()),
			zap.Error(err))
		return &failedRegistry{err: err, logger: logger}
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), config.Etcd.DialTimeout())
	defer cancel()

	_, err = client.Status(ctx, config.Etcd.GetAddr())
	if err != nil {
		client.Close()
		logger.Error("连接etcd服务器失败，将使用空实现",
			zap.String("endpoint", config.Etcd.GetAddr()),
			zap.Error(err))
		return &failedRegistry{err: err, logger: logger}
	}

	logger.Info("成功连接到etcd服务器",
		zap.String("endpoint", config.Etcd.GetAddr()))

	// 构建服务键
	serviceKey := fmt.Sprintf("/services/%s", config.Server.ServerName)

	return &etcdRegistry{
		client:     client,
		config:     config,
		logger:     logger,
		serviceKey: serviceKey,
	}
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

	// 关闭etcd客户端连接
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
}

// Register 始终返回错误
func (f *failedRegistry) Register(ctx context.Context) error {
	f.logger.Error("无法注册服务，ETCD连接失败", zap.Error(f.err))
	return f.err
}

// Deregister 空实现
func (f *failedRegistry) Deregister(ctx context.Context) error {
	return nil
}

// etcdZapLogger 适配zap日志到etcd客户端的日志接口
type etcdZapLogger struct {
	lg        *zap.Logger
	verbosity int
}

func newEtcdZapLogger(lg *zap.Logger) *etcdZapLogger {
	return &etcdZapLogger{
		lg:        lg,
		verbosity: 0, // 默认最低详细度
	}
}

// 添加SetVerbosity方法
func (l *etcdZapLogger) SetVerbosity(v int) {
	l.verbosity = v
}

// 修改V方法实现
func (l *etcdZapLogger) V(level int) bool {
	return level <= l.verbosity
}

func (l *etcdZapLogger) Info(args ...interface{}) {
	l.lg.Sugar().Info(args...)
}

func (l *etcdZapLogger) Infoln(args ...interface{}) {
	l.lg.Sugar().Info(args...)
}

func (l *etcdZapLogger) Infof(format string, args ...interface{}) {
	l.lg.Sugar().Infof(format, args...)
}

func (l *etcdZapLogger) Warning(args ...interface{}) {
	l.lg.Sugar().Warn(args...)
}

func (l *etcdZapLogger) Warningln(args ...interface{}) {
	l.lg.Sugar().Warn(args...)
}

func (l *etcdZapLogger) Warningf(format string, args ...interface{}) {
	l.lg.Sugar().Warnf(format, args...)
}

func (l *etcdZapLogger) Error(args ...interface{}) {
	l.lg.Sugar().Error(args...)
}

func (l *etcdZapLogger) Errorln(args ...interface{}) {
	l.lg.Sugar().Error(args...)
}

func (l *etcdZapLogger) Errorf(format string, args ...interface{}) {
	l.lg.Sugar().Errorf(format, args...)
}

func (l *etcdZapLogger) Fatal(args ...interface{}) {
	l.lg.Sugar().Fatal(args...)
}

func (l *etcdZapLogger) Fatalln(args ...interface{}) {
	l.lg.Sugar().Fatal(args...)
}

func (l *etcdZapLogger) Fatalf(format string, args ...interface{}) {
	l.lg.Sugar().Fatalf(format, args...)
}
