# API 调用代码示例

本文档提供了多种编程语言调用 API 的示例代码。

## Go 示例

### 基础 HTTP 客户端
```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

// APIClient API 客户端
type APIClient struct {
    baseURL    string
    token      string
    httpClient *http.Client
}

// NewAPIClient 创建 API 客户端
func NewAPIClient(baseURL, token string) *APIClient {
    return &APIClient{
        baseURL: baseURL,
        token:   token,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

// doRequest 执行 HTTP 请求
func (c *APIClient) doRequest(method, path string, body interface{}) ([]byte, error) {
    var reqBody io.Reader
    if body != nil {
        jsonBody, err := json.Marshal(body)
        if err != nil {
            return nil, err
        }
        reqBody = bytes.NewBuffer(jsonBody)
    }

    req, err := http.NewRequest(method, c.baseURL+path, reqBody)
    if err != nil {
        return nil, err
    }

    // 设置请求头
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+c.token)
    req.Header.Set("X-Request-ID", fmt.Sprintf("req-%d", time.Now().UnixNano()))

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}
```

### 用户管理示例
```go
// User 用户结构
type User struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Phone     string    `json:"phone"`
    Status    int       `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Email    string `json:"email"`
    Phone    string `json:"phone"`
}

// GetUser 获取用户详情
func (c *APIClient) GetUser(userID string) (*User, error) {
    data, err := c.doRequest("GET", "/api/users/"+userID, nil)
    if err != nil {
        return nil, err
    }

    var resp struct {
        Code    int    `json:"code"`
        Message string `json:"message"`
        Data    User   `json:"data"`
    }
    if err := json.Unmarshal(data, &resp); err != nil {
        return nil, err
    }

    if resp.Code != 0 {
        return nil, fmt.Errorf("API error: %s", resp.Message)
    }

    return &resp.Data, nil
}

// CreateUser 创建用户
func (c *APIClient) CreateUser(req CreateUserRequest) (*User, error) {
    data, err := c.doRequest("POST", "/api/users", req)
    if err != nil {
        return nil, err
    }

    var resp struct {
        Code    int    `json:"code"`
        Message string `json:"message"`
        Data    User   `json:"data"`
    }
    if err := json.Unmarshal(data, &resp); err != nil {
        return nil, err
    }

    if resp.Code != 0 {
        return nil, fmt.Errorf("API error: %s", resp.Message)
    }

    return &resp.Data, nil
}
```

### 数据中心示例
```go
// DataCenterRequest 数据中心请求
type DataCenterRequest struct {
    PageNo   int `json:"pageNo"`
    PageSize int `json:"pageSize"`
}

// DataCenterItem 数据中心数据项
type DataCenterItem struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Value int    `json:"value"`
}

