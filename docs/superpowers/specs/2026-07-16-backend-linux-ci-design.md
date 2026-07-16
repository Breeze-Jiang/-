# P0 后端 Linux CI 设计

## 目标

在 GitHub Actions 的独立 Ubuntu 环境中，自动验证沿途 P0 Go 后端的静态质量、数据库迁移、Docker Compose 启动顺序与游客找点闭环，避免 Windows 本地开发环境掩盖 Linux 部署问题。

## 触发与范围

工作流位于 `.github/workflows/backend-ci.yml`，在推送 `main`、对 `main` 发起拉取请求，且改动涉及 `backend/**`、工作流文件或本项目根 `.gitignore` 时运行。工作流只验证后端 P0，不构建 Android、iOS、论坛、聊天、计划或真实供应商集成。

## 作业一：quality

在 `ubuntu-latest` 上使用 Go 1.24.x，按固定顺序执行：

1. `gofmt -l` 检查；
2. 安装固定版本 sqlc 并执行 `sqlc generate`，随后用 `git diff --exit-code -- backend/internal/platform/database/dbgen` 阻止生成代码漂移；
3. `go test ./...`；
4. `go test -race ./...`；
5. `go vet ./...`；
6. OpenAPI lint；
7. `docker build` API 镜像。

Go、Node 和 sqlc 的版本均由工作流显式固定或由项目模块文件约束，不依赖 Actions 运行器预装工具。

## 作业二：compose-e2e

作业依赖 `quality` 成功。它在 `backend/` 生成仅 CI 使用、不会被提交的 `.env`：使用随机 JWT/Cursor 密钥、假高德 key、开发环境标记，并不提供真实高德或阿里云凭证。

该作业执行 `docker compose up -d --build`，轮询 `/health/ready`，验证：

- `migrate` 服务退出码为 0；
- `postgres`、`redis` 健康；
- `goose_db_version` 的 1、2 迁移已应用，PostGIS 与 pgcrypto 扩展存在；
- API 的 `/health/live` 和 `/health/ready` 返回 200；
- 通过伪供应商完成游客厕所附近查询、平台地点详情、步行路线与外部导航请求。

无论成功或失败，作业都输出 Compose 状态与容器日志，并在结束时执行 `docker compose down -v` 清理 CI 专用资源。

## 安全与可靠性边界

- 所有正式凭证只允许通过 GitHub Actions Secrets 或生产环境注入；本工作流不读取或输出它们。
- 工作流日志不得打印 `.env` 内容、令牌、验证码、完整手机号或精确用户位置。
- CI 使用伪供应商，不调用真实高德、阿里云短信或任何外部写接口。
- Docker 镜像和 Compose 构建必须使用仓库内容，不依赖开发者 Windows 缓存。

## 非目标

- 不发布镜像、不部署阿里云、不创建数据库备份或生产基础设施。
- 不替代高德类型码核验、合法营业时间数据源、OpenTelemetry 接入或 Android 真机联调。
- 不提交本地 `backend/.env`、下载的工具链、缓存或用户凭证。
