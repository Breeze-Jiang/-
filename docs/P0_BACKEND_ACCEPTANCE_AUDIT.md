# 沿途 P0 后端验收审计

> 状态：进行中。本文记录当前实现与总规划的可验证证据；未列为“已验证”的项目不得作为发布通过依据。

## 范围与事实来源

- 后端唯一实现目录：`backend/`。
- 业务、边界与优先级以 `docs/ALONGTU_MASTER_PLAN.md` 第 8、9、13、14 节为准。
- `backend/api/openapi.yaml` 是 P0 客户端联调的唯一接口契约。
- 工作区里旧 FastAPI 路径或旧客户端代码不属于本后端的兼容目标。

## 已实现并具备本地静态证据

| 验收项 | 当前实现与证据 |
| --- | --- |
| 游客找点闭环 | `/api/v1/places/nearby`、`/search`、`/{id}`、路线与外部导航路由已实现；`internal/app/visitor_e2e_test.go` 覆盖附近厕所、详情、路线、外部导航和输入失败。 |
| 距离与坐标 | PostGIS 仅保存/计算 WGS84；地点展示返回已保存的 GCJ-02；供应商周边结果由客户端重新计算距离、过滤半径外点，并在合并前按距离排序。 |
| 可信信息 | 地点存在、营业信息确认与匿名来源摘要已实现；入口/错误信息只作为追加的事实反馈进入纠错审核，不进入公开可信摘要。确认/反馈均支持幂等键、账号/会话状态校验。 |
| 认证安全 | 短信 challenge、设备绑定、Redis 频控、失败次数、刷新令牌单次轮换及重放撤销已实现。 |
| 供应商可靠性 | 2.2 秒总截止时间、一次有限重试、五次失败熔断、Redis 新鲜/过期缓存和明确 `live`/`cached`/`degraded` 状态已实现；未核验类别返回 `coverage: not_covered` 空结果。 |
| 导航边界 | 步行、骑行、驾车、公交地铁、打车入口统一建模；高德、百度、腾讯外部链接均明确坐标系统，高德 GCJ-02 链接使用 `dev=1`。 |
| 运行安全 | 请求 ID、严格 JSON、请求大小限制、CORS 白名单、双监听健康检查、JSON 日志、Prometheus 指标与优雅停止已实现。 |
| 镜像安全 | Dockerfile 只复制编译所需的 `cmd/` 与 `internal/`；`.dockerignore` 排除环境文件、缓存、二进制与常见私钥文件。 |

本机已实际运行并通过：

- `gofmt` 检查；
- `go vet ./...`；
- `go build ./cmd/api`；
- `go test ./...`（本机一次完整运行通过；单独运行新生成测试二进制仍可能被 Windows Application Control 拦截）；
- `npx --yes @stoplight/spectral-cli lint api/openapi.yaml`；
- sqlc 1.29.0 生成。

2026-07-16 补充：打车模式已按总规划收口为仅第三方叫车跳转，`navigation.Service` 不会向路线供应商请求或返回伪装的应用内打车路线；`TestTaxiReturnsOnlyConfiguredExternalEntryWithoutCallingRouteProvider` 已通过。另已修复供应商失败、本地地点降级时遗漏 OpenAPI 必填 `coverage` 字段的问题；`TestNearbyFallsBackToStoredData` 与完整 `go test ./...` 已通过。

同日补充：附近与搜索游标现在以 HMAC 绑定查询指纹、供应商页码、最后内部地点 ID，以及在距离排序可用时的最后距离；下一页会过滤重复边界项和已越过的距离项。`TestNearbyCursorDropsRepeatedProviderBoundaryItem`、`TestSearchCursorDropsRepeatedProviderBoundaryItem`、游标篡改/过期与旧页码游标兼容测试，以及完整 `go test ./...` 已通过。实时供应商在游标有效期内仍可能发生外部数据变动，响应的 `dataFreshness` 将继续如实标明 live/cached/degraded，而不会伪称为静态快照。

同日补充：认证输入错误（手机号/设备 ID）返回稳定的 400，不再误报短信服务不可用；验证码挑战存储（Redis）故障返回 503，数据库或其他内部错误返回 500，不再伪装成验证码无效 401。OpenAPI 的验证码校验错误响应与 `TestSendRejectsInvalidClientInputWithoutClaimingSMSOutage`、`TestVerifyReportsChallengeStoreFailureAsServiceUnavailable`、完整 `go test ./...` 已同步通过。