// GetDataCenterData 获取数据中心数据
func (c *APIClient) GetDataCenterData(req DataCenterRequest) ([]DataCenterItem, int, error) {
    data, err := c.doRequest("POST", "/api/datacenter", req)
    if err != nil {
        return nil, 0, err
    }

    var resp struct {
        Code    int    `json:"code"`
        Message string `json:"message"`
        Data    struct {
            Data []DataCenterItem `json:"data"`
            Num  int             `json:"num"`
        } `json:"data"`
    }
    if err := json.Unmarshal(data, &resp); err != nil {
        return nil, 0, err
    }

    if resp.Code != 0 {
        return nil, 0, fmt.Errorf("API error: %s", resp.Message)
    }

    return resp.Data.Data, resp.Data.Num, nil
}
```

### 使用示例
```go
func main() {
    client := NewAPIClient("http://api.example.com", "your-token-here")

    // 获取用户详情
    user, err := client.GetUser("12345")
    if err != nil {
        panic(err)
    }
    fmt.Printf("User: %+v\n", user)

    // 创建用户
    newUser, err := client.CreateUser(CreateUserRequest{
        Username: "zhangsan",
        Password: "123456",
        Email:    "zhangsan@example.com",
        Phone:    "13800138000",
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("New user: %+v\n", newUser)

    // 获取数据中心数据
    items, total, err := client.GetDataCenterData(DataCenterRequest{
        PageNo:   1,
        PageSize: 10,
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("Total: %d, Items: %+v\n", total, items)
}
```

## Python 示例

### 基础 HTTP 客户端
```python
import time
import requests
from typing import Optional, Any, Dict

class APIClient:
    def __init__(self, base_url: str, token: str):
        self.base_url = base_url
        self.token = token
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {token}'
        })

    def do_request(self, method: str, path: str, data: Optional[Dict] = None) -> Dict:
        url = self.base_url + path
        headers = {
            'X-Request-ID': f'req-{int(time.time() * 1000000)}'
        }

        response = self.session.request(
            method=method,
            url=url,
            json=data,
            headers=headers
        )
        response.raise_for_status()
        return response.json()
```

### 用户管理示例
```python
from dataclasses import dataclass
from datetime import datetime
from typing import Optional

@dataclass
class User:
    id: str
    username: str
    email: str
    phone: str
    status: int
    created_at: datetime
    updated_at: datetime

    @classmethod
    def from_dict(cls, data: dict) -> 'User':
        return cls(
            id=data['id'],
            username=data['username'],
            email=data['email'],
            phone=data['phone'],
            status=data['status'],
            created_at=datetime.fromisoformat(data['created_at'].replace('Z', '+00:00')),
            updated_at=datetime.fromisoformat(data['updated_at'].replace('Z', '+00:00'))
        )

class UserAPI:
    def __init__(self, client: APIClient):
        self.client = client

    def get_user(self, user_id: str) -> User:
        response = self.client.do_request('GET', f'/api/users/{user_id}')
        if response['code'] != 0:
            raise Exception(f"API error: {response['message']}")
        return User.from_dict(response['data'])

    def create_user(self, username: str, password: str, email: str, phone: Optional[str] = None) -> User:
        data = {
            'username': username,
            'password': password,
            'email': email
        }
        if phone:
            data['phone'] = phone

        response = self.client.do_request('POST', '/api/users', data)
        if response['code'] != 0:
            raise Exception(f"API error: {response['message']}")
        return User.from_dict(response['data'])
```

### 数据中心示例
```python
from dataclasses import dataclass
from typing import List, Tuple

@dataclass
class DataCenterItem:
    id: str
    name: str
    value: int

    @classmethod
    def from_dict(cls, data: dict) -> 'DataCenterItem':
        return cls(
            id=data['id'],
            name=data['name'],
            value=data['value']
        )

class DataCenterAPI:
    def __init__(self, client: APIClient):
        self.client = client

    def get_data(self, page_no: int = 1, page_size: int = 10) -> Tuple[List[DataCenterItem], int]:
        data = {
            'pageNo': page_no,
            'pageSize': page_size
        }
        response = self.client.do_request('POST', '/api/datacenter', data)
        if response['code'] != 0:
            raise Exception(f"API error: {response['message']}")

        items = [DataCenterItem.from_dict(item) for item in response['data']['data']]
        total = response['data']['num']
        return items, total
```

### 使用示例
```python
def main():
    client = APIClient('http://api.example.com', 'your-token-here')
    user_api = UserAPI(client)
    datacenter_api = DataCenterAPI(client)

    try:
        # 获取用户详情
        user = user_api.get_user('12345')
        print(f'User: {user}')

        # 创建用户
        new_user = user_api.create_user(
            username='zhangsan',
            password='123456',
            email='zhangsan@example.com',
            phone='13800138000'
        )
        print(f'New user: {new_user}')

        # 获取数据中心数据
        items, total = datacenter_api.get_data(page_no=1, page_size=10)
        print(f'Total: {total}')
        for item in items:
            print(f'Item: {item}')

    except Exception as e:
        print(f'Error: {e}')

if __name__ == '__main__':
    main()
```

## 最佳实践

### 错误处理
1. 使用自定义错误类型
2. 实现重试机制
3. 处理超时情况
4. 记录详细错误日志

### 性能优化
1. 使用连接池
2. 启用 Keep-Alive
3. 实现请求缓存
4. 使用异步客户端

### 安全建议
1. 使用环境变量存储敏感信息
2. 实现请求签名
3. 验证服务器证书
4. 实现令牌自动刷新 