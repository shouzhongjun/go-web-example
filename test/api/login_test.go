package api

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLogin(t *testing.T) {
	var wg sync.WaitGroup
	errChan := make(chan error, 100) // 用于存储错误

	// 100 个并发请求
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Task %d started\n", id)

			url := "http://192.168.16.111:8080/api/users/login"
			method := "POST"

			payload := strings.NewReader(`{
				"username": "testuser",
				"password": "123456"
			}`)

			client := &http.Client{
				Timeout: 5 * time.Second, // 超时控制
			}
			req, err := http.NewRequest(method, url, payload)
			if err != nil {
				errChan <- fmt.Errorf("task %d: request creation error: %v", id, err)
				return
			}

			req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnb1dlYkV4YW1wbGUiLCJleHAiOjE3NDI2MzM0OTUsIm5iZiI6MTc0MjU0NzA5NSwiaWF0IjoxNzQyNTQ3MDk1LCJ1c2VyX2lkIjoiNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAwIiwidXNlcm5hbWUiOiJ0ZXN0dXNlciIsIm5pY2tuYW1lIjoibmljaGVuZyIsImlzX2FkbWluIjpmYWxzZX0.tYgUHDHfekJYsB30jCpgiT02aRz2QfsUqVJmdblDUQ0")
			req.Header.Add("Content-Type", "application/json")

			res, err := client.Do(req)
			if err != nil {
				errChan <- fmt.Errorf("task %d: request error: %v", id, err)
				return
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				errChan <- fmt.Errorf("task %d: read response error: %v", id, err)
				return
			}

			fmt.Printf("Task %d response: %s\n", id, string(body))
		}(i)
	}

	// 等待所有 goroutine 完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误
	for err := range errChan {
		t.Errorf("Error: %v", err)
	}

	fmt.Println("All tasks completed.")
}
