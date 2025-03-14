package datacenter

import (
	"fmt"
	"time"
)

const ServiceName = "datacenter"

type MockDataCenter struct{}

func NewMockDataCenter() *MockDataCenter {
	return &MockDataCenter{}
}

// GetMockData 获取模拟数据
func (m *MockDataCenter) GetMockData(page, pageSize int) ([]MockItem, int, error) {
	totalItems := 95 // 总共95条数据
	totalPages := (totalItems + pageSize - 1) / pageSize

	if page < 1 || page > totalPages {
		return nil, 0, fmt.Errorf("invalid page number")
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > totalItems {
		end = totalItems
	}

	items := make([]MockItem, 0, pageSize)
	for i := start; i < end; i++ {
		items = append(items, MockItem{
			ID:        i + 1,
			Title:     fmt.Sprintf("Item %d", i+1),
			Content:   fmt.Sprintf("Content for item %d", i+1),
			CreatedAt: time.Now(),
		})
	}

	return items, totalItems, nil
}

type MockItem struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
