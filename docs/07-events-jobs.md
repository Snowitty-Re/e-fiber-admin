# 07 - 事件与异步任务设计

> 状态：规范基线
> 实现约束：技术栈固定为 Redis Pub/Sub 事件总线 + asynq 任务队列（`AGENTS.md` §4.1）。

---

## 1. 总体模型

```
service（事务内）→ 写 DB
   └（事务提交后）→ events.Bus.Publish(event)
                        │
          ┌─────────────┴──────────────┐
          ▼                            ▼
   Redis Pub/Sub（事件总线）        asynq（任务队列）
   Subscriber 消费事件             worker 消费任务
   → 触发 asynq 任务               → 执行邮件/库存/统计/导出
   → 同步副作用（轻量）            → 重试/延时/定时
                        │
                        └→ Webhook 出口（外部系统订阅）
```

- **事件**：领域发生事实的**通知**（已发生），消费者只读，不可回滚事务。
- **任务**：具体**副作用执行**（发邮件、扣库存、生成报表），可重试、可延时、可定时。
- **职责分离**：Subscriber 收到事件 → 投递 asynq 任务 → 任务执行真正副作用；避免事件处理阻塞或丢失。

---

## 2. 事件总线（Redis Pub/Sub）

### 2.1 事件结构
```json
{
  "id": "evt_01HXYZ...",          // ULID/UUID
  "name": "order.placed",         // 事件名
  "occurred_at": "2026-06-18T08:30:00Z",
  "aggregate": "order",           // 聚合类型
  "aggregate_id": "123",
  "actor": { "type": "admin", "id": "5" },
  "data": { ... },                // 事件负载
  "version": 1
}
```

### 2.2 Bus 接口
```
events.Bus.Publish(ctx, event) error
events.Bus.Subscribe(topic, handler)            // topic = 事件名或通配 "order.*"
```
- channel 命名：`efa:events:{event_name}`。
- 通配订阅：`PSUBSCRIBE efa:events:order.*`。
- `Publish` 在**事务提交后**调用（service post-commit hook）；事务回滚则不发。

### 2.3 一致性策略（ADR-0005）
- **MVP：post-commit publish**（事务提交后同步发 Redis）。简单；风险：进程在提交后、发送前崩溃则事件丢失（极少）。
- **升级：outbox**。事务内写 `event_outbox` 表（与业务同事务）→ 独立 relay 把 outbox 投到 Redis → 标记已发。保证至少一次。
- MVP 先 post-commit，预留 `EventOutbox` 实体与 relay，按需切换。

### 2.4 Subscriber
- 运行于 `cmd/worker`（也可在 admin 进程内嵌，MVP 集中在 worker）。
- handler 签名：`func(ctx, event) error`；panic recover；错误记 `event_log`。
- 投递语义：**at-least-once**，消费者需幂等（用 `event.id` 去重，Redis SET 去重窗口）。

---

## 3. asynq 任务队列

### 3.1 用途
- 耗时/可重试/延时/定时的副作用：发邮件、生成缩略图、导出报表、Abandoned Cart 召回、库存同步、统计聚合。

### 3.2 任务定义
```
jobs.Task(type, payload, opts)
opts: MaxRetry, Timeout, RetryDelay, Queue, ProcessAt(延时), PeriodicConfig(定时)
```
- 队列分级：`critical`（订单相关）/ `default` / `low`（统计/清理）。
- worker 并发：`ASYNQ_CONCURRENCY`（默认 10）。

### 3.3 重试与死信
- 失败指数退避重试（默认 25 次）；超过入死信队列 `asynq:dead`，告警 + 后台可手动重投。

### 3.4 定时任务（asynq scheduler）
- Abandoned Cart 召回：每小时扫描 `cart.status=active AND updated_at<now-2h` → 投召回任务。
- 优惠到期：每日扫 `discount.ends_at<now` → 标记 expired + 事件。
- 通知重发：每分钟扫 `notification.status=failed AND attempts<3` → 重投。

---

## 4. Webhook 出口

### 4.1 机制
- `webhook` 表登记外部订阅：`(url, events[], secret, is_active)`。
- 事件发布后，`WebhookDispatcher`（Subscriber）对匹配事件投递 HTTP POST 到 url，body = event payload + `X-EFA-Signature`（HMAC-SHA256(secret, body)）。
- 重试：5xx/超时按 asynq 退避重试；4xx 不重试（配置错误）。

