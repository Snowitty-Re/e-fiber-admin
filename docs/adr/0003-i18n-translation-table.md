# ADR-0003: 多语言内容采用主表 + 翻译表

- 状态：Accepted
- 日期：2026-06-18
- 关联：`docs/02` §2.2、`docs/03` §1.4/§5.2 等

## 背景

外贸站点多语言是标配，需对产品名/描述、CMS 内容、菜单、表单标签等做翻译。存储方式影响查询、索引、校验与扩展。

## 决策

可翻译字段从主表拆出到 `{entity}_translation(entity_id, locale, <fields>)`，`(entity_id, locale)` 唯一。主表仅留不可翻译字段（slug、状态、重量等）。请求按 `Accept-Language` 解析，缺失回退 `store.default_locale`。

## 备选

| 方案 | 否决理由 |
|---|---|
| JSONB 翻译列（主表 text 字段存 `{en:...,zh:...}`） | 索引/查询/必填校验复杂，字段膨胀，不利于约束 |
| 混合（JSONB + 翻译表按字段取舍） | 复杂度增加，MVP 统一一种更可维护 |

## 后果

- 正面：查询清晰、可对翻译字段建索引、扩展性好（MedusaJS 风格）。
- 负面：表数量增多、JOIN 增多（用 ent edges 预加载缓解）。
- 约束：所有可翻译实体统一该模式，ent schema 用 edges 表达 `1—N`。
