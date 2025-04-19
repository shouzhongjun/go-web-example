package middleware

import (
	"bytes"
	"encoding/json"
	"goWebExample/internal/configs"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestParamLogger 请求参数日志中间件
// 打印路由，请求方式、请求参数
func RequestParamLogger(logger *zap.Logger, config *configs.AllConfig) gin.HandlerFunc {
	if config == nil || !config.Log.PrintParam {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return func(c *gin.Context) {
		// 获取路由和请求方式
		path := c.Request.URL.Path
		method := c.Request.Method

		// 获取请求参数
		var requestParams map[string]interface{}

		// 获取Content-Type
		contentType := c.ContentType()

		// 处理不同类型的请求参数
		if c.Request.Method != http.MethodGet {
			// 对于非GET请求，尝试读取请求体
			if c.Request.Body != nil && (contentType == "" || (contentType != "multipart/form-data" && !strings.Contains(contentType, "multipart/form-data"))) {
				bodyBytes, _ := io.ReadAll(c.Request.Body)
				// 重新设置请求体，因为读取后会消耗
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// 尝试解析JSON
				if len(bodyBytes) > 0 {
					// 使用json.Decoder，它可以提供更详细的错误信息
					decoder := json.NewDecoder(bytes.NewReader(bodyBytes))
					decoder.UseNumber() // 使用Number类型来处理数字，避免精度问题

					err := decoder.Decode(&requestParams)
					if err != nil {
						// 记录详细的错误信息，但继续处理请求
						logger.Error("解析请求体失败",
							zap.Error(err),
							zap.String("content_type", contentType),
							zap.String("raw_body", string(bodyBytes)),
						)

						// 将原始请求体作为字符串存储
						requestParams = map[string]interface{}{
							"raw_body": string(bodyBytes),
						}
					}
				}
			}
		}

		// 获取URL查询参数
		queryParams := make(map[string]interface{})
		for k, v := range c.Request.URL.Query() {
			if len(v) == 1 {
				queryParams[k] = v[0]
			} else {
				queryParams[k] = v
			}
		}

		// 获取表单参数
		formParams := make(map[string]interface{})
		urlEncodedParams := make(map[string]interface{})
		isURLEncoded := contentType == "application/x-www-form-urlencoded" || strings.Contains(contentType, "application/x-www-form-urlencoded")

		if err := c.Request.ParseForm(); err == nil {
			// 处理 POST 表单数据
			for k, v := range c.Request.PostForm {
				if len(v) == 1 {
					if isURLEncoded {
						urlEncodedParams[k] = v[0]
					} else {
						formParams[k] = v[0]
					}
				} else {
					if isURLEncoded {
						urlEncodedParams[k] = v
					} else {
						formParams[k] = v
					}
				}
			}
		}

		// 获取multipart/form-data参数
		multipartFormParams := make(map[string]interface{})
		if contentType != "" && (contentType == "multipart/form-data" || strings.Contains(contentType, "multipart/form-data")) {
			form, err := c.MultipartForm()
			if err == nil {
				// 处理表单字段
				for k, v := range form.Value {
					if len(v) == 1 {
						multipartFormParams[k] = v[0]
					} else {
						multipartFormParams[k] = v
					}
				}

				// 处理文件字段（只记录文件名，不记录文件内容）
				for k, files := range form.File {
					fileNames := make([]string, len(files))
					for i, file := range files {
						fileNames[i] = file.Filename
					}
					multipartFormParams[k] = fileNames
				}
			}
		}

		logger.Info("请求参数日志",
			zap.String("路由", path),
			zap.String("请求方式", method),
			zap.Any("请求体参数", requestParams),
			zap.Any("查询参数", queryParams),
			zap.Any("x-www-form-urlencoded参数", urlEncodedParams),
			zap.Any("表单参数", formParams),
			zap.Any("form-data参数", multipartFormParams),
		)

		// 继续处理请求
		c.Next()
	}
}
