# API 接口文档

## 目录
- [通用说明](#通用说明)
  - [接口规范](#接口规范)
  - [公共请求头](#公共请求头)
  - [公共响应结构](#公共响应结构)
  - [错误码说明](#错误码说明)
- [用户管理](#用户管理)
  - [获取用户详情](#获取用户详情)
  - [创建用户](#创建用户)
  - [更新用户](#更新用户)
  - [删除用户](#删除用户)
  - [获取用户列表](#获取用户列表)
- [数据中心](#数据中心)
  - [获取数据中心数据](#获取数据中心数据)
- [停诊管理](#停诊管理)
  - [获取停诊列表](#获取停诊列表)

## 通用说明

### 接口规范
- 基础路径: `http://api.example.com`
- 请求方式: REST
- 数据格式: JSON
- 字符编码: UTF-8
- API 版本: v1

### 公共请求头
| 参数名 | 类型 | 必选 | 描述 |
|--------|------|------|------|
| Authorization | string | 是 | 认证令牌，格式：`Bearer {token}` |
| Content-Type | string | 是 | 内容类型，固定值：`application/json` |
| X-Request-ID | string | 否 | 请求追踪ID |
| X-Client-Version | string | 否 | 客户端版本号 |

### 公共响应结构
```json
{
    "code": 0,           // 业务状态码，0 表示成功
    "message": "success", // 状态描述
    "data": {},          // 数据负载
    "trace_id": "xxx"    // 追踪ID
}
```

### 错误码说明
| 错误码 | 描述 | 说明 |
|--------|------|------|
| 0 | 成功 | 请求成功 |
| 400 | 参数错误 | 请求参数不合法 |
| 401 | 未授权 | 未登录或 token 已过期 |
| 403 | 禁止访问 | 无权限访问该接口 |
| 404 | 资源不存在 | 请求的资源不存在 |
| 500 | 服务器错误 | 服务器内部错误 |

## 用户管理

### 获取用户详情

#### 接口说明
- 请求路径: `/api/users/{userId}`
- 请求方式: GET
- 接口说明: 获取指定用户的详细信息

#### 请求参数
##### 路径参数
| 参数名 | 类型 | 必选 | 描述 |
|--------|------|------|------|
| userId | string | 是 | 用户ID |

##### 查询参数
| 参数名 | 类型 | 必选 | 描述 |
|--------|------|------|------|
| fields | string | 否 | 指定返回字段，多个字段用逗号分隔 |

#### 响应参数
| 参数名 | 类型 | 描述 |
|--------|------|------|
| id | string | 用户ID |
| username | string | 用户名 |
| email | string | 邮箱 |
| phone | string | 手机号 |
| status | int | 用户状态（1:正常 2:禁用） |
| created_at | string | 创建时间 |
| updated_at | string | 更新时间 |

#### 请求示例
```curl
curl -X GET \
  'http://api.example.com/api/users/12345?fields=id,username,email' \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.xxx' \
  -H 'Content-Type: application/json'
```

#### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": "12345",
        "username": "zhangsan",
        "email": "zhangsan@example.com",
        "phone": "13800138000",
        "status": 1,
        "created_at": "2024-03-11T10:00:00Z",
        "updated_at": "2024-03-11T10:00:00Z"
    },
    "trace_id": "trace-xxx"
}
```

### 创建用户

#### 接口说明
- 请求路径: `/api/users`
- 请求方式: POST
- 接口说明: 创建新用户

#### 请求参数
| 参数名 | 类型 | 必选 | 描述 | 校验规则 |
|--------|------|------|------|----------|
| username | string | 是 | 用户名 | 长度: 3-20 |
| password | string | 是 | 密码 | 长度: 6-20 |
| email | string | 是 | 邮箱 | 有效邮箱格式 |
| phone | string | 否 | 手机号 | 11位数字 |

#### 请求示例
```curl
curl -X POST \
  'http://api.example.com/api/users' \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.xxx' \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "zhangsan",
    "password": "123456",
    "email": "zhangsan@example.com",
    "phone": "13800138000"
}'
```

#### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": "12345",
        "username": "zhangsan",
        "email": "zhangsan@example.com",
        "phone": "13800138000",
        "status": 1,
        "created_at": "2024-03-11T10:00:00Z",
        "updated_at": "2024-03-11T10:00:00Z"
    },
    "trace_id": "trace-xxx"
}
```

## 数据中心

### 获取数据中心数据

#### 接口说明
- 请求路径: `/api/datacenter`
- 请求方式: POST
- 接口说明: 获取数据中心的分页数据

#### 请求参数
| 参数名 | 类型 | 必选 | 描述 | 默认值 |
|--------|------|------|------|--------|
| pageNo | int | 是 | 页码 | 1 |
| pageSize | int | 是 | 每页数量 | 10 |

#### 响应参数
| 参数名 | 类型 | 描述 |
|--------|------|------|
| data | array | 数据列表 |
| num | int | 总数量 |

#### 请求示例
```curl
curl -X POST \
  'http://api.example.com/api/datacenter' \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.xxx' \
  -H 'Content-Type: application/json' \
  -d '{
    "pageNo": 1,
    "pageSize": 10
}'
```

#### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "data": [
            {
                "id": "1",
                "name": "数据1",
                "value": 100
            },
            {
                "id": "2",
                "name": "数据2",
                "value": 200
            }
        ],
        "num": 2
    },
    "trace_id": "trace-xxx"
}
```

## 停诊管理

### 获取停诊列表

#### 接口说明
- 请求路径: `/api/v1/stop/list`
- 请求方式: GET
- 接口说明: 获取所有科室的停诊信息

#### 响应参数
| 参数名 | 类型 | 描述 |
|--------|------|------|
| code | int | 业务状态码 |
| msg | string | 状态描述 |
| data | array | 停诊数据列表 |

#### 请求示例
```curl
curl -X GET \
  'http://api.example.com/api/v1/stop/list' \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.xxx' \
  -H 'Content-Type: application/json'
```

#### 响应示例
```json
{
    "code": 0,
    "msg": "success",
    "data": [
        {
            "room": "手术室1",
            "status": "停诊",
            "reason": "设备维护",
            "start_time": "2024-03-11T09:00:00Z",
            "end_time": "2024-03-11T17:00:00Z"
        },
        {
            "room": "手术室2",
            "status": "正常",
            "reason": "",
            "start_time": "",
            "end_time": ""
        }
    ]
}
```

## 最佳实践

### 接口设计规范
1. URL 命名规范
   - 使用小写字母、数字和连字符
   - 使用名词表示资源
   - 使用复数形式表示集合
   - 避免使用动词

2. HTTP 方法使用规范
   - GET: 获取资源
   - POST: 创建资源
   - PUT: 更新资源
   - DELETE: 删除资源

3. 状态码使用规范
   - 200: 成功
   - 201: 创建成功
   - 204: 删除成功
   - 400: 请求参数错误
   - 401: 未授权
   - 403: 禁止访问
   - 404: 资源不存在
   - 500: 服务器错误

### 安全建议
1. 所有接口都应该使用 HTTPS
2. 敏感数据传输时需要加密
3. 实现请求频率限制
4. 实现接口访问权限控制
5. 记录详细的访问日志

### 性能优化
1. 合理使用缓存
2. 实现数据分页
3. 支持条件查询
4. 使用适当的数据压缩
5. 控制响应数据大小 