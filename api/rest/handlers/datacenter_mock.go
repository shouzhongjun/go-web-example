package handlers

import (
	"github.com/gin-gonic/gin"
)

// 一共10页，每页返回10条数据，最后一页返回5条数据

type Page struct {
	PageSize int `json:"pageSize"`
	PageNo   int `json:"pageNo"`
}

func ReturnMock(ctx *gin.Context) {
	var page Page
	if err := ctx.ShouldBindJSON(&page); err != nil {
		ctx.JSON(400, gin.H{
			"message": "Invalid request body",
		})
		return
	}
	if page.PageNo < 10 && page.PageNo >= 1 {
		ctx.JSON(200, gin.H{
			"data": []map[string]interface{}{
				{"id": 1, "name": "Item 1"},
				{"id": 2, "name": "Item 2"},
				{"id": 3, "name": "Item 3"},
				{"id": 4, "name": "Item 4"},
				{"id": 5, "name": "Item 5"},
				{"id": 6, "name": "Item 6"},
				{"id": 7, "name": "Item 7"},
				{"id": 8, "name": "Item 8"},
				{"id": 9, "name": "Item 9"},
				{"id": 10, "name": "Item 10"},
			},
		})
		return
	} else if page.PageNo == 10 {
		ctx.JSON(200, gin.H{
			"data": []map[string]interface{}{
				{"id": 1, "name": "Item 1"},
				{"id": 2, "name": "Item 2"},
				{"id": 3, "name": "Item 3"},
			},
		})
		return
	}
	ctx.JSON(404, gin.H{
		"message": "Invalid page number",
	},
	)

}
