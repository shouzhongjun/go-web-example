package tracer

import (
	"log"
	"sync"

	"go.uber.org/zap"
)

// OtelLoggerWriter 是一个自定义编写器，将标准日志输出重定向到 zap 日志记录器
type OtelLoggerWriter struct {
	logger *zap.Logger
	mu     sync.Mutex
	buf    []byte
}

// NewOtelLoggerWriter 创建一个新的 OtelLoggerWriter
func NewOtelLoggerWriter(logger *zap.Logger) *OtelLoggerWriter {
	return &OtelLoggerWriter{
		logger: logger,
		buf:    make([]byte, 0, 1024),
	}
}

// Write implements io.Writer interface
func (w *OtelLoggerWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Copy the bytes to avoid modifying the original slice
	line := make([]byte, len(p))
	copy(line, p)

	// Remove trailing newline if present
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}

	// Parse the message to determine the appropriate log level
	msg := string(line)

	// Skip empty messages
	if msg == "" {
		return len(p), nil
	}

	// Log all messages as errors to ensure they're captured
	// This is especially important for trace export errors
	w.logger.Error(msg)

	// Return 0 bytes written to suppress the original log message
	// This prevents the message from being displayed in the original format
	return 0, nil
}

// SetupOtelLogger configures the standard logger to use our custom writer
func SetupOtelLogger(logger *zap.Logger) func() {
	// Create a custom writer that redirects to zap
	writer := NewOtelLoggerWriter(logger)

	// Save the original writer
	originalOutput := log.Writer()
	originalFlags := log.Flags()

	// Set the standard logger to use only our custom writer
	log.SetOutput(writer)
	log.SetFlags(0) // Remove timestamp and other prefixes

	// Return a cleanup function
	return func() {
		log.SetOutput(originalOutput)
		log.SetFlags(originalFlags)
	}
}
