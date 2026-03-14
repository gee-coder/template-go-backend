# template-go-backend

一套适用于一人公司创业场景的 Golang 后端基础模板，内置 Gin、GORM、Viper、Zap、JWT、RBAC 与 Swagger/OpenAPI 文档基础能力。

## 技术栈

- Go 1.22.10
- Gin 1.9.1
- GORM 1.25.12
- Viper 1.18.2
- Zap 1.26.0
- Validator 10.27.0
- JWT 5.3.1
- Swagger/OpenAPI 3
- MySQL 8.0.45
- Redis 7.2.13

## 目录说明

- `cmd/api`: API 启动入口，仅负责组装依赖和启动服务
- `internal/api`: Handler、中间件、请求结构体与路由注册
- `internal/service`: 核心业务逻辑
- `internal/repository`: Repository 接口、GORM Model 与 MySQL 实现
- `internal/config`: 配置结构与加载逻辑
- `internal/utils`: 通用工具，如响应、密码、Token、错误等
- `configs`: YAML 配置文件
- `docs`: OpenAPI 文档与说明
- `scripts`: 本地开发脚本

## 快速开始

1. 复制环境变量模板：

```powershell
Copy-Item .env.example .env
```

2. 启动依赖：

```powershell
docker compose up -d
```

3. 启动服务：

```powershell
$env:APP_CONFIG = "configs/config.local.yaml"
go run ./cmd/api
```

## 默认账号

- 用户名：`admin`
- 密码：`Admin123!`

首次启动会自动创建基础角色、菜单与管理员账号。

## 常用命令

```powershell
go test ./...
go build ./cmd/api
```

## API 概览

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/profile`
- `GET /api/v1/healthz`
- `GET/POST/PUT/DELETE /api/v1/system/users`
- `GET/POST/PUT/DELETE /api/v1/system/roles`
- `GET/POST/PUT/DELETE /api/v1/system/menus`
- `POST /api/v1/public/contact-submissions`

## 扩展建议

- 在 `internal/service` 中新增业务模块
- 在 `internal/repository/model` 中扩展领域模型
- 在 `internal/api/handler` 中新增控制器并通过路由暴露
- 如需多租户，可在 Model、Claims 与 Repository 查询层统一引入租户条件
