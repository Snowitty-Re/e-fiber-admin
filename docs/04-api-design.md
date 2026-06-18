# 04 - API 设计规范

> 状态：规范基线（API 唯一真源）
> 实现约束：REST 资源命名/错误码/分页/版本以本文件为准（`AGENTS.md` §4.3）。
> OpenAPI 3.1 规范文件落地于 `api/openapi/`。

---

## 1. 总则

- 风格：**REST + JSON**，资源命名复数、小写、连字符（`/admin/products`、`/admin/order-items`）。
- 协议：HTTPS（开发 HTTP）；所有响应 `application/json; charset=utf-8`。
- 字符集：UTF-8；时间：ISO 8601 UTC（`2026-06-18T08:30:00Z`）。
- 金额：**整数 + 货币代码**，不传浮点（`{"amount": 1234, "currency_code": "USD"}`）。
- 语言协商：请求头 `Accept-Language: en` / `zh` / `en-US`；响应返回对应 locale 翻译，缺失回退 `store.default_locale`。

---

## 2. 版本与路径

- 版本前缀：`/api/v1`（URI 版本化，简单显式）。
- 两套入口（鉴权体系隔离，见 `docs/06-auth-rbac.md`）：
  - **Admin API**：`/api/v1/admin/...` —— 后台运营，JWT(admin) + RBAC。
  - **Storefront API**：`/api/v1/store/...` —— 前台消费者/询盘，JWT(customer) 可选 + 公开读。
- 健康检查不走版本：`/healthz`、`/readyz`。

示例：
```
GET    /api/v1/admin/products
POST   /api/v1/admin/products
GET    /api/v1/admin/products/{id}
PATCH  /api/v1/admin/products/{id}
DELETE /api/v1/admin/products/{id}

GET    /api/v1/store/products
POST   /api/v1/store/inquiries
POST   /api/v1/store/cart/{cart_id}/items
```

---

## 3. 资源命名与操作约定

### 3.1 CRUD 映射
| 操作 | 方法 | 路径 | 成功码 |
|---|---|---|---|
| 列表 | GET | `/resources` | 200 |
| 创建 | POST | `/resources` | 201 |
| 详情 | GET | `/resources/{id}` | 200 |
| 全量更新 | PUT | `/resources/{id}` | 200 |
| 部分更新 | PATCH | `/resources/{id}` | 200 |
| 删除（软删） | DELETE | `/resources/{id}` | 204 |

### 3.2 子资源
- 嵌套最多 2 层：`/admin/products/{id}/variants`、`/admin/orders/{id}/fulfillments`。
- 跨资源关联操作用顶层 + 查询参数，避免深嵌套：`/admin/products?collection_id=12`。

### 3.3 动作（非 CRUD）
- 用动词子资源，POST：`/admin/orders/{id}/cancel`、`/admin/orders/{id}/fulfill`、`/admin/inquiries/{id}/convert`。

### 3.4 批量
- 批量操作走 `POST /resources/batch`，body 含 `{ "ops": [ { "op": "delete", "id": 1 }, ... ] }`，响应逐项结果（MVP 按需实现）。

---

## 4. 分页 / 排序 / 过滤 / 字段投影

### 4.1 分页（游标 + 偏移双支持）
- 偏移分页（默认）：`?page=1&page_size=20`（page_size ≤100）。
- 游标分页（大数据集）：`?cursor=<base64>&limit=20`。
- 响应固定结构：
```json
{
  "data": [ ... ],
  "pagination": {
    "page": 1, "page_size": 20, "total": 137, "has_more": true,
    "next_cursor": "eyJpZCI6MTIzfQ=="
  }
}
```

### 4.2 排序
- `?sort=-created_at,title` —— `-` 降序，多字段逗号分隔。

### 4.3 过滤
- 等值：`?status=published`；多值：`?status=published,draft`（IN）。
- 区间：`?created_at[gte]=2026-01-01&created_at[lt]=2026-07-01`。
- 模糊：`?q=keyword` 走全文检索。

### 4.4 字段投影
- `?fields=id,slug,variants(id,sku,price)` —— 减少载荷，按需实现（MVP 可仅支持顶层 fields）。

### 4.5 展开
- `?expand=variants,variants.prices,category` —— 控制关联加载，避免 N+1 与过载。

---

## 5. 请求与响应结构

### 5.1 请求体
- 一律 JSON；`Content-Type: application/json`。
- 资源包一层，便于扩展与校验：
```json
{ "product": { "slug": "t-shirt", "product_type": "simple", "variants": [...] } }
```

### 5.2 单资源响应
```json
{ "product": { "id": 1, "slug": "t-shirt", "..." : "..." } }
```

### 5.3 列表响应
```json
{ "data": [ {"product": {...}}, {"product": {...}} ], "pagination": {...} }
```

### 5.4 空值
- 可空字段显式 `null`，不省略（客户端类型友好）；列表无数据返回 `data: []`。

---

## 6. 鉴权与会话

### 6.1 Admin
- 登录：`POST /api/v1/admin/auth/login` → `{ access_token, refresh_token, expires_in }`。
- 请求带：`Authorization: Bearer <access_token>`。
- 刷新：`POST /api/v1/admin/auth/refresh`（body: refresh_token）。
- 登出：`POST /api/v1/admin/auth/logout`（吊销 refresh）。
- RBAC：token 内含 permissions，路由层校验 (resource, action)。

