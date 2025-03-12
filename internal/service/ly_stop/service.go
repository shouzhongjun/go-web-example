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

type Price struct {
	Id       string `json:"id"`
	ItemCode string `json:"itemCode"`
	ItemName string `json:"itemName"`
	ItemDesc string `json:"itemDesc"`
	Price    string `json:"price"`
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

func (s *Service) GetPrice() []Price {

	return []Price{
		{
			Id:       "1",
			ItemCode: "20140526",
			ItemName: "血液常规检查--血液常规检查",
			ItemDesc: "每次",
			Price:    "23.6",
		},
		{
			Id:       "2",
			ItemCode: "20140526",
			ItemName: "血液常规检查",
			ItemDesc: "每次",
			Price:    "23.6",
		},
		{
			Id:       "3",
			ItemCode: "20140526",
			ItemName: "血液常规检查",
			ItemDesc: "每次",
			Price:    "23.6",
		},
		{
			Id:       "4",
			ItemCode: "20140526",
			ItemName: "血液常规检查",
			ItemDesc: "每次",
			Price:    "23.6",
		},
		{
			Id:       "5",
			ItemCode: "20140526",
			ItemName: "血液常规检查",
			ItemDesc: "每次",
			Price:    "23.6",
		},
		{
			Id:       "6",
			ItemCode: "20140526",
			ItemName: "血液常规检查",
			ItemDesc: "每次",
			Price:    "23.6",
		},
		{
			Id:       "7",
			ItemCode: "20140526",
			ItemName: "血液常规检查",
			ItemDesc: "每次",
			Price:    "23.6",
		},
		{
			Id:       "8",
			ItemCode: "20140526",
			ItemName: "血液常规检查",
			ItemDesc: "每次",
			Price:    "23.6",
		},
		{
			Id:       "9",
			ItemCode: "20140526",
			ItemName: "血液常规检查",
			ItemDesc: "每次",
			Price:    "23.6",
		},
		{
			Id:       "10",
			ItemCode: "20140526",
			ItemName: "血液常规检查",
			ItemDesc: "每次",
			Price:    "23.6",
		},
	}

}
