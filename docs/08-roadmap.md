# 08 - 实现路线图

> 状态：规范基线
> 实现顺序以本文件里程碑为准（`AGENTS.md` §4.3）。每个里程碑内按「数据→服务→接口→测试」原子提交（`AGENTS.md` §2）。
> 本期范围：仅后端 API + 设计文档（不含前端 UI）。

---

## 里程碑总览

| 里程碑 | 主题 | 价值 | 验收（能做什么） |
|---|---|---|---|
| **M0** | 基座 | 一切依赖 | 启动服务、登录、健康检查、迁移、OpenAPI 脚手架 |
| **M1** | 产品 + 区域 + CMS | **企业站可上线** | 多语言多币种产品目录 + 落地页/博客/菜单 |
| **M2** | 客户 + 询盘 + 通知 | **企业站完整** | lead generation 闭环（询盘→邮件通知→跟进） |
| **M3** | 交易链路 | 商城核心 | 购物车→结账→订单→履约→退换 |
| **M4** | 支付 + 物流 + 优惠 | 商城完整 | 真实支付/运费/券，端到端下单成交 |
| **M5** | 事件 + 任务 + Webhook + 可观测 | 运营与扩展 | 全链路事件驱动、第三方集成、指标告警 |

> **关键节点**：M1 完成即可对外提供「企业站」MVP；M3 完成提供「商城」MVP。M5 完成进入可运营态。

---

## M0 — 基座

### 目标
可启动、可登录、可迁移、可文档化的最小工程。

### 任务（逐项原子提交）
1. `config` 加载（.env→结构体，强校验）+ `database`（PG/Redis/S3 连接装配）。
2. ent mixin（BaseMixin：id/时间戳/软删/审计/版本/tenant_id）+ 空骨架 `make ent-gen` 可跑。
3. Atlas 迁移初始化 + `migrate-apply` 可跑（建库基线）。
4. Fiber app 装配：Router + 中间件链（Recover/RequestID/Logger/CORS/RateLimit/healthz/readyz）。
5. `pkg/errors`：AppError + 错误码注册 + 统一错误中间件。
6. `pkg/auth`：JWT 签发/校验；`pkg/pagination`、`pkg/money`、`pkg/i18n`。
7. auth 域：admin_user/role/permission ent schema + seed（权限全集 + 系统角色）。
8. admin 登录/刷新/登出/me 接口 + JWTAuth/RBAC 中间件。
9. OpenAPI 脚手架：`api/openapi/openapi.yaml` + errors.yaml + 鉴权 schema。
10. slog 结构化日志 + Prometheus `/metrics`。

### 验收
- `make up && make migrate-apply && make api` 启动成功。
- `POST /api/v1/admin/auth/login` 返回 token；`GET /admin/auth/me` 正常。
- `GET /healthz` 200、`/readyz` 依赖就绪 200。
- RBAC：无权限访问 `product:write` 路由返回 403。

### 不含
- 任何业务实体（产品/订单等）。

---

## M1 — 产品 + 区域 + CMS（企业站可上线）

### 目标
多语言、多币种产品目录 + CMS 落地页，企业站展示完整。

### 任务
1. region/locale/currency/tax_rate schema + service + admin 接口。
2. product/variant/variant_price schema + service（多币种存价）+ admin 接口。
3. product_option/option_value/variant_option_value 规格。
4. category/collection/tag + product_media 媒体关联。
5. media 域：S3 上传/读取/删除 + admin 接口。
6. product/category/collection translation（i18n）。
7. storefront 只读接口：`/store/products`、`/store/products/{slug}`（按 locale+currency 展示价）。
8. CMS：page/blog_post/block/menu + translation + admin/storefront 接口。
9. settings/store + feature_flags 读写接口。
10. PG tsvector 全文检索（产品）+ `/store/products?q=`。

### 验收
- 可创建多语言产品，按币种独立存价，storefront 按语言/币种返回正确内容与价格。
- 企业站形态（site_type=corporate，关闭 cart/checkout）下，交易相关端点不暴露/返回 409。
- 可建落地页/博客/菜单，storefront 按 slug+locale 渲染。
- 全文检索可用。

---

## M2 — 客户 + 询盘 + 通知（企业站完整）

### 目标
lead generation 闭环：客户注册/询盘提交/邮件通知/跟进/转单预留。

### 任务
1. customer/customer_address/customer_group schema + service + storefront 注册登录 + admin 接口。
2. customer JWT（独立密钥/命名空间）+ storefront 鉴权中间件。
3. form_definition + translation schema + service + admin 接口。
4. inquiry schema + service + storefront 提交接口（含 guest）。
5. events bus 基础实现（Redis Pub/Sub）+ asynq 基础（client/server/worker 启动）。
6. email_template + translation + notification schema + service。
7. Subscriber：`inquiry.received` → 通知邮件任务（客户 + 站点）。
8. admin 询盘管理：列表/分配/状态流转 + `inquiry.updated/assigned` 事件。
9. inquiry 转 cart/checkout（转单）接口（依赖 M3 cart/checkout，可先 stub，M3 补全）。

