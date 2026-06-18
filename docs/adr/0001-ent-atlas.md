# ADR-0001: 数据层采用 ent + Atlas

- 状态：Accepted
- 日期：2026-06-18
- 决策者：用户 + AI（经会话确认）
- 关联：`AGENTS.md` §4.1、`docs/01-architecture.md` §2

## 背景

后台系统数据模型多关联（product↔variant↔price、order↔item↔fulfillment、i18n 翻译表、多币种价格表），需类型安全 ORM 与可审查的迁移机制。Go 生态可选 ent、GORM、sqlc 等。

## 决策

采用 **ent + Atlas**：
- ent 作为 schema 即代码的 ORM，生成类型安全访问代码，edges 表达多关联优雅。
- Atlas 作为迁移工具，从 ent schema 生成并应用声明式迁移。

## 备选

| 方案 | 优点 | 否决理由 |
|---|---|---|
| GORM | 流行、上手快 | 复杂查询易 N+1，关联预加载弱，schema 与代码分离 |
| sqlc | SQL 控制力强、性能可控 | 多关联场景写法繁重，ORM 能力弱，开发效率低 |

## 后果

- 正面：类型安全、关系表达清晰、迁移可审查、与 MedusaJS 风格的多关联模型契合。
- 负面：ent 学习曲线略高；生成代码需 `make ent-gen`。
- 约束：禁止擅自引入 GORM/sqlc 替代（`AGENTS.md` §4.1）。
