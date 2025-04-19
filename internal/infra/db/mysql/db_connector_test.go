package mysql

import (
	"context"
	"go.uber.org/zap"
	"goWebExample/internal/configs"
	"testing"
)

func TestConnect(t *testing.T) {
	tests := []struct {
		name         string
		config       *configs.Database
		mockLogger   *zap.Logger
		expectedErr  bool
		expectedLogs string
	}{
		{
			name: "Valid configuration should connect successfully",
			config: &configs.Database{
				Host:         "localhost",
				Port:         3306,
				UserName:     "root",
				Password:     "root",
				DBName:       "dev",
				MaxOpenConns: 10,
				MaxIdleConns: 5,
			},
			mockLogger:   zap.NewNop(),
			expectedErr:  false,
			expectedLogs: "正在连接MySQL数据库",
		},
		{
			name:         "Nil configuration should fail",
			config:       nil,
			mockLogger:   zap.NewNop(),
			expectedErr:  true,
			expectedLogs: "",
		},
		{
			name: "Invalid DSN should fail to connect",
			config: &configs.Database{
				Host:         "invalid-host",
				Port:         3306,
				UserName:     "user",
				Password:     "secret",
				DBName:       "testdb",
				MaxOpenConns: 10,
				MaxIdleConns: 5,
			},
			mockLogger:   zap.NewNop(),
			expectedErr:  true,
			expectedLogs: "连接MySQL失败",
		},
		{
			name: "Nil logger should fail to log",
			config: &configs.Database{
				Host:         "localhost",
				Port:         3306,
				UserName:     "root",
				Password:     "root",
				DBName:       "dev",
				MaxOpenConns: 10,
				MaxIdleConns: 5,
			},
			mockLogger:   nil,
			expectedErr:  true,
			expectedLogs: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			connector := NewDBConnector(test.config, test.mockLogger)

			// 检查 connector 是否为 nil
			if connector == nil {
				// 如果 connector 为 nil，且我们期望错误，则测试通过
				if test.expectedErr {
					return
				}
				// 如果我们期望成功，但 connector 为 nil，测试失败
				t.Errorf("expected connector to not be nil, but it was nil")
				return
			}

			// 现在可以安全调用 Connect
			err := connector.Connect(context.Background())
			if (err != nil) != test.expectedErr {
				t.Errorf("expected error: %v, got: %v", test.expectedErr, err != nil)
			}
		})
	}
}

func TestNewDBConnector(t *testing.T) {
	tests := []struct {
		name      string
		config    *configs.Database
		logger    *zap.Logger
		expectNil bool
	}{
		{
			name: "Valid configuration",
			config: &configs.Database{
				Host:     "localhost",
				Port:     3306,
				UserName: "root",
				Password: "root",
				DBName:   "dev",
			},
			logger:    zap.NewNop(),
			expectNil: false,
		},
		{
			name:      "Nil configuration",
			config:    nil,
			logger:    zap.NewNop(),
			expectNil: false,
		},
		{
			name: "Nil logger",
			config: &configs.Database{
				Host:     "localhost",
				Port:     3306,
				UserName: "root",
				Password: "root",
				DBName:   "dev",
			},
			logger:    nil,
			expectNil: true,
		},
		{
			name: "Empty database config",
			config: &configs.Database{
				Host:     "",
				Port:     0,
				UserName: "",
				Password: "",
				DBName:   "",
			},
			logger:    zap.NewNop(),
			expectNil: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			connector := NewDBConnector(test.config, test.logger)
			if (connector == nil) != test.expectNil {
				t.Errorf("expected nil: %v, got: %v", test.expectNil, connector == nil)
			}
		})
	}
}
