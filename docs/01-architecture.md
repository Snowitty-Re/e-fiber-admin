# 01 - 技术架构

> 状态：规范基线
> 与 `AGENTS.md` §4.1 技术栈固定约束一致，禁止擅自替换选型。

---

## 1. 架构总览

```
                        ┌──────────────────────────────────────────┐
                        │            客户端（前端 / APP / ERP）        │
                        └───────────────────┬──────────────────────┘
                                            │ HTTPS / REST / OpenAPI 3.1
                                            ▼
        ┌───────────────────────────────────────────────────────────────┐
        │                     cmd/admin  (GoFiber v3)                     │
        │  Router → Middleware → Handler → Service → Repository → ent     │
        │            ↑                        ↑                ↑          │
        │   JWT/RBAC/RateLimit/Recover    事件发布        PostgreSQL      │
        │   RequestID/Logger/CORS         (Redis Pub/Sub)   事务/查询      │
        └───────────────────────┬───────────────────────────┬───────────┘
                                │ 发布事件                    │
                                ▼                            ▼
        ┌──────────────────────────────────┐   ┌─────────────────────────┐
        │      cmd/worker (asynq)            │   │   PostgreSQL 16          │
        │  Subscriber → 邮件/库存/统计/积分   │   │   (ent + Atlas 迁移)      │
        │  定时任务 / 批量任务 / 重试         │   └─────────────────────────┘
        └───────────────┬────────────────────┘   ┌─────────────────────────┐
                        │                        │   Redis 7                │
                        └───────────────────────▶│  缓存/限流/RefreshToken/ │
                                                  │  事件总线/asynq 队列      │
                                                  └─────────────────────────┘
                                                  ┌─────────────────────────┐
                                                  │   MinIO / S3 兼容         │
                                                  │   媒体对象存储             │
                                                  └─────────────────────────┘
```

- 两个可执行：`cmd/admin`（同步 HTTP API）、`cmd/worker`（asynq 异步任务消费者）。
- 两者共享 `internal/` 全部领域代码（service/repository/ent/events），仅入口与运行方式不同。
- 中间件全部 Docker 化，禁止本机安装（`AGENTS.md` §1）。

---

## 2. 分层架构与职责

自外向内，单向依赖（外层依赖内层，内层不反向依赖）：

| 层 | 路径 | 职责 | 规则 |
|---|---|---|---|
| **入口** | `cmd/{admin,worker}` | 进程 bootstrap：装配依赖、启动 Fiber / asynq | 不含业务逻辑 |
| **传输** | `internal/http/fiber/{router,middleware,handler,dto}` | 路由注册、中间件、HTTP↔DTO 转换、参数校验 | handler 不直接操作 repository |
| **领域** | `internal/domain/*` | 聚合根、实体、值对象、领域服务、领域事件定义 | 不依赖 Fiber / DB 具体 impl |
| **应用服务** | `internal/service` | 用例编排：跨 repository、事务边界、事件发布 | 唯一事务边界出口 |
| **持久化** | `internal/repository` | ent 查询封装、仓储接口实现 | 不含业务规则 |
| **ORM** | `internal/ent` | ent schema + 生成代码 | 仅 schema 手写，代码生成 |
| **基础设施** | `internal/{config,database,events,jobs}` | 配置、DB 连接、Redis、事件总线、asynq 装配 | 可被各层依赖的最底层 |
| **公共库** | `internal/pkg/{auth,pagination,i18n,errors,money}` | 纯工具，无状态 | 任意层可引用 |

### 硬规则
- **跨域禁止直接操作对方 repository**：`domain/product` 不得 import `domain/order` 的 repository；跨域协作只能经 `service` 编排或事件。
- **事务边界只在 `service` 开启**：repository 接收事务客户端（`ent.Tx` 或上下文事务），自身不开事务。
- **领域层不 import Fiber**：保持领域可测试、可被 worker 复用。
- **handler 只做 DTO↔领域转换 + 调 service**：禁止在 handler 写业务分支。

---

## 3. 项目结构

