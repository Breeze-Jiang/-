# Alongtu backend

沿途 P0 后端是一个按业务能力拆分的 Go 模块化单体。`backend/` 与总规划文档是当前后端唯一事实来源；不读取旧 SQLite 数据，也不兼容旧 API。

`api/openapi.yaml` 是客户端联调的唯一 API 契约。工作区中仍可能存在使用旧 FastAPI 路径（例如 `/api/pois/*`）的 React Native 文件；它们不是 P0 后端的兼容目标，不能据此修改本服务的路径、字段或错误结构。

## 当前 P0 边界

- 游客地点搜索、附近地点和地点详情。
- 高德 POI 适配与供应商无关地点模型。
- 手机验证码最小会话、登录后地点确认，以及仅供纠错审核的入口/错误信息事实反馈。
- 步行、骑行、驾车、公交地铁路线展示，以及外部地图安全跳转参数。
- 可配置的第三方打车跳转；不承载订单、支付或调度。
- PostGIS、Redis、健康检查、Prometheus 指标和结构化日志基础。

语音逐向导航、偏航重算、收藏、标注、论坛、好友、聊天和实时位置不在此阶段。

## 本地启动

1. 安装 Go 1.24+、Docker 和 sqlc；如果使用 `make migrate-up`，还需安装 Goose。
2. 复制 `.env.example` 为本机环境变量，替换所有密钥。
3. 执行 `docker compose up -d postgres redis`。
4. 执行 `make migrate-up`、`make generate`、`make test`。
5. 执行 `make run`。

Docker Compose 会同时启动伪高德服务。CI 依次校验格式、sqlc 生成结果、单元/竞态测试、真实 PostGIS 集成测试、OpenAPI、Compose 游客闭环、`go vet` 与镜像构建。游客闭环使用 `go test -tags=e2e ./test/e2e`，验证 API、PostGIS、Redis 和伪供应商之间的找点、详情、路线与外部导航流程；本机可在 `docker compose up -d --build api` 后执行相同命令。若 Compose 阶段失败，CI 会输出所有 Compose 服务的状态和日志，便于定位迁移、依赖或 API 启动问题。

开发环境的模拟发送器不会输出验证码、完整手机号或令牌。生产环境使用阿里云官方短信 SDK，缺少 AccessKey、签名或模板时启动失败；自动化测试通过注入测试发送器读取验证码。

业务 API 默认监听 `:8080`；Prometheus 管理端默认仅监听 `127.0.0.1:9090`。容器环境通过内部网络暴露管理端口，不映射到公网。

## 稳定性约束

- GCJ-02 与 WGS84 必须显式标注，PostGIS geography 仅保存 WGS84。
- 供应商响应不得直接作为核心实体或公开 API。
- 未核验供应商分类返回 `coverage: not_covered` 的空结果，不得伪造成已分类地点或实时数据。
- Redis 数据可以丢失或重建；永久事实必须写入 PostgreSQL。
- 地点确认只能追加，重复提交通过 `Idempotency-Key` 返回原记录。
- 日志禁止输出访问令牌、刷新令牌、完整手机号和精确用户位置。
- 启动时 PostgreSQL 与 Redis 均必须在 5 秒内可连接；否则进程失败退出，由部署系统重试或告警。
