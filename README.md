<p align="center">
  <img src="./favicon.svg" width="200" height="200" />
</p>

### 用心后台管理系统 (Careful System)

一个基于 Gin 的后端服务，集成 GORM、Redis、Zap 日志、请求中间件、表单校验国际化以及 Swagger 接口文档。

- **运行端口**: 默认 `8080`
- **API 前缀**: `/api`
- **健康检查**: `/health`
- **Swagger**: `/swagger/index.html`
- **静态资源**: `/static`（将 `favicon.ico` 放到 `./static/` 即可挂载为 `/static/favicon.ico`）

#### 技术栈

- GO, Gin, GORM (MySQL), go-redis, Viper, Zap, swag (Swagger)
- 校验国际化（默认 `zh`，可切换为 `en`）

#### 快速开始

1) 安装依赖

```bash
# Go 1.23+
go mod download
```

2) 配置本地环境

- 在项目根目录创建 `application.yaml`（开发）和 `application-pro.yaml`（生产），示例见下文。
- 本项目启动时会读取本地 `application.yaml` 获取信息配置。

3) 启动服务

```bash
go run ./main.go
```

4) 访问

- 健康检查: `http://127.0.0.1:8080/health`
- API 基础路径: `http://127.0.0.1:8080/api`
- Swagger: `http://127.0.0.1:8080/swagger/index.html`

#### 配置说明

- 本地配置文件（根目录）

```yaml
# application.yaml
# 服务配置
server:
  host: "localhost"  # 服务监听地址
  port: 8080         # 服务端口

# 应用配置
application:
  name: "CarefulAdmin 后台管理"
  version: "1.0.0"
  environment: "development"  # 环境: development/production
  debug: true                 # 是否开启调试模式

# 数据库配置 
# 支持多数据源，键名建议为业务含义，如 careful
database:
  careful:
    type: mysql
    host: 127.0.0.1
    port: 3306
    username: root
    password: 123456
    dbname: careful_admin_go_gin
    charset: utf8mb4
    maxIdleConn: 10
    maxOpenConn: 100
    connMaxLifetime: 30m

# 缓存配置
cache:
  host: 127.0.0.1
  port: 6379
  password: ""
  db: 0

# Token配置
token:
  secret: "your-secret-key"  # 替换为你的密钥
  expire: 86400              # 过期时间(秒)
```

#### Swagger 说明

- 已内置 `docs/` 文档，直接可用。
- 若更新了路由注释，可使用 swag 重新生成：

```bash
# 安装 swag（如未安装）
go install github.com/swaggo/swag/cmd/swag@latest
# 在项目根目录运行
swag init -g main.go
```

#### 目录简述

- `config`: 配置结构体定义
- `ioc`: 依赖注入与初始化（配置、DB、缓存、服务器等）
- `internal`: 中间件、路由与处理器等内部文件
- `docs`: Swagger 相关
- `static`: 静态资源（放置 `favicon.ico` 等）

基于Go语言和Gin框架的后台管理项目，注重性能优化与安全性设计，适用于快速构建稳定可靠的Web服务。
