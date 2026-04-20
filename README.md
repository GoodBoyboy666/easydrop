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

### 1. 准备配置

复制示例配置：

```bash
cp example/config.example.yaml data/config.yaml
```

Windows PowerShell:

```powershell
Copy-Item example\config.example.yaml data\config.yaml
```

默认配置使用 SQLite，本地开发开箱即可跑。

### 2. 生成 JWT 密钥

```bash
go run . generate-jwt-token data/jwt --force
```

执行后会生成：

- `data/jwt/private.pem`
- `data/jwt/public.pem`

如果希望在启动时自动检查并在缺失时生成，可以启用启动参数：

```bash
go run . --auto-generate-jwt
```

自动模式仅在 `private.pem` 和 `public.pem` 都不存在时生成；若仅存在一个文件，会直接报错并终止启动。

### 3. 启动后端

```bash
go run .
```

或显式指定配置目录：

```bash
go run . --config-dir data
```

默认监听地址为 `:8080`。

### 4. 首次初始化系统

首次部署时，需要创建首个管理员账号并写入基础站点配置。

如果你已经有前端站点在运行，直接访问初始化页：

```text
/init
```

如果你只启动了后端，也可以直接调用初始化接口：

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

初始化完成后：

- 首个管理员会被创建
- 系统会写入基础站点设置
- 再次初始化会返回 `409 Conflict`

## 前端开发

安装依赖：

```bash
cd web
pnpm install
```

启动开发服务器：

```bash
pnpm dev
```

默认前端地址为 `http://localhost:3000`。

前端默认请求 `/api/v1`。如果前后端分开启动，开发时通常需要显式指定后端地址：

```bash
VITE_API_BASE_URL=http://localhost:8080/api/v1 pnpm dev
```

Windows PowerShell:

```powershell
$env:VITE_API_BASE_URL="http://localhost:8080/api/v1"
pnpm dev
```

## 常用命令

### 后端

```bash
go run .
go run . --config-dir data
go run -tags embed_frontend .
go run . generate-jwt-token data/jwt --force
go test ./...
go test -tags embed_frontend ./...
go build ./...
go build -tags embed_frontend ./
go generate ./internal/di
docker build -t easydrop .
docker compose pull
docker compose up -d
```

### 前端

```bash
cd web
pnpm dev
pnpm build
pnpm preview
pnpm test
pnpm lint
pnpm format
pnpm check
```

## 配置说明

### 后端配置来源

- 配置文件：`data/config.yaml`
- 环境变量前缀：`EASYDROP_`
- 环境变量会覆盖配置文件中的同名项

例如：

```bash
EASYDROP_SERVER_ADDR=:9090
EASYDROP_DB_DRIVER=postgres
```

### 主要配置项

- `server`：运行模式、监听地址、超时、可信代理
- `auth_cookie`：登录 Cookie 配置
- `db`：数据库配置，支持 `sqlite`、`mysql`、`postgres`
- `redis`：Redis 连接配置
- `rate_limit`：限流规则配置
- `email`：SMTP 邮件发送配置
- `jwt`：JWT 密钥路径、签发者、过期时间
- `captcha`：验证码开关与供应商配置
- `storage`：文件存储，支持 `local` 和 `s3`
- `token`：站内令牌命名空间配置

### 存储说明

本地存储默认目录：

- `data/uploads/file`
- `data/uploads/avatar`

使用本地存储时，后端会自动注册静态访问路由：

- `/api/file/...`
- `/api/avatar/...`

如果配置了 `storage.local.url_prefix`，路由会变成 `/api/<prefix>/file/...` 和 `/api/<prefix>/avatar/...`。

## API 与文档

- API 前缀：`/api/v1`
- Swagger 地址：`/api/swagger/index.html`

Swagger 只会在 `server.mode=development` 时注册。

主要接口分组：

- `/api/v1/auth`：注册、登录、登出、找回密码、邮箱验证
- `/api/v1/init`：系统初始化
- `/api/v1/posts`：公共日志读取
- `/api/v1/comments`：公共评论读取
- `/api/v1/users/me`：当前用户资料、密码、邮箱、评论
- `/api/v1/attachments`：当前用户附件
- `/api/v1/admin/*`：后台管理接口

