# 本地 Goose 迁移镜像设计

## 目标

让 `backend/docker-compose.yml` 在没有 GitHub Container Registry 凭证的开发与 Linux CI 环境中，仍可稳定执行数据库迁移，并保持“迁移成功后才启动 API”的依赖顺序。

## 已确认方案

`backend/Dockerfile` 增加一个命名为 `migrate` 的构建目标。该目标基于现有固定版本的 `golang:1.24-alpine` 构建阶段，使用 `go install github.com/pressly/goose/v3/cmd/goose@v3.24.1` 生成迁移二进制文件；运行阶段仅包含 Goose 二进制文件。

`backend/docker-compose.yml` 的 `migrate` 服务改为使用本仓库 Dockerfile 的 `migrate` target，并保留现有 Goose 命令、只读迁移挂载、PostGIS 健康检查依赖与 `service_completed_successfully` API 依赖。删除对 `ghcr.io/pressly/goose:3.24.1` 的运行时拉取依赖。

## 边界与安全

- 不新增 GitHub、Docker Hub、高德或阿里云凭证。
- 不修改迁移 SQL、数据库 schema、公开 API 或业务模块。
- `goose` 固定为 `v3.24.1`，避免构建时漂移到未验证版本。
- 生产仍应使用发布流水线构建的迁移镜像；本变更只移除开发/CI 对 GHCR 匿名拉取的依赖。

## 验证

1. `docker compose build migrate` 成功完成。
2. `docker compose up -d --build` 先启动 PostGIS，随后 `migrate` 成功退出，最后 API、Redis 与伪供应商处于运行状态。
3. 对 PostGIS 执行迁移状态检查，确认 `00001_p0_schema.sql` 和 `00002_p0_place_feedback.sql` 均已应用。
4. 请求 API 存活与就绪检查，确认服务已能依赖 PostgreSQL 和 Redis 正常工作。

## 非目标

- 不在本次修改中实现真实高德、阿里云短信或论坛/聊天功能。
- 不将开发 `.env` 或任何密钥提交到仓库。
