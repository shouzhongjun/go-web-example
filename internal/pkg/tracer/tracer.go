package tracer

import (
	"context"
	"errors"
	"fmt"
	"goWebExample/internal/configs"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"
)

// InitTracer 初始化链路追踪
func InitTracer(cfg *configs.AllConfig, logger *zap.Logger) (*sdktrace.TracerProvider, error) {
	if !cfg.Trace.Enable {
		logger.Info("追踪器未启用")
		return nil, nil
	}

	logger.Info("追踪器已启用")
	logger.Info("正在初始化追踪器")

	if cfg.Trace == nil {
		return nil, fmt.Errorf("tracer config is nil")
	}

	exp, err := createOTLPExporter(cfg.Trace)
	if err != nil {
		return nil, err
	}

	res := createResource(cfg.Trace)
	tp := createTracerProvider(exp, cfg.Trace, res)

	otel.SetTracerProvider(tp)
	return tp, nil
}

// createOTLPExporter 初始化并返回一个具有给定跟踪配置的OTLP跟踪导出器或错误。
func createOTLPExporter(traceCfg *configs.Trace) (*otlptrace.Exporter, error) {
	// 初始化 OTLP HTTP 客户端，配置 Endpoint、Insecure 模式、Timeout 和 Retry 策略
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(traceCfg.Endpoint),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithTimeout(traceCfg.GetClientTimeout()),
		otlptracehttp.WithRetry(otlptracehttp.RetryConfig{
			Enabled:         true,
			InitialInterval: traceCfg.GetRetryInitial(),
			MaxInterval:     traceCfg.GetRetryMax(),
			MaxElapsedTime:  traceCfg.GetRetryElapsed(),
		}),
	)

	// 使用配置好的客户端创建 OTLP Trace Exporter
	exp, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return nil, fmt.Errorf("创建 OTLP exporter 失败: %w", err)
	}
	return exp, nil
}

// createResource 创建一个基于提供的跟踪配置的OpenTelemetry资源，并带有属性。
func createResource(traceCfg *configs.Trace) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(traceCfg.ServiceName),
		semconv.ServiceVersion(traceCfg.ServiceVersion),
		semconv.DeploymentEnvironment(traceCfg.Environment),
		semconv.HostName(traceCfg.ServiceName),
		semconv.TelemetrySDKName("opentelemetry"),
		semconv.TelemetrySDKLanguageGo,
	)
}

// createTracerProvider 初始化并返回一个配置了批处理、采样和资源详细信息的TracerProvider。
func createTracerProvider(exp *otlptrace.Exporter, traceCfg *configs.Trace, res *resource.Resource) *sdktrace.TracerProvider {
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp,
			sdktrace.WithBatchTimeout(traceCfg.GetBatchTimeout()),
			sdktrace.WithMaxExportBatchSize(traceCfg.GetMaxBatchSize()),
			sdktrace.WithMaxQueueSize(traceCfg.GetMaxQueueSize()),
		),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(traceCfg.SamplingRatio)),
		sdktrace.WithResource(res),
	)
}

// ShutdownConfig 包含追踪器关闭的配置选项
type ShutdownConfig struct {
	Timeout  time.Duration
	Logger   *zap.Logger
	Provider *sdktrace.TracerProvider
}

// DefaultTimeout 默认关闭超时时间
const DefaultTimeout = 5 * time.Second

// DefaultShutdownConfig 返回默认的关闭配置
func DefaultShutdownConfig(tp *sdktrace.TracerProvider, logger *zap.Logger) *ShutdownConfig {
	return &ShutdownConfig{
		Timeout:  DefaultTimeout,
		Logger:   logger,
		Provider: tp,
	}
}

// Shutdown 优雅关闭追踪器
func Shutdown(ctx context.Context, cfg *ShutdownConfig) error {
	if cfg == nil {
		return errors.New("关闭配置不能为空")
	}

	if cfg.Provider == nil {
		return nil
	}

	cfg.Logger.Info("正在关闭追踪器")

	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	if err := cfg.Provider.Shutdown(shutdownCtx); err != nil {
		cfg.Logger.Error("关闭追踪器失败", zap.Error(err))
		return errors.New("关闭追踪器失败: " + err.Error())
	}

	return nil
}