```
e-fiber-admin/
├── cmd/
│   ├── admin/main.go            # API 进程
│   └── worker/main.go           # asynq worker 进程
├── internal/
│   ├── config/                  # 配置加载（.env → 结构体）
│   ├── database/                # PG / Redis / S3 连接装配
│   ├── ent/                     # ent schema + 生成代码
│   │   └── schema/              # 手写 schema（唯一数据模型真源）
│   ├── http/fiber/
│   │   ├── router/              # 路由注册、路由分组
│   │   ├── middleware/          # JWT/RBAC/RateLimit/Recover/RequestID/Logger/CORS
│   │   ├── handler/             # HTTP handler
│   │   └── dto/                 # 请求/响应 DTO + 校验
│   ├── domain/                  # 八大领域
│   │   ├── auth/  settings/ media/  product/ customer/ region/
│   │   ├── order/ payment/ shipping/ discount/ cms/ inquiry/
│   │   └── <each>: entity.go  service.go  event.go  repo.go(interface)
│   ├── service/                 # 应用服务（用例编排）
│   ├── repository/              # ent 仓储实现
│   ├── events/                  # Redis 事件总线 + 路由
│   ├── jobs/                    # asynq 任务定义 + handler
│   └── pkg/
│       ├── auth/                # JWT 签发/校验
│       ├── pagination/          # 分页解析与响应
│       ├── i18n/                # locale 解析、翻译表加载
│       ├── errors/              # 统一错误码与 AppError
│       └── money/               # Money 值对象（amount + currency）
├── api/openapi/                 # OpenAPI 3.1 规范文件
├── migrations/                  # Atlas 迁移文件
├── docs/                        # 设计文档（交付物，已提交）
├── docker-compose.yml           # 中间件（禁止本机安装）
├── Makefile                     # 开发与基础设施命令
├── atlas.hcl                    # Atlas 迁移配置
├── go.mod
└── .env.example                 # 环境变量模板
```

> `AGENTS.md`、`.opencode/` 等 AI 指令文件本地忽略，不入库。

---

## 4. 请求处理流（典型）

```
HTTP Request
  → middleware: Recover → RequestID → Logger → CORS → RateLimit
  → middleware: JWTAuth（解析 access token → 注入 admin context）
  → middleware: RBAC（校验 permission）
  → handler: 解析 DTO + 参数校验 → 调 service
  → service: 开事务 → 调若干 repository → 发布领域事件 → 提交事务
  → events: 同步广播到 Redis Pub/Sub（事务提交后发，保证一致性）
  → worker: Subscriber 消费事件 → 触发 asynq Job（邮件/库存/统计）
  → handler: 组装响应 DTO → JSON
```

> 事件在**事务提交后**发布（outbox 或 post-commit hook），避免「事件已发但事务回滚」的不一致。MVP 用 post-commit；高一致性场景可升级 outbox（见 `docs/07-events-jobs.md`）。

---

## 5. 配置管理

- 来源：`.env`（开发）/ 环境变量（生产）。`.env.example` 入库为模板，`.env` 忽略。
- 加载：`internal/config` 用 `env` 标签映射到强类型结构体，启动时一次性校验必填项。
- 注入：依赖注入到 service/repository，**禁止全局单例读取配置**（除 logger）。
- 安全：密钥（JWT secret / DB 密码）禁止硬编码、禁止打印、禁止入日志（`AGENTS.md` §5）。

---

## 6. 可观测性

| 维度 | 方案 |
|---|---|
| 日志 | `log/slog`，结构化 JSON，带 `request_id` / `admin_id` / `module` |
| 指标 | Prometheus `/metrics`：HTTP 延迟/状态码、DB 池、asynq 队列深度、事件处理耗时 |
| 链路 | 请求级 `X-Request-ID` 透传；后续可接 OpenTelemetry（非 MVP） |
| 健康检查 | `/healthz`（进程存活）、`/readyz`（DB/Redis 依赖就绪） |

---

## 7. 部署拓扑（单租户一仓一部署）

```
┌──────── 单机 / 单节点 ────────┐
│  docker compose up -d         │
│   ├─ postgres:16              │
│   ├─ redis:7                  │
│   └─ minio                    │
│  cmd/admin   (systemd / pm2)  │
│  cmd/worker (systemd / pm2)   │
└───────────────────────────────┘
```

- MVP 单节点即可；架构允许 admin/worker 水平扩多副本（无状态，状态在 PG/Redis）。
- 生产对象存储可换任 S3 兼容（AWS S3 / R2 / 阿里 OSS），通过 `internal/database` 的 S3 client 抽象切换。

---

## 8. 关键设计约束（与 AGENTS.md 联动）

- 技术栈固定，禁止擅自替换（`AGENTS.md` §4.1）。
- 提交前 `go fmt ./...` + `go vet ./...` + `go build ./...`（`AGENTS.md` §4.4）。
- 不添加注释，除非用户明确要求（`AGENTS.md` §4.2）。
- 实现严格对应 `docs/03` 数据模型 / `docs/04` API / `docs/05` 模块边界。