### 验收
- 客户可注册登录；guest 可提交询盘。
- 询盘提交后客户与站点均收到邮件（asynq 任务发）。
- admin 可查看/分配/流转询盘。
- 事件总线 + asynq worker 正常运行。

---

## M3 — 交易链路（商城核心）

### 目标
购物车→结账→订单→履约→退换完整链路。

### 任务
1. cart/cart_item schema + service + storefront 接口。
2. checkout schema + service + storefront 分步接口（地址/物流/支付占位/完成）。
3. order/order_item schema + service + admin 接口 + storefront 订单查看。
4. fulfillment/fulfillment_item + return/return_item + swap。
5. order 状态机（pending/paid/fulfilled/cancelled/refunded）+ cancel/fulfill/return 动作接口。
6. 事件：`order.placed/paid/fulfilled/cancelled/returned` + `cart.abandoned`。
7. 库存扣减/回滚（事件驱动 + 事务内同步扣减二选一，MVP 同步扣减 + 事件通知）。
8. Abandoned Cart 延时召回任务。
9. 补全 M2 inquiry→order 转单。

### 验收
- storefront 完整下单流程跑通（不含真实支付，支付占位）。
- order 状态机正确流转，库存正确扣减/回滚。
- 退换流程可用。
- Abandoned Cart 召回任务触发。

---

## M4 — 支付 + 物流 + 优惠（商城完整）

### 目标
真实支付、运费、优惠使端到端成交可用。

### 任务
1. payment_provider 接口化 + 内置占位 + Stripe 占位实现。
2. payment_session/transaction + authorize/capture/refund + webhook 处理。
3. shipping_profile/shipping_option + 报价接口（Quote）。
4. product_shipping_profile 关联 + storefront 物流选项选择。
5. discount/coupon/discount_rule/discount_condition + 校验/应用。
6. checkout 集成支付会话 + 优惠码 + 物流选项 → totals 聚合。
7. 事件：`payment.*`、`shipping_option.quoted`、`discount.redeemed/expired`。
8. 优惠定时过期任务。

### 验收
- 可配置一个支付 provider 跑通 authorize/capture/refund（沙箱）。
- storefront 结账时返回正确物流选项与运费。
- 优惠码可用，规则（百分比/固定/包邮/买X送Y）正确应用。
- totals（subtotal/discount/tax/shipping/total）计算正确。

---

## M5 — 事件 + 任务 + Webhook + 可观测（运营与扩展）

### 目标
全链路事件驱动、第三方集成、指标告警，进入可运营态。

### 任务
1. 全域事件补齐 + Subscriber 注册中心。
2. Webhook 出口：dispatcher + 签名 + 重试 + event_log。
3. outbox 升级（可选，按一致性需求）：event_outbox + relay。
4. 幂等去重（Redis dedup + notification 去重窗口）。
5. 死信处理 + 后台重投。
6. Prometheus 指标全量 + 告警规则（死信/通知连续失败/webhook 连续失败）。
7. OpenAPI 补齐全部端点 + errors.yaml 完整。
8. 导出任务（订单/询盘 CSV）asynq 实现。
9. 缩略图生成、CDN purge、RSS 等次要任务。

### 验收
- 关键事件均有消费者且幂等。
- Webhook 订阅可成功接收带签名事件。
- 指标 dashboard 可观测关键链路。
- 导出任务可用。
- OpenAPI 与实现一致（lint 通过）。

---

## 风险与对策

| 风险 | 影响 | 对策 |
|---|---|---|
| ent + Atlas 学习曲线 | M0/M1 进度 | 先做最小 schema 跑通生成+迁移，再扩 |
| 多币种定价复杂 | M1 | MVP 仅按币种独立存价，不做 price_list 规则引擎 |
| 支付 provider 集成 | M4 | 接口化 + 占位，按需接真实网关，不阻塞商城 MVP |
| 事件一致性 | M5 | MVP post-commit，预留 outbox，按需切换 |
| 范围蔓延 | 全程 | 严守非目标清单（`docs/00` §3），复杂玩法留事件钩子不内置 |

---

## 提交规范提醒

- 每个任务项 = 一个原子 commit（`AGENTS.md` §2.1）。
- Conventional Commits：`feat(product): add variant price by currency` 等。
- 提交前 `go fmt ./...` + `go vet ./...` + `go build ./...`（`AGENTS.md` §4.4）。
- 不主动 push / 开 PR（`AGENTS.md` §2.4）。

---

## 验收总则

- 每个 milestone 完成 → 跑 `make test`（单测 + 集成）+ `make build` + 手动冒烟对应「验收」清单。
- 数据模型/API/模块边界实现与 `docs/03/04/05` 一致；偏离须记 ADR。