同日补充：刷新令牌与当前会话注销已区分“会话无效/重放”的 401 和持久化层故障的 500；会话查询只在 `pgx.ErrNoRows` 时判定令牌无效。OpenAPI 与 `TestSessionStorageFailureIsNotReportedAsInvalidRefreshToken`、完整 `go test ./...` 已同步通过。

同日补充：受保护写操作的访问令牌校验已区分会话状态查询失败的 500 与失效/撤销/伪造令牌的 401；`TestWriteAuthenticationStorageFailureIsNotReportedAsInvalidToken` 与完整 `go test ./...` 已通过。

同日补充：OpenAPI 已明确认证、刷新、注销和路线请求拒绝未知 JSON 字段，与服务端严格解码行为一致；契约测试会检查该约束，Spectral 与完整 `go test ./...` 已通过。

同日补充：附近结果在距离相同（或均缺失距离）时按内部地点 ID 稳定排序，与签名游标的距离/ID 边界保持一致；`TestMergeOrdersEqualDistancesByPlaceID` 与完整 `go test ./...` 已通过。

同日补充：附近查询的下一页游标从稳定排序后的最大“距离 + 内部地点 ID”边界生成，不依赖供应商数组原始顺序；`TestNearbyCursorUsesStableTieBreakerForProviderBoundary` 与完整 `go test ./...` 已通过。

同日补充：供应商层将原生 `context.DeadlineExceeded` 识别为可重试失败并计入熔断，连续五次超时后打开熔断；`TestDeadlineFailuresOpenCircuit` 与完整 `go test ./...` 已通过。

同日补充（Docker Compose 本地验收）：Windows Docker Desktop + WSL 2 环境已实际完成 Compose 构建与启动。由于 `ghcr.io/pressly/goose:3.24.1` 在当前网络被拒绝匿名拉取，迁移服务已改为本仓库 Dockerfile 的本地 `migrate` target；该 target 固定构建 `github.com/pressly/goose/v3/cmd/goose@v3.24.1`，并实际输出 `goose version: v3.24.1`。`docker compose up -d` 的结果为 PostGIS、Redis 健康，伪供应商和 API 运行，`migrate` 以 0 退出；`goose_db_version` 的版本 1、2 均为已应用，`postgis` 与 `pgcrypto` 扩展存在，`/health/live`、`/health/ready` 均返回 200。游客冒烟闭环实际验证了厕所附近查询、平台地点详情（地点存在/营业时间均如实为 `unknown`）、应用内步行路线（1200 米/900 秒）以及 3 个外部导航入口。此证据仅证明本地 Docker 环境可用，不替代 Linux CI。

## 发布前必须由 Linux CI 证明

- `go test ./...` 与 `go test -race ./...`；
- 真实 PostGIS/Redis 容器的迁移和集成测试；
- Docker Compose 游客闭环 E2E；
- Docker 镜像构建；
- sqlc 生成文件相对已提交基线的无差异检查。

本机不能替代以上验证：Windows Application Control 会阻止新生成的 Go 测试二进制执行；虽然本机现已完成真实 Docker Compose 验收，仍不能替代 Linux CI 的独立可复现证明；当前仓库尚无首次提交，无法用 `git diff` 作为生成文件基线。

## 产品与供应商待决项

1. 停车场、充电宝、药店、医院、便利店、警务点、美食缺少正式高德类型码。它们当前明确返回 `coverage: not_covered` 空结果；取得官方资料并完成适配层测试前，不得宣称“完整类别筛选已验收”。
2. 地点详情营业时间目前统一为未知/待确认。仅在取得合法授权的数据源后，才可保存来源、更新时间与结构化周营业时段。
3. OpenTelemetry 依赖下载受当前网络到 Go 代理的限制而未接入；不得以伪实现替代。联网 CI 或受控依赖镜像可完成该项。

## 当前结论

P0 后端处于“实现与自测（G4）”阶段，尚未达到“联调与验收（G5）”。在 Linux CI 通过、供应商类型码与产品待决项关闭前，不得标记为可发布。