### 6.2 Customer（Storefront）
- `POST /api/v1/store/auth/register`、`/login`、`/refresh`、`/logout`。
- 公开读（产品/页面）无需 token；下单/询盘需 token 或 guest 标识。

### 6.3 通用
- access token 短 TTL（15m），refresh 长 TTL（30d）存 Redis 可吊销。
- 401（token 失效/缺失）、403（权限不足）区分。

---

## 7. 统一错误响应

### 7.1 结构
```json
{
  "error": {
    "code": "PRODUCT_NOT_FOUND",
    "message": "product not found",
    "status": 404,
    "request_id": "req_abc123",
    "details": [ { "field": "id", "issue": "must exist" } ]
  }
}
```

### 7.2 HTTP 状态码约定
| 码 | 语义 |
|---|---|
| 200 | 成功（GET/PATCH/PUT） |
| 201 | 创建成功 |
| 204 | 删除成功（无响应体） |
| 400 | 请求参数/校验错误 |
| 401 | 未认证 / token 失效 |
| 403 | 已认证但无权限 |
| 404 | 资源不存在 |
| 409 | 冲突（唯一约束/状态非法） |
| 422 | 业务校验失败（语义层） |
| 429 | 限流 |
| 500 | 服务器内部错误 |
| 503 | 依赖不可用（DB/Redis） |

### 7.3 错误码命名
- `<DOMAIN>_<REASON>` 大写下划线：`PRODUCT_NOT_FOUND`、`ORDER_INVALID_STATE`、`VARIANT_OUT_OF_STOCK`、`AUTH_TOKEN_EXPIRED`。
- 完整错误码表落地 `api/openapi/errors.yaml`，service 抛 `pkg/errors.AppError{Code, Status, Message}`。

### 7.4 校验错误
- 400 + `details[]` 列出每个字段错误，便于前端定位。

---

## 8. 幂等与并发

### 8.1 幂等键
- 写操作支持 `Idempotency-Key` 请求头（创建订单/支付/询盘等关键写）；服务端缓存响应 24h，重复键返回首次结果。

### 8.2 乐观并发
- 资源响应含 `version`；PATCH 带 `If-Match: <version>`，不匹配返回 409 `VERSION_CONFLICT`。

---

## 9. 限流

- 全局 + 按端 + 按 IP/用户：
  - Admin：宽松（已鉴权）。
  - Storefront 登录/询盘：严格（防刷），如 `10/min/IP`。
- 超限返回 429 + `Retry-After` 头。
- 实现：Redis 滑动窗口（`internal/http/fiber/middleware/ratelimit`）。

---

## 10. 通用响应头

| 头 | 说明 |
|---|---|
| `X-Request-ID` | 每请求生成，响应回写，便于链路追踪 |
| `X-RateLimit-Remaining` | 剩余配额 |
| `Retry-After` | 429 时返回 |
| `Idempotency-Key`（请求） | 关键写幂等 |

---

## 11. OpenAPI 规范

- 单一文件起步 `api/openapi/openapi.yaml`，按域拆 `components/schemas/<domain>.yaml`。
- 由 handler 注解或独立维护（MVP 独立维护 yaml，保证与 docs 一致）。
- 暴露 `/api/v1/openapi.json`（可选）+ Swagger UI（按需，非 MVP）。
- 错误码集中 `api/openapi/errors.yaml`。

---

## 12. 资源端点总表（按域，概要；细节见 `docs/05-modules.md`）

| 域 | Admin 端点（节选） | Store 端点（节选） |
|---|---|---|
| Auth | `/admin/auth/{login,refresh,logout,me}` | `/store/auth/{register,login,refresh,logout,me}` |
| Product | `/admin/products`, `/admin/products/{id}/variants`, `/admin/categories`, `/admin/collections` | `/store/products`, `/store/products/{slug}` |
| Region | `/admin/regions`, `/admin/currencies`, `/admin/locales`, `/admin/tax-rates` | （由 store 上下文隐式选择） |
| Customer | `/admin/customers`, `/admin/customers/{id}/addresses` | `/store/customers/me`, `/store/customers/me/addresses` |
| Cart/Checkout | （管理视角只读）`/admin/carts`, `/admin/orders` | `/store/cart`, `/store/cart/items`, `/store/checkout` |
| Order | `/admin/orders`, `/admin/orders/{id}/{cancel,fulfill,return}` | `/store/orders`, `/store/orders/{id}` |
| Payment/Ship/Discount | `/admin/payment-providers`, `/admin/shipping-options`, `/admin/discounts` | `/store/shipping-options`, `/store/payment-sessions` |
| CMS | `/admin/pages`, `/admin/blog-posts`, `/admin/menus`, `/admin/blocks` | `/store/pages`, `/store/blog-posts`, `/store/menus/{slug}` |
| Inquiry | `/admin/forms`, `/admin/inquiries`, `/admin/inquiries/{id}/convert` | `/store/inquiries`（提交） |
| Media/Settings/Webhook | `/admin/media`, `/admin/settings`, `/admin/webhooks` | `/store/media/{id}`（公开读） |

> 完整端点清单与请求/响应 schema 在 `docs/05-modules.md` 逐模块给出，并落 OpenAPI。
