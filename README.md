> [!CAUTION]
> 受上游TanStack供应链被投毒影响，请勿通过源码构建本项目前端，如已部署，请立即轮换构建机器中存储的所有密钥！！！

# EasyDrop

EasyDrop 是一个轻量级的说说 / 日志服务，提供完整的日志发布、后台管理能力。

它适合用来搭建个人动态站、轻量博客式时间线等小型内容站点。

## 功能特性

- 基于时间线的日志发布与展示，支持 Markdown 内容
- 支持日志置顶、隐藏、关闭评论
- 评论系统，支持最新评论列表与按日志查看评论
- 用户注册、登录、退出、个人资料维护、修改密码、修改邮箱
- 后台管理：概览、用户、日志、评论、附件、站点设置
- 附件上传与头像上传，支持本地存储和 S3
- 可选邮件能力：找回密码、邮箱验证、邮箱变更确认
- 可选验证码能力：Turnstile、reCAPTCHA、hCaptcha、GeeTest v4
- 开发模式内置 Swagger 文档

## 技术栈

### 后端

- Go 1.25
- Gin
- Gorm
- Viper
- Wire
- SQLite / MySQL / PostgreSQL
- Redis

### 前端

- React 19
- TanStack Start + TanStack Router + TanStack Query
- Vite
- Tailwind CSS 4
- Vitest

## 仓库结构

```text
.
├─ main.go                  # CLI 入口与应用启动
├─ internal/
│  ├─ config/               # 静态配置加载与默认值
│  ├─ di/                   # Wire 依赖注入
│  ├─ router/               # Gin 路由注册
│  ├─ handler/              # HTTP Handler
│  ├─ service/              # 业务逻辑
│  ├─ repo/                 # 数据访问层
│  ├─ middleware/           # 鉴权、限流、请求体限制等中间件
│  ├─ model/                # Gorm 模型
│  └─ pkg/                  # 数据库、JWT、邮箱、存储、验证码等基础组件
├─ docs/                    # Swagger 生成产物
├─ example/                 # 示例配置与邮件模板
├─ web/                     # 前端项目
└─ data/                    # 本地运行时数据目录
```

## 快速开始

1. 准备配置与 JWT 密钥。

```bash
mkdir -p data
cp example/config.example.yaml data/config.yaml
go run . generate-jwt-token data/jwt --force
```

Windows PowerShell:

```powershell
New-Item -ItemType Directory -Force data | Out-Null
Copy-Item example\config.example.yaml data\config.yaml
go run . generate-jwt-token data/jwt --force
```

2. 启动后端。

```bash
go run .
```

需要自动补齐 JWT 密钥时，可使用：

```bash
go run . --auto-generate-jwt
```

3. 完成首次初始化。

- 页面方式：访问 `http://localhost:8080/init`
- API 方式：

```bash
curl -X POST http://localhost:8080/api/v1/init \
  -H "Content-Type: application/json" \
  -d '{
    "username":"admin",
    "nickname":"管理员",
    "email":"admin@example.com",
    "password":"Admin123456",
    "site_name":"EasyDrop",
    "site_url":"http://localhost:3000",
    "site_announcement":"欢迎使用 EasyDrop",
    "allow_register":true
  }'
```

初始化只允许一次，再次执行会返回 `409 Conflict`。

## Docker 部署

### 准备运行目录

```bash
mkdir -p data
```

Windows PowerShell:

```powershell
New-Item -ItemType Directory -Force data | Out-Null
```

容器首次启动时，若 `data/config.yaml` 不存在，会自动生成一份带默认值的配置文件。
如果你希望先手动调整配置，也可以提前复制模板：

```bash
cp example/config.example.yaml data/config.yaml
```

Windows PowerShell:

```powershell
Copy-Item example\config.example.yaml data\config.yaml
```

如需覆盖配置，可直接编辑 `data/config.yaml`，或在 `docker-compose.yml` 的 `environment` 中补充 `EASYDROP_` 变量（环境变量优先级更高）。

