# 00 - 项目总览

> 状态：规范基线（实现唯一来源）
> 适用：所有后端实现与 AI 编码代理（约束见根目录 `AGENTS.md`，本地忽略不入库）

---

## 1. 项目定位

**e-fiber-admin** 是一个 **API-first 的外贸独立站后台系统**，用于快速建立适用于「企业站」与「商城」两类形态的独立站点。

- 参考对象：
  - **WooCommerce**：功能完备、插件生态丰富、企业站/商城兼顾、CMS 能力强。
  - **MedusaJS**：headless、API-first、模块化、多区域/多币种抽象清晰。
- 取舍原则：**不照搬**，结合两者特点做减法与重组，目标是「易用、可快速建立、可维护」的后台，而非功能全集。
- 交付范围（本期）：**仅后端 API + 规范设计文档**。管理前端（React/Vue）不在本期范围，后续以独立工程对接本 API。

---

## 2. 核心目标

| # | 目标 | 说明 |
|---|---|---|
| G1 | 快速建站 | 单租户一仓一部署，开箱即用，最小配置上线 |
| G2 | 企业站 + 商城统一 | 一套数据模型 + 站点级功能开关，按形态裁剪行为与 API |
| G3 | 外贸刚需内置 | 多语言、多币种、多区域、询盘（B2B lead）、邮件通知开箱即用 |
| G4 | API-first / headless | 前后端分离，可对接 Web/APP/ERP/导出等多渠道 |
| G5 | 可扩展 | 内置模块 + 事件钩子 + Provider 接口，预留插件位（不做市场） |
| G6 | 可维护 | ent schema 即代码、Atlas 迁移、统一错误码、规范提交 |

---

## 3. 非目标（明确不做）

- ❌ 多租户 SaaS 平台（本期单租户；数据层预留 `tenant_id` 扩展位但不实现隔离逻辑）
- ❌ 管理后台前端 UI（本期仅 API）
- ❌ 插件市场 / 应用商店
- ❌ 复杂营销玩法（拼团/分销/秒杀）——预留事件钩子，不内置实现
- ❌ 实时竞价 / 复杂价格引擎（按币种独立存价，不做规则引擎）
- ❌ 自建搜索引擎（MVP 用 PostgreSQL `tsvector`，预留 Meilisearch 接口）

---

## 4. 术语表

| 术语 | 含义 |
|---|---|
| **Site / Store** | 单个独立站，由 `store` 单条记录表示，持站点类型与功能开关 |
| **site_type** | 站点形态：`corporate`（企业站）/ `store`（商城）/ `hybrid`（混合） |
| **feature_flags** | 站点级功能开关：`enable_cart` `enable_checkout` `enable_inquiry` `enable_blog` 等 |
| **Product** | 商品主对象，持公共属性与可翻译内容 |
| **Variant** | 商品规格（SKU/库存/重量/价格的载体），simple 类型为单 variant |
| **Product Type** | `simple` / `variable` / `virtual` / `grouped`（Woo 四类，统一用 Variant 表达） |
| **Region** | 区域 = 语言 + 币种 + 税率 + 可用支付/物流的组合（简化 Medusa 模型） |
| **Money** | `(amount, currency_code)` 二元组，金额以最小货币单位整数存储 |
| **i18n / Translation** | 可翻译字段拆主表 + 翻译表 `(entity_id, locale)` |
| **Inquiry** | 询盘：B2B lead generation 的表单提交与跟进记录 |
| **Admin User** | 后台运营/管理员账号，走 RBAC |
| **Customer** | 前台消费者/询盘客户账号 |
| **Event Bus** | 基于 Redis Pub/Sub 的进程内事件总线，解耦领域副作用 |
| **Job** | 基于 asynq（Redis）的异步任务 |

---

## 5. 技术栈速览

| 层 | 选型 |
|---|---|
| 语言 | Go 1.22+ |
| Web 框架 | GoFiber v3 |
| 数据库 | PostgreSQL 16 |
| 缓存 / 总线 / 队列后端 | Redis 7 |
| ORM / 迁移 | **ent + Atlas**（schema 即代码） |
| 异步任务 | asynq（Redis 后端） |
| 鉴权 | JWT（access + refresh）+ RBAC |
| 对象存储 | S3 兼容（MinIO 开发 / 任 S3 生产） |
| 文档 | OpenAPI 3.1 |
| 日志 | `log/slog` |
| 指标 | Prometheus |
| 全文检索 | PostgreSQL `tsvector`（MVP），预留 Meilisearch |

> 详见 `docs/01-architecture.md`、`docs/02-tradeoff-woo-medusa.md`。

---

## 6. 文档地图

| 文档 | 主题 |
|---|---|
| `00-overview.md` | 本文件：定位、目标、术语、范围 |
| `01-architecture.md` | 技术架构、分层、项目结构、部署拓扑 |
| `02-tradeoff-woo-medusa.md` | WooCommerce vs MedusaJS 取舍决策矩阵 |
| `03-data-model.md` | ER 设计、核心实体、关系、i18n/多币种/审计策略 |
| `04-api-design.md` | REST 约定、鉴权、错误码、分页、版本、OpenAPI 规范 |
| `05-modules.md` | 八大模块详细设计（职责/实体/接口/流程/扩展点） |
| `06-auth-rbac.md` | JWT 双 Token、RBAC、管理员/客户两套体系 |
| `07-events-jobs.md` | 事件总线、Subscriber、asynq 任务、关键事件清单 |
| `08-roadmap.md` | 实现里程碑、MVP 优先级、验收标准 |
| `adr/0001~` | 关键架构决策记录 |

---

## 7. 部署形态

**单租户一仓一部署**：
- 一个后台进程（`cmd/admin`）+ 一个 worker 进程（`cmd/worker`）对应一个商户站点。
- 中间件（PostgreSQL / Redis / MinIO）统一通过 `docker compose` 拉取，禁止本机安装（见 `AGENTS.md` §1）。
- 「快速建立」由「最小配置 + 一键 `make up && make api`」实现，而非 SaaS 多租户隔离。

---

## 8. 阅读顺序建议

1. `00-overview.md` → 2. `01-architecture.md` → 3. `02-tradeoff-woo-medusa.md`
2. `03-data-model.md` → 4. `04-api-design.md` → 5. `05-modules.md`
3. `06-auth-rbac.md` → 6. `07-events-jobs.md` → 7. `08-roadmap.md`
4. 遇到关键决策回查 `adr/`。
