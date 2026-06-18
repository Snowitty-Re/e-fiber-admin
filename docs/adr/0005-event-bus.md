# ADR-0005: 事件总线采用 Redis Pub/Sub，MVP 用 post-commit，预留 outbox

- 状态：Accepted
- 日期：2026-06-18
- 关联：`docs/02` §2.6、`docs/07-events-jobs.md` §2.3

## 背景

系统需解耦副作用（邮件/库存/统计/缓存）与外部集成（Webhook）。事件需在事务提交后发布，避免「事件已发但事务回滚」的不一致。需在简单性与一致性间取舍。

## 决策

- 事件总线基于 **Redis Pub/Sub**，技术栈统一复用 Redis（`AGENTS.md` §4.1）。
- **MVP 用 post-commit publish**：事务提交后同步发 Redis。简单、无额外表。
- **预留 outbox 升级**：事务内写 `event_outbox` 表（与业务同事务）+ 独立 relay 投递，保证至少一次。MVP 不实现，按一致性需求切换。

事件投递语义 at-least-once，消费者用 `event.id` 幂等去重。

## 备选

| 方案 | 否决理由 |
|---|---|
| 同步 hooks（Woo 风格） | 阻塞主流程，失败影响事务，扩展性差 |
| 直接上 outbox | MVP 复杂度过高，简单性不足 |
| Kafka / MQ | 引入新中间件，违反「禁本机装中间件，统一 Docker」精简原则，MVP 过重 |

## 后果

- 正面：简单起步，Redis 复用，扩展性好（升 outbox 仅加表 + relay）。
- 负面：post-commit 在「提交后发送前崩溃」会丢事件（极小概率）；消费者需幂等。
- 升级路径：当关键业务（支付/库存）要求强一致时，启用 outbox，无需改事件 API。
