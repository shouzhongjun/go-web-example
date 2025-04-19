# Go Web Example

一个基于 Go 语言的 Web 服务示例项目，集成了多种常用组件和最佳实践。

## 功能特性

- 服务注册与发现 (ETCD)
- RESTful API 支持 (Gin)
- 数据库集成 (MySQL, MongoDB)
- 缓存支持 (Redis)
- 消息队列 (Kafka)
- 配置管理 (Viper)
- 结构化日志 (Zap)
- 监控指标 (Prometheus)
- 健康检查
- 优雅关闭

## 技术栈

- Go 1.21+
- ETCD 3.5+
- MySQL 8.0+
- Redis 6.0+
- MongoDB 4.4+
- Kafka 2.8+

## 快速开始

### 环境要求

确保已安装以下组件：
- Go 1.21 或更高版本
- Docker 和 Docker Compose（可选，用于本地开发）

### 安装

1. 克隆项目
```bash
git clone https://github.com/yourusername/goWebExample.git
cd goWebExample
```

2. 安装依赖
```bash
go mod tidy
```

3. 配置环境
```bash
cp configs/config.dev.yaml configs/config.yaml
# 根据需要修改配置文件
```

4. 运行服务
```bash
go run cmd/main.go
```

### 使用 Docker 运行

1. 构建镜像
```bash
docker build -t go-web-example .
```

2. 运行容器
```bash
docker-compose up -d
```

## 项目结构

```
.
├── cmd/                # 主程序入口
├── configs/            # 配置文件
├── internal/           # 内部包
│   ├── configs/        # 配置结构
│   ├── infrastructure/ # 基础设施
│   │   ├── connector/  # 连接器
│   │   ├── discovery/  # 服务发现
│   │   └── store/     # 存储层
│   ├── middleware/    # 中间件
│   ├── models/        # 数据模型
│   ├── router/        # 路由
│   └── service/       # 业务逻辑
├── pkg/               # 公共包
├── scripts/          # 脚本
└── test/             # 测试
```

## 配置说明

主要配置项说明（configs/config.yaml）：

### 服务器配置
```yaml
server:
  serverName: go-server-rest-api
  port: 8080
  host: 0.0.0.0
  version: v1.0.0
```

### ETCD 配置
```yaml
etcd:
  enable: true
  host: localhost
  port: 2379
  dialTimeOut: 5
  leaseTTL: 60
  maxRetries: 3
  retryInterval: 1
  healthCheck:
    enable: true
    interval: 10
```

## API 文档

### 健康检查
```
GET /health
响应: {"status": "UP"}
```

### 服务信息
```
GET /info
响应: {
  "name": "go-server-rest-api",
  "version": "v1.0.0",
  "status": "running"
}
```

## 监控指标

项目集成了 Prometheus 监控，主要指标包括：

- `etcd_service_registrations_total`: 服务注册次数
- `etcd_service_registration_failures_total`: 服务注册失败次数
- `etcd_lease_renewals_total`: 租约续租次数
- `etcd_lease_ttl_seconds`: 租约 TTL

访问 `/metrics` 端点获取完整指标。

## 测试

运行单元测试：
```bash
go test ./...
```

运行集成测试：
```bash
go test -tags=integration ./...
```

## 部署

### 使用 Docker

1. 构建镜像：
```bash
docker build -t go-web-example:latest .
```

2. 运行容器：
```bash
docker run -d -p 8080:8080 go-web-example:latest
```

### 使用 Kubernetes

1. 应用配置：
```bash
kubectl apply -f deploy/kubernetes/
```

## 自动构建与打包

本项目使用 GitHub Actions 实现自动构建和打包。每当代码推送到主分支或创建 Pull Request 时，自动构建流程会被触发。

### 自动构建流程

1. 检出代码
2. 设置 Go 环境
3. 安装依赖
4. 生成 Wire 依赖注入代码
5. 运行测试
6. 执行代码质量检查
7. 生成 Swagger 文档
8. 构建多平台二进制文件
9. 创建分发包
10. 上传构建产物

### 支持的平台

自动构建支持以下平台：
- Linux AMD64
- Linux ARM64
- macOS AMD64
- macOS ARM64

### 发布流程

当创建新的 Git 标签（tag）时，GitHub Actions 会自动创建一个新的 GitHub Release，并将构建产物作为附件上传。

使用以下命令创建新版本：

```bash
git tag v1.0.0
git push origin v1.0.0
```

### 手动触发构建

您也可以在 GitHub 仓库的 Actions 页面手动触发构建流程。

### 下载构建产物

每次工作流运行后，构建产物会通过 actions/upload-artifact@v3 被上传并保存在 GitHub Actions 中。您可以通过以下步骤下载这些构建产物：

1. 访问您的 GitHub 仓库
2. 点击 "Actions" 标签页
3. 从列表中选择一个工作流运行记录
4. 在工作流运行详情页面，滚动到页面底部的 "Artifacts" 部分
5. 点击 "binaries" 下载包含所有构建产物的 zip 文件

下载的 zip 文件包含：
- 所有平台的二进制文件 (`build/my-app-*`)
- 所有平台的分发包 (`dist/*.tar.gz`)

注意：构建产物在 GitHub 上的保存期限为 90 天。

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

项目维护者 - [@yourusername](https://github.com/yourusername)

项目链接: [https://github.com/yourusername/goWebExample](https://github.com/yourusername/goWebExample) 