### 4.2 事件订阅过滤
- `webhook.events` 为事件名数组或通配；匹配后投递。
- 记 `event_log`（attempts、last_response_code、status）。

---

## 5. 关键事件清单（全域汇总）

### 产品域
| 事件 | 触发 | 典型消费者 |
|---|---|---|
| `product.created` | 创建 | 搜索索引任务、webhook |
| `product.updated` | 更新 | 搜索索引、缓存失效 |
| `product.published` | 发布 | webhook、通知 |
| `variant.inventory.changed` | 库存增减 | 低库存告警任务、webhook |
| `variant.price.updated` | 改价 | 缓存失效、价格快照 |
| `category.updated` | 分类变更 | 缓存失效 |

### 客户域
| 事件 | 消费者 |
|---|---|
| `customer.registered` | 欢迎邮件任务、webhook |
| `customer.updated` | 缓存失效 |
| `customer.group.changed` | 差异化定价刷新 |

### 交易域
| 事件 | 消费者 |
|---|---|
| `cart.abandoned` | 召回邮件延时任务 |
| `checkout.completed` | （内部）转 order.placed |
| `order.placed` | 确认邮件、扣库存任务、webhook、统计 |
| `order.paid` | 发货通知、财务统计、webhook |
| `order.fulfilled` | 发货邮件、webhook |
| `order.cancelled` | 取消邮件、回滚库存任务、webhook |
| `order.returned` | 退款邮件、财务、webhook |
| `fulfillment.created` | 物流 webhook |

### 支付/物流/优惠
| 事件 | 消费者 |
|---|---|
| `payment.authorized/captured/refunded` | 财务统计、webhook |
| `shipping_option.quoted` | （观测） |
| `discount.redeemed` | 用量统计、webhook |
| `discount.expired` | 缓存失效 |

### CMS
| 事件 | 消费者 |
|---|---|
| `page.published/updated` | 缓存失效、CDN purge 任务 |
| `blog.published` | RSS 生成、通知 |
| `menu.updated` | 缓存失效 |

### 询盘/通知
| 事件 | 消费者 |
|---|---|
| `inquiry.received` | 通知邮件（客户+站点）、webhook、CRM 推送任务 |
| `inquiry.updated/assigned` | 跟进通知 |
| `inquiry.converted` | 统计、webhook |
| `notification.sent/failed` | 重发任务、告警 |
| `webhook.delivered/failed` | 告警、死信处理 |

### 基础设施
| 事件 | 消费者 |
|---|---|
| `store.updated` | 缓存失效 |
| `media.uploaded` | 缩略图生成任务 |
| `media.deleted` | 存储清理 |
| `admin.logged_in` | 审计日志 |

---

## 6. 幂等与去重

- 事件消费者：用 `event.id` 在 Redis `SETNX efa:dedup:{event.id}`（TTL 24h）去重；已存在则跳过。
- 任务消费者：关键任务 payload 带 `idempotency_key`（=event.id 或业务键），执行前查 `task_run` 表去重（可选）。
- 邮件类：`notification` 表持 `(template_code, entity_id, channel)` 去重窗口，防重复发送。

---

## 7. 失败与可观测

- 事件处理失败 → `event_log(status=failed, attempts, error)`；asynq 自带重试 + 死信。
- 指标：`event_publish_total`、`event_handle_duration`、`event_handle_failed_total`、`asynq_queue_depth`、`task_processed_total{type,status}`。
- 告警：死信队列长度、通知连续失败、webhook 连续失败。

---

## 8. 包结构

```
internal/events/      # Bus + 路由 + outbox relay（预留）
internal/jobs/        # asynq client/server + 各域 task 定义 + handler
domain/<x>/event.go   # 该域事件类型定义 + 发布辅助
domain/<x>/handler.go # 该域 Subscriber handler（注册到 worker）
```
- handler 注册：worker 启动时遍历各域 `RegisterHandlers(bus, jobsClient)`。
- 跨域：handler 只消费事件 + 调本域 service，不跨域直接操作。
