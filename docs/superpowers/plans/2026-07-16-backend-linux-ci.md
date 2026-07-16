# P0 后端 Linux CI 实施计划

> 执行时保持后端 P0 范围；不得提交 `.env`、真实密钥、缓存或未审查前端。

## 目标

在 GitHub Actions 的 Ubuntu 环境中验证 Go 质量、sqlc 生成、OpenAPI、Docker 镜像、PostGIS/Redis 迁移，以及游客找点闭环。

## 任务 1：创建工作流

文件：创建 `.github/workflows/backend-ci.yml`。

- [ ] 在 `push` 与 `pull_request` 的 `main` 分支上触发；路径限制为 `backend/**`、工作流自身和根 `.gitignore`。权限只给 `contents: read`。
- [ ] `quality` job 使用 `ubuntu-latest`、`actions/checkout@v4` 与 `actions/setup-go@v5` 的 Go `1.24.5`，并从 `backend/` 依次执行：

```bash
test -z "$(gofmt -l .)"
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0
sqlc generate
git diff --exit-code -- internal/platform/database/dbgen
go test ./...
go test -race ./...
go vet ./...
npm install --global @stoplight/spectral-cli@6.15.0
spectral lint api/openapi.yaml
docker build --tag alongtu-api-ci .
```

- [ ] `compose-e2e` job 依赖 `quality`，创建仅 CI 使用的 `backend/.env`（随机 JWT/Cursor 密钥、假高德 key、无阿里云凭证），再执行 `docker compose up -d --build`。
- [ ] 在 60 秒内轮询 `http://127.0.0.1:8080/health/ready`；超时时输出 `docker compose ps -a` 和 `docker compose logs --no-color` 后失败。
- [ ] 断言 `migrate` 退出码为 0，PostGIS 的 Goose 版本 1、2 已应用，`postgis`/`pgcrypto` 扩展存在；`/health/live` 和 `/health/ready` 返回 200。
- [ ] 用 Python 标准库请求伪供应商链路：厕所 nearby、返回地点详情、步行 routes、external-links；断言类别为 `toilet`、两项可信状态均为 `unknown`、路线距离/耗时为正、外部链接数为 3。
- [ ] 在 `if: always()` 步骤输出 Compose 状态与日志，并运行 `docker compose down -v`；不得输出 `.env` 内容。

## 任务 2：本地验证与首次提交

文件：使用 `.github/workflows/backend-ci.yml`、`backend/docker-compose.yml`。

- [ ] 用 Python YAML 解析器读取工作流，确认 YAML 有效。
- [ ] 在 `backend/` 运行 `docker compose config --quiet`，确认迁移服务仍构建 Dockerfile 的 `migrate` target。
- [ ] 检查暂存集合只包含 `.github/`、`backend/`、`docs/` 和根 `.gitignore`，并用 `git diff --cached --check` 检查空白错误。
- [ ] 在用户范围确认后执行：

```powershell
git add .github backend docs .gitignore
git commit -m "ci: add backend Linux validation"
git push -u origin main
```

## 任务 3：验证远程运行

- [ ] 运行 `gh run list --workflow "Backend Linux CI" --limit 1` 确认推送触发运行。
- [ ] 运行 `gh run watch --exit-status` 等待 `quality` 与 `compose-e2e`；若失败，只读取相关作业日志并修复最小原因后重新推送。

## 验收

- CI 不读取正式高德、阿里云或用户密钥。
- 两个 job 均通过，且 Docker Compose 验证基于干净的 Linux 运行器。
- 第一次提交不包含 `place_app/`、`design/`、`.claude/` 或 `start_frontend.bat`。
