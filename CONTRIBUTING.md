# Contributing to EasyDrop

感谢你为 EasyDrop 做贡献。

这个文档主要说明如何在本地开发、提交代码和发起 Pull Request。运行与部署相关的完整说明请优先参考 [README.md](README.md)。

## 贡献方式

欢迎以下类型的贡献：

- Bug 修复
- 新功能开发
- 性能优化
- 测试补充
- 文档改进
- 工程化与可维护性提升

如果改动较大，建议先开 Issue 或发起讨论，避免实现方向和项目规划冲突。

## 开发环境

建议本地具备以下环境：

- Go `1.25`
- Node.js `24`
- pnpm `10`

项目主要包含两部分：

- 后端：Go + Gin + Gorm + Viper + Wire
- 前端：React 19 + TanStack Start + Vite + Tailwind CSS 4

## 快速开始

### 1. 克隆仓库

```bash
git clone <your-fork-or-repo-url>
cd easydrop
```

### 2. 准备配置文件

```bash
cp example/config.example.yaml data/config.yaml
```

Windows PowerShell:

```powershell
Copy-Item example\config.example.yaml data\config.yaml
```

### 3. 生成本地 JWT 密钥

```bash
go run . generate-jwt-token data/jwt --force
```

### 4. 启动后端

```bash
go run .
```

如需显式指定配置目录：

```bash
go run . --config-dir data
```

### 5. 启动前端

```bash
cd web
pnpm install
pnpm dev
```

如果前后端分开启动，通常需要指定后端 API 地址：

```bash
VITE_API_BASE_URL=http://localhost:8080/api/v1 pnpm dev
```

Windows PowerShell:

```powershell
$env:VITE_API_BASE_URL="http://localhost:8080/api/v1"
pnpm dev
```

## 仓库结构

核心目录如下：

- `main.go`：CLI 入口与应用启动
- `internal/`：后端核心代码
- `docs/`：Swagger 生成产物
- `example/`：示例配置与模板
- `web/`：前端项目
- `.github/`：GitHub 工作流与协作模板

## 开发约定

### 代码风格

- Go 代码请保持 `gofmt` 风格。
- Go 源码使用标准 tab 缩进。
- 导出标识符使用 `CamelCase`，非导出标识符使用 `camelCase`。
- 注释保持简洁，项目代码中优先使用中文注释。
- 保持 handler / service / repo 分层清晰，不要把多层职责揉在一起。

### 生成文件

不要手动修改以下生成文件：

- `internal/di/wire_gen.go`
- `docs/docs.go`
- `web/docs/docs.go`
- `web/src/routeTree.gen.ts`

如果你的改动影响这些内容，请使用正确命令重新生成。

### 配置与安全

- 本地配置默认使用 `data/config.yaml`，环境变量前缀为 `EASYDROP_`。
- JWT 使用文件路径配置，不要把密钥内容直接写进配置文件。
- 提交前请确认没有把本地密钥、敏感配置、数据库文件或临时调试内容带入版本库。

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

## 提交前检查

提交前请尽量完成以下检查：

- 相关代码已经格式化
- 相关测试已经补充或更新
- 后端测试可通过：`go test ./...`
- 如果涉及嵌入式前端构建，执行：`go test -tags embed_frontend ./...`
- 如果涉及前端改动，至少执行：`cd web && pnpm test`
- 如果涉及构建或发布行为，执行对应的 `build` / `check`
- 如果涉及 Wire、Swagger 或路由生成产物，已重新生成并确认结果正确

如果某些检查没有执行，请在 PR 描述中明确说明原因。

## 提交信息

提交信息建议使用轻量级 Conventional Commit 风格，例如：

```text
feat: 支持附件批量删除
fix: 修复初始化接口重复提交时的状态码
docs: 补充本地开发说明
refactor: 简化评论查询逻辑
```

不强制要求非常严格的格式，但请保证提交信息能清楚表达改动目的。

## Pull Request 规范

发起 PR 时，请使用仓库内的 PR 模板，并尽量写清楚以下内容：

- 这次改动解决了什么问题
- 核心改动有哪些
- 影响范围是什么
- 如何验证
- 是否存在破坏性变更、配置变更或迁移步骤
- 如果涉及 UI 或接口展示，附上截图、录屏或示例响应

如果你的改动较大，请把 PR 描述写到让 reviewer 不需要翻完整个 diff 才能理解背景。

## 测试建议

### 后端

- 使用 Go 原生 `testing`。
- 测试文件以 `_test.go` 结尾。
- 测试函数以 `Test...` 开头。
- 对新增业务逻辑、边界条件和修复过的 Bug，优先补测试。

### 前端

- 前端测试位于 `web/` 目录。
- 使用 Vitest。
- 涉及交互、状态流转、路由和请求行为的改动，建议补充对应测试。

## 文档更新

出现以下情况时，请同步更新文档：

- 新增或修改配置项
- 新增或修改公开接口
- 调整初始化、部署或升级流程
- 调整前端使用方式或开发命令
- 修改 Swagger、示例配置或邮件模板

## Review 预期

代码评审通常会重点关注以下内容：

- 行为是否正确
- 是否引入兼容性风险
- 分层是否合理
- 测试是否覆盖关键路径
- 文档和生成产物是否同步
- 是否混入无关改动

请尽量让每个 PR 聚焦一个明确目标，避免把重构、功能、格式化和无关清理混在同一个 PR 中。

## 需要帮助？

如果你不确定某个改动该怎么做，可以：

- 先开 Issue 说明背景和方案
- 提交 Draft PR 提前讨论
- 在 PR 中明确标出你希望 reviewer 重点关注的部分