持久化目录统一挂载到 `./data:/app/data`，常见文件包括：

- `data/config.yaml`
- `data/easydrop.db`
- `data/jwt/private.pem`
- `data/jwt/public.pem`
- `data/uploads/...`

### JWT 密钥策略

- `private.pem` 与 `public.pem` 都不存在：自动生成
- 两者都存在：直接启动
- 仅存在一个：启动失败，需手动修复

手动生成命令：

```bash
docker compose pull
docker compose run --rm app generate-jwt-token data/jwt --force
```

### 启动服务

```bash
docker compose pull
docker compose up -d
```

默认地址：`http://localhost:8080`。
镜像已使用 `embed_frontend`，根路径 `/` 可直接访问前端页面，API 仍为 `/api/v1`。

### 首次初始化

访问：`http://localhost:8080/init`

### 可选：本地构建镜像

```bash
docker build -t easydrop .
```

`docker-compose.yml` 默认拉取 GHCR 已发布镜像，不默认使用本地 `build`。

## 前端开发

```bash
cd web
pnpm install
pnpm dev
```

默认访问地址：`http://localhost:3000`。
前端默认请求 `/api/v1`；前后端分开运行时可显式指定：

```bash
VITE_API_BASE_URL=http://localhost:8080/api/v1 pnpm dev
```

Windows PowerShell:

```powershell
$env:VITE_API_BASE_URL="http://localhost:8080/api/v1"
pnpm dev
```

## 配置说明

### 加载规则

- 配置文件：`data/config.yaml`
- 环境变量前缀：`EASYDROP_`
- 环境变量覆盖配置文件同名项

示例：

```bash
EASYDROP_SERVER_ADDR=:9090
EASYDROP_DB_DRIVER=postgres
```

### 关键配置段

- `server`：运行模式、监听地址、超时、可信代理
- `auth_cookie`：登录 Cookie
- `db`：`sqlite` / `mysql` / `postgres`
- `redis`：Redis 连接
- `rate_limit`：限流规则
- `email`：SMTP
- `jwt`：密钥路径、签发者、过期时间
- `captcha`：验证码开关与供应商
- `storage`：`local` / `s3`
- `token`：站内令牌命名空间

### 本地存储路径

- 默认目录：`data/uploads/file`、`data/uploads/avatar`
- 默认访问：`/api/file/...`、`/api/avatar/...`
- 设置 `storage.local.url_prefix` 后，访问路径变为 `/api/<prefix>/file/...` 与 `/api/<prefix>/avatar/...`

## API 与文档

- API 前缀：`/api/v1`
- Swagger：`/api/swagger/index.html`（仅 `server.mode=development` 时启用）
- 主要分组：
  - `/api/v1/auth`
  - `/api/v1/init`
  - `/api/v1/posts`
  - `/api/v1/comments`
  - `/api/v1/users/me`
  - `/api/v1/attachments`
  - `/api/v1/admin/*`

## 测试

```bash
go test ./...
go test -tags embed_frontend ./...
```

```bash
cd web
pnpm test
```

## 开发约定

- Go 代码使用 `gofmt`
- 不要修改生成文件：
  - `internal/di/wire_gen.go`
  - `docs/docs.go`
  - `web/docs/docs.go`
  - `web/src/routeTree.gen.ts`

## 部署提示

- 默认 `go build .` / `go run .` 产物仅包含后端
- 需要后端二进制直接托管前端时，先执行 `cd web && pnpm build`，再用 `-tags embed_frontend`
- Docker 官方镜像已完成前端构建并启用 `embed_frontend`
- 若使用反向代理，务必配置 `server.trusted_proxies`
- JWT 密钥应独立管理，不要复用开发环境密钥
- 启用邮件能力前请先完成 `email` 配置
- fork 项目后请同步更新 `docker-compose.yml` 中镜像地址

## License

[MIT](./LICENSE)
