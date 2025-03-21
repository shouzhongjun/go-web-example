package stream

import (
	"time"

	"go.uber.org/zap"
)

const ServiceName = "stream"

type StreamService struct {
	logger *zap.Logger
}

func NewStreamService(logger *zap.Logger) *StreamService {
	return &StreamService{
		logger: logger,
	}
}

// GenerateStream 生成流式数据
func (s *StreamService) GenerateStream(messages []string) chan string {
	stream := make(chan string)

	go func() {
		defer func() {
			close(stream)
			s.logger.Info("stream closed")
		}()

		for _, msg := range messages {
			select {
			case stream <- msg:
				s.logger.Debug("sent message", zap.String("message", msg))
				time.Sleep(500 * time.Millisecond)
			case <-time.After(5 * time.Second):
				s.logger.Warn("stream send timeout")
				return
			}
		}
	}()

	return stream
}