## 测试

后端测试：

```bash
go test ./...
```

前端测试：

```bash
cd web
pnpm test
```

嵌入前端测试：

```bash
go test -tags embed_frontend ./...
```

## 开发约定

- Go 代码默认使用 `gofmt`
- 请勿修改生成文件：
  - `internal/di/wire_gen.go`
  - `docs/docs.go`
  - `web/docs/docs.go`
  - `web/src/routeTree.gen.ts`

## Docker 部署

项目根目录已提供：

- `Dockerfile`：多阶段构建前端与后端，并将 `web/dist` 通过 `embed_frontend` 打进后端二进制
- `docker-compose.yml`：默认从 `ghcr.io/goodboyboy666/easydrop:latest` 拉取镜像并启动单个 `app` 服务

### 1. 准备运行目录

先准备配置文件：

```bash
mkdir -p data
cp example/config.example.yaml data/config.yaml
```

Windows PowerShell:

```powershell
New-Item -ItemType Directory -Force data | Out-Null
Copy-Item example\config.example.yaml data\config.yaml
```

容器默认读取 `/app/data/config.yaml`，因此宿主机的 `./data` 会作为运行时持久化目录：

- `data/config.yaml`
- `data/easydrop.db`
- `data/jwt/private.pem`
- `data/jwt/public.pem`
- `data/uploads/...`

### 2. JWT 密钥策略

镜像默认启动参数已包含 `--auto-generate-jwt`，容器启动时会执行以下策略：

- 当 `data/jwt/private.pem` 和 `data/jwt/public.pem` 都不存在时：自动生成一对密钥。
- 当两个文件都存在时：跳过生成，直接启动。
- 当仅存在其中一个文件时：启动失败并提示手动修复（安全优先，避免误轮换密钥）。

如需手动生成（例如首次离线准备或强制轮换），可以执行：

```bash
docker compose pull
docker compose run --rm app generate-jwt-token data/jwt --force
```

### 3. 启动服务

```bash
docker compose pull
docker compose up -d
```

启动后默认监听：

```text
http://localhost:8080
```

因为镜像使用了 `embed_frontend` 构建标签，所以根路径 `/` 会直接提供前端页面，API 仍然位于 `/api/v1`。

如果你希望自己在本地重新构建镜像，仍然可以单独执行：

```bash
docker build -t easydrop .
```

但 `docker-compose.yml` 的默认行为是拉取 GHCR 已发布镜像，而不是使用本地 `build`。

### 4. 首次初始化

首次部署后，访问：

```text
http://localhost:8080/init
```

如果通过 Docker 部署到其他域名，请记得同步调整 `data/config.yaml` 里的站点地址、Cookie 域名、可信代理等生产配置。

## 部署提示

- 默认 `go build .` / `go run .` 产物为纯后端，不托管前端页面
- 若希望单个后端二进制直接提供前端，请先执行 `cd web && pnpm build`，再使用 `go build -tags embed_frontend ./` 或 `go run -tags embed_frontend .`
- `embed_frontend` 构建会把 `web/dist` 打进二进制，并在 `/` 提供前端站点；非 `/api` 的页面路由会回退到 `index.html`
- Docker 镜像已默认执行前端构建并使用 `embed_frontend` 打包，无需手动先构建 `web/dist`
- 仓库会在推送 `v*` tag 时自动发布 GHCR 镜像到 `ghcr.io/goodboyboy666/easydrop`
- `docker compose` 默认拉取 `ghcr.io/goodboyboy666/easydrop:latest`，并挂载 `./data:/app/data`
- 若你 fork 了仓库，需同步调整 `docker-compose.yml` 中的镜像地址到自己的 GHCR 路径
- 生产环境建议显式配置 `server.trusted_proxies`
- 生产环境应使用独立的 JWT 密钥文件，不要复用开发环境密钥
- 若启用邮件找回密码或邮箱验证，需要先正确配置 `email`
- 若前端与后端不在同一域名 / 同一路径下，需正确配置 API 基地址、Cookie 域和反向代理
- 若部署在反向代理后，请确认真实 IP 头与可信代理配置匹配

## License

[MIT](./LICENSE)
