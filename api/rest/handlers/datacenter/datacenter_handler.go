package datacenter

import (
	"goWebExample/internal/service/datacenter_service"

	"github.com/gin-gonic/gin"
)

type DataCenterHandler struct {
	dataCenterService *datacenter_service.MockDataCenter
}

type Data struct {
	PageNo   int `json:"pageNo"`
	PageSize int `json:"pageSize"`
}

func NewDataCenterHandler(dataCenterService *datacenter_service.MockDataCenter) *DataCenterHandler {
	return &DataCenterHandler{dataCenterService: dataCenterService}
}

func (h *DataCenterHandler) PostDataCenter(c *gin.Context) {
	// Add nil check for dataCenterService
	if h.dataCenterService == nil {
		c.JSON(500, gin.H{"error": "datacenter service not initialized"})
		return
	}

	var data Data
	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	mockData, num, err := h.dataCenterService.GetMockData(data.PageNo, data.PageSize)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"data": mockData,
		"num":  num,
	})
	return
}

// RegisterRoutes registers the routes for the DataCenterHandler
func (h *DataCenterHandler) RegisterRoutes(apiGroup *gin.RouterGroup) {
	// Add nil check for the handler itself
	if h == nil {
		panic("DataCenterHandler is nil when registering routes")
	}

	dataCenterGroup := apiGroup.Group("/datacenter")
	{
		dataCenterGroup.POST("", h.PostDataCenter)
	}
}
