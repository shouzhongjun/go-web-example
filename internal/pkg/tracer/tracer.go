package tracer

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"
)

// Config 链路追踪配置
type Config struct {
	ServiceName    string  // 服务名称
	ServiceVersion string  // 服务版本
	Environment    string  // 环境（dev/prod等）
	Endpoint       string  // OTLP endpoint URL
	SamplingRatio  float64 // 采样率 0.0-1.0
}

// InitTracer 初始化链路追踪
func InitTracer(cfg *Config, logger *zap.Logger) (*sdktrace.TracerProvider, error) {
	logger.Info("正在初始化追踪器")
	if cfg == nil {
		return nil, fmt.Errorf("tracer config is nil")
	}

	// 创建 OTLP exporter
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(cfg.Endpoint),
		otlptracehttp.WithInsecure(),
	)

	exp, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return nil, fmt.Errorf("创建 OTLP exporter 失败: %w", err)
	}

	// 创建资源属性
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion(cfg.ServiceVersion),
		semconv.DeploymentEnvironment(cfg.Environment),
	)

	// 创建 TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(100),
		),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SamplingRatio)),
		sdktrace.WithResource(res),
	)

	// 设置全局 TracerProvider
	otel.SetTracerProvider(tp)
	return tp, nil
}

// Shutdown 优雅关闭追踪器
func Shutdown(ctx context.Context, tp *sdktrace.TracerProvider, logger *zap.Logger) error {
	logger.Info("正在关闭追踪器")
	if tp == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := tp.Shutdown(ctx); err != nil {
		logger.Error("关闭追踪器失败", zap.Error(err))
		return fmt.Errorf("关闭追踪器失败: %w", err)
	}
	return nil
}
