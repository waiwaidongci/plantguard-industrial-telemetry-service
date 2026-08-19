# PlantGuard

项目编号：01  
模块路径：`github.com/acme/plantguard`

PlantGuard 是一个纯 Go 企业级工业设备遥测与维护编排平台。它面向工厂、能源站和园区设备，负责设备档案、遥测接入、心跳、规则触发事件、维护计划与任务编排，并通过 REST API 向外部系统提供多租户服务。

## Architecture

```text
cmd/api                 HTTP API 入口
cmd/worker              后台 worker 入口
api/openapi.yaml        接口说明
configs/config.yaml     默认 YAML 配置
migrations/sqlite       本地开发 SQLite 迁移
migrations/postgres     PostgreSQL 迁移
deploy                  PostgreSQL Compose 与容器示例
scripts                 本地启动、迁移、统计脚本
internal/tenant         租户与站点领域
internal/device         设备与设备型号领域
internal/telemetry      遥测批次、汇总与校验领域
internal/rule           规则与阈值求值领域
internal/event          事件与确认领域
internal/maintenance    维护计划与任务领域
internal/notification   可插拔通知适配层
internal/shared         跨领域错误、分页、HTTP、基础设施
```

每个核心领域按 `domain -> application -> adapter -> infrastructure` 分层。应用层只依赖接口，具体仓储通过构造函数注入。

## Local Startup

默认使用 `modernc.org/sqlite`，无需 Docker 或本地 PostgreSQL：

```bash
cd 01-plantguard
./scripts/run-dev.sh
```

服务默认监听 `:8080`。若端口被占用，可覆盖：

```bash
PLANTGUARD_SERVER_ADDRESS=:18081 ./scripts/run-dev.sh
```

健康检查：

```bash
curl http://127.0.0.1:18081/healthz
curl http://127.0.0.1:18081/readyz
curl http://127.0.0.1:18081/metrics
```

## Configuration

配置先读取 `configs/config.yaml`，再由环境变量覆盖。环境变量前缀为 `PLANTGUARD`，层级用 `_` 表示，例如：

| 环境变量 | 说明 |
| --- | --- |
| `PLANTGUARD_SERVER_ADDRESS` | HTTP 监听地址 |
| `PLANTGUARD_DATABASE_DRIVER` | `sqlite` 或 PostgreSQL 驱动 |
| `PLANTGUARD_DATABASE_DSN` | 数据库连接串 |
| `PLANTGUARD_DATABASE_MIGRATIONS_DIR` | PostgreSQL 迁移目录 |
| `PLANTGUARD_LOGGING_FORMAT` | `json` 或 `text` |
| `PLANTGUARD_WORKER_INTERVAL` | worker 扫描间隔 |

## Database Migration

SQLite 迁移会在启动时自动执行。PostgreSQL 使用：

```bash
./scripts/migrate-postgres.sh
```

也可通过 Compose 启动 PostgreSQL：

```bash
docker compose -f deploy/docker-compose.postgres.yml up -d
```

## Example Requests

```bash
curl -X POST http://127.0.0.1:18081/api/v1/tenants \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: tenant-1' \
  -d '{"name":"Factory A","slug":"factory-a"}'
```

详细接口说明见 `api/openapi.yaml`。创建租户、站点、设备型号、设备、遥测批次、规则、事件和维护计划后，可用列表接口和健康端点验证。

## Scripts and Makefile

```bash
make fmt
make test
make vet
make build
make run-dev
make count
```

`scripts/run-dev.sh` 是本地开发验证入口，`scripts/migrate-postgres.sh` 执行 PostgreSQL 迁移，`scripts/count-go.sh` 统计非测试 Go 文件数和非测试 Go 源码行数。

## Verification

提交前应执行：

```bash
go mod tidy
gofmt -w .
go test ./...
go vet ./...
go build ./...
./scripts/count-go.sh
```
