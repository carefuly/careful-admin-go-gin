<p align="center">
  <img src="./favicon.svg" width="200" height="200" />
</p>

### 用心系统 (Careful System)

一个基于 Gin 的后端服务，集成 GORM、Redis、Nacos 配置中心、Zap 日志、请求中间件、表单校验国际化以及 Swagger 接口文档。

- **运行端口**: 默认 `8080`
- **API 前缀**: `/dev-api`
- **健康检查**: `/health`
- **Swagger**: `/swagger/index.html`
- **静态资源**: `/static`（将 `favicon.ico` 放到 `./static/` 即可挂载为 `/static/favicon.ico`）

#### 技术栈

- GO, Gin, GORM (MySQL), go-redis, Nacos, Viper, Zap, swag (Swagger)
- 校验国际化（默认 `zh`，可切换为 `en`）

#### 快速开始

1) 安装依赖

```bash
# Go 1.23+
go mod download
```

2) 配置本地环境

- 在项目根目录创建 `.env.development.yaml`（开发）和 `.env.production.yaml`（生产），示例见下文。
- 本项目启动时会读取本地 `.env.*.yaml` 获取服务和 Nacos 连接信息，并从 Nacos 拉取全量业务配置。

3) 启动服务

```bash
go run ./main.go
```

4) 访问

- Swagger: `http://127.0.0.1:8080/swagger/index.html`
- 健康检查: `http://127.0.0.1:8080/health`
- API 基础路径: `http://127.0.0.1:8080/dev-api`

#### 配置说明

本地配置文件用于告诉服务如何启动以及如何连接 Nacos。实际业务配置（数据库/缓存/令牌等）存放在 Nacos 中。

- 本地配置文件（根目录）

```yaml
# application.yaml
server:
  host: "localhost"
  port: 8080
application:
  name: "CarefulAdmin 后台管理"
  version: "1.0.0"
  environment: "development" # production
  debug: true
nacos:
  host: 127.0.0.1
  port: 8848
  namespace: public
  user: nacos
  password: nacos
  dataId: careful-admin.yaml
  group: DEFAULT_GROUP
```

- Nacos 配置内容（示例见 `docs/nacos.example.yaml`）
    - 注意：`database` 为 Map 结构，键名需与代码匹配（示例使用 `careful`）。

```yaml
server:
  host: 0.0.0.0
  port: 8080
# 支持多数据源，键名建议为业务含义，如 careful
database:
  careful:
    type: mysql
    host: 127.0.0.1
    port: 3306
    username: root
    password: 123456
    dbname: careful_db
    charset: utf8mb4
    prefix: ""
    maxIdleConns: 10
    maxOpenConns: 100
    connMaxLifetime: 30m

cache:
  host: 127.0.0.1
  port: 6379
  password: ""
  db: 0

token:
  secret: "replace-with-strong-secret"
  expire: 86400
```

- 运行时行为
    - 程序会读取本地 `.env.*.yaml`，连接 Nacos，拉取远程 YAML 并解析为服务全局配置。
    - 本地的 `server` 配置优先级高于 Nacos 中的同名配置。
    - 静态资源目录 `./static` 若存在将自动挂载为 `/static`。
    - 主程序默认初始化校验翻译器语言为 `zh`（见 `main.go` 和 `ioc/server.go`）。

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
- `internal/web`: 中间件、路由与处理器
- `docs`: Swagger 相关
- `static`: 静态资源（放置 `favicon.ico` 等）
