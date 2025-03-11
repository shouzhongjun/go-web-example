package ly_stop

import (
	"goWebExample/internal/service"

	"go.uber.org/zap"
)

const ServiceName = "ly_stop"

// Service 停诊服务
type Service struct {
	logger *zap.Logger
}

// NewService 创建停诊服务
func NewService(logger *zap.Logger) *Service {
	svc := &Service{
		logger: logger,
	}
	// 自动注册到服务注册器
	service.GetRegistry().Register(ServiceName, svc)
	return svc
}

// DataMock 停诊数据结构
type DataMock struct {
	OPERATINGROOMNO string `json:"OPERATING_ROOM_NO"`
	NAME            string `json:"NAME"`
	BEDNO           string `json:"BED_NO"`
	STA             string `json:"STA"`
}

// GetData 获取停诊数据
func (s *Service) GetData() []DataMock {
	return []DataMock{
		{
			OPERATINGROOMNO: "超声科",
			NAME:            "车惠娟",
			BEDNO:           "停诊",
			STA:             "2025-03-11",
		},
		{
			OPERATINGROOMNO: "心电科",
			NAME:            "李红",
			BEDNO:           "停诊",
			STA:             "2025-03-11",
		},
		{
			OPERATINGROOMNO: "普外科",
			NAME:            "施炜",
			BEDNO:           "停诊",
			STA:             "2025-03-11",
		},
		{
			OPERATINGROOMNO: "小儿哮喘(西后街院区)",
			NAME:            "彭家明",
			BEDNO:           "停诊",
			STA:             "2025-03-11",
		},
	}
}
