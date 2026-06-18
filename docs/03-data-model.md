# 03 - 数据模型设计

> 状态：规范基线（数据模型唯一真源）
> 实现约束：ent schema 字段/命名/关系以本文件为准；禁止擅自偏离（`AGENTS.md` §4.3）。
> 决策来源：`docs/02-tradeoff-woo-medusa.md`，关键决策见 `docs/adr/`。

---

## 1. 全局约定

### 1.1 命名
- 表名：`snake_case` 单数（`product`、`variant`、`order_item`）。
- 字段名：`snake_case`；外键 = `<表单数>_id`（`product_id`、`variant_id`）。
- ent schema：Go 驼峰命名，`ent` 注解映射为表/列名。
- 货币代码：ISO 4217 大写（`USD`、`CNY`、`EUR`）；语言代码：BCP 47 小写（`en`、`zh`、`en-US`）。

### 1.2 公共字段（所有实体默认含）
| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | bigint / uuid | 主键；MVP 用 bigint 自增（ent default） |
| `created_at` | timestamptz | 创建时间 |
| `updated_at` | timestamptz | 更新时间 |
| `deleted_at` | timestamptz nullable | 软删标记；查询默认过滤 `deleted_at IS NULL` |
| `created_by` | bigint nullable | 创建人 admin_id（审计） |
| `updated_by` | bigint nullable | 更新人 admin_id（审计） |
| `version` | int default 1 | 乐观锁，更新时 `WHERE version = ? ... SET version = v+1` |

> 预留扩展位：所有表预留 `tenant_id bigint nullable`（MVP 不实现隔离逻辑，留作未来多租户升级，见 ADR-0004）。

### 1.3 Money 值对象
- 金额以**最小货币单位整数**存储（如 USD 存分：$12.34 → 1234；JPY 无小数则原值）。
- 数据库列：`amount bigint` + `currency_code char(3)`，不存浮点。
- 价格表统一 `variant_price(variant_id, currency_code, amount)`。

### 1.4 i18n 翻译策略（主表 + 翻译表）
- 可翻译字段（标题/描述/正文/SEO 文案等）从主表拆出到 `{entity}_translation(entity_id, locale, <fields>)`。
- 主表仅留不可翻译字段 + 默认语言回退字段（`default_locale`、`slug` 等不翻译字段留主表）。
- `(entity_id, locale)` 唯一索引。
- 详见 ADR-0003。

### 1.5 软删与唯一约束
- 唯一约束含软删时，用 `slug` + `deleted_at IS NULL` 部分唯一索引；或唯一键加 `deleted_at` 列。
- 业务上「已删 slug 可复用」。

---

## 2. ER 总览（领域分组）

```
┌─ 基础设施 ─┐  ┌─ 区域域 ────┐  ┌─ 产品域 ────────────┐  ┌─ 客户域 ──────┐
│ store       │  │ region       │  │ product ─ variant  │  │ customer      │
│ settings    │  │ locale       │  │   ├ option/value   │  │  └ address     │
│ media       │  │ currency     │  │   ├ price(currency)│  │  └ group       │
│ webhook     │  │ tax_rate     │  │   ├ translation    │  └────────────────┘
│ admin_user  │  │   (region*)  │  │   category/coll/tag│
│ role/perm   │  └──────────────┘  └────────────────────┘
└─────────────┘
┌─ 交易域 ────────────────────┐  ┌─ 支付/物流/优惠 ─┐  ┌─ CMS 域 ────┐  ┌─ 询盘/通知 ─┐
│ cart ─ cart_item            │  │ payment_session  │  │ page         │  │ inquiry      │
│ checkout                    │  │ payment_provider │  │ blog_post    │  │ form_def     │
│ order ─ order_item          │  │ transaction      │  │ block        │  │ notification │
│   ├ fulfillment ─ item      │  │ shipping_profile │  │ menu/item    │  │ email_tmpl   │
│   └ return/swap ─ item      │  │ shipping_option  │  └──────────────┘  └──────────────┘
└─────────────────────────────┘  │ discount/coupon  │
                                 │ discount_rule    │
                                 └──────────────────┘
```

---

## 3. 基础设施域

### 3.1 store（站点，单条）
| 字段 | 类型 | 说明 |
|---|---|---|
| id | bigint | 1（单租户单条） |
| name | varchar | 站点名 |
| slug | varchar unique | 站点标识 |
| site_type | enum | `corporate` / `store` / `hybrid` |
| default_locale | char(5) | 默认语言 |
| default_currency | char(3) | 默认币种 |
| feature_flags | jsonb | `{enable_cart,enable_checkout,enable_inquiry,enable_blog,enable_review,...}` |
| timezone | varchar | 默认时区 |
| status | enum | `active` / `maintenance` |

### 3.2 settings（键值配置，按命名空间）
| 字段 | 类型 | 说明 |
|---|---|---|
| namespace | varchar | `payment`/`shipping`/`seo`/`mail`/... |
| key | varchar | 配置键 |
| value | jsonb | 配置值（强类型由 service 校验） |
| unique | | `(namespace, key)` |

### 3.3 media（媒体库）
| 字段 | 类型 | 说明 |
|---|---|---|
| key | varchar | S3 object key |
| url | varchar | 可访问 URL |
| mime_type | varchar | |
| size_bytes | bigint | |
| width/height | int nullable | 图片尺寸 |
| alt | varchar nullable | 可翻译 alt 拆 `media_translation` |
| kind | enum | `image`/`document`/`video` |

> 媒体与产品/CMS 通过关联表多对多引用，避免重复上传。

### 3.4 admin_user / role / permission（RBAC，详见 06）
| 表 | 关键字段 |
|---|---|
| `admin_user` | email unique、password_hash(bcrypt)、status、last_login_at |
| `role` | name、slug、is_system |
| `permission` | resource、action（如 `product:write`） |
| `admin_role` | admin_id, role_id（多对多） |
| `role_permission` | role_id, permission_id（多对多） |

### 3.5 webhook / event_log
| 表 | 关键字段 |
|---|---|
| `webhook` | url、events(jsonb array)、secret、is_active |
| `event_log` | event_name、payload(jsonb)、status、attempts、last_response_code |

---

## 4. 区域域

### 4.1 region
| 字段 | 类型 | 说明 |
|---|---|---|
| name | varchar | 区域名 |
| locale | char(5) | 主语言 |
| currency_code | char(3) | 主币种 |
| tax_inclusive | bool | 价格是否含税 |
| countries | jsonb | 可达国家代码数组 |
| status | enum | active/inactive |

> 简化 Medusa：region 不强捆绑 payment/shipping，改为多对多引用（见 4.x 关联表 `region_payment_provider`、`region_shipping_option`）。

### 4.2 locale / currency
- `locale(code, name, is_active)` —— 启用语言字典。
- `currency(code, name, symbol, precision, is_active)` —— 币种字典（precision 决定 amount 小数位）。

### 4.3 tax_rate
| 字段 | 类型 | 说明 |
|---|---|---|
| region_id | bigint | 所属区域 |
| country_code | char(2) nullable | 国家级税率 |
| rate | numeric(5,4) | 税率（0.0825） |
| name | varchar | 税名（VAT/GST/Sales Tax） |
| priority | int | 多税率优先级 |

---

## 5. 产品域

### 5.1 product
| 字段 | 类型 | 说明 |
|---|---|---|
| slug | varchar unique | URL 标识 |
| product_type | enum | simple/variable/virtual/grouped |
| status | enum | draft/published/archived |
| category_id | bigint nullable | 主分类 |
| weight_g | int nullable | 重量（克），virtual 可空 |
| is_virtual | bool | 虚拟/数字商品 |
| is_downloadable | bool | 可下载 |
| seo_title / seo_desc | 翻译表 | SEO 字段拆翻译 |
| published_at | timestamptz nullable | |

### 5.2 product_translation
| 字段 | 类型 | 说明 |
|---|---|---|
| product_id | bigint | |
| locale | char(5) | |
| title | varchar | 商品名 |
| subtitle | varchar nullable | 副标题 |
| description | text | 富文本描述 |
| material / origin / packing | text nullable | 外贸属性（可翻译） |
| unique | | `(product_id, locale)` |

### 5.3 variant
| 字段 | 类型 | 说明 |
|---|---|---|
| product_id | bigint | |
| sku | varchar unique | SKU |
| barcode | varchar nullable | 条码 |
| weight_g | int nullable | 规格重量 |
| allow_backorder | bool | 允许缺货下单 |
| inventory | int | 库存（MVP 单库存，预留 warehouse） |
| position | int | 排序 |

### 5.4 variant_price（按币种独立存价，ADR-0002）
| 字段 | 类型 | 说明 |
|---|---|---|
| variant_id | bigint | |
| currency_code | char(3) | |
| amount | bigint | 最小货币单位整数 |
| compare_at_amount | bigint nullable | 划线价 |
| unique | | `(variant_id, currency_code)` |

### 5.5 option / option_value（商品规格）
| 表 | 字段 |
|---|---|
| `product_option` | product_id, name(翻译), position |
| `product_option_value` | option_id, value(翻译) |
| `variant_option_value` | variant_id, option_value_id（variant 选定的规格组合） |

### 5.6 category / collection / tag
- `category(id, slug, parent_id, position)` + `category_translation(category_id, locale, name, description)` 树形。
- `collection(id, slug, status)` + `collection_translation(...)`；`collection_product(collection_id, product_id, position)` 多对多。
- `tag(id, slug)` + `tag_translation(...)`；`product_tag(product_id, tag_id)` 多对多。

### 5.7 product_media
- `product_media(product_id, media_id, position, role)` —— role: `image`/`gallery`/`document`。
- `variant_media(variant_id, media_id)` 可选规格级图。

---

## 6. 客户域

### 6.1 customer
| 字段 | 类型 | 说明 |
|---|---|---|
| email | varchar unique | |
| phone | varchar nullable | |
| first_name / last_name | varchar | |
| password_hash | varchar nullable | 询盘客户可无密码 |
| status | enum | active/disabled/guest |
| default_currency | char(3) | |
| default_locale | char(5) | |

### 6.2 customer_address
| 字段 | 类型 | 说明 |
|---|---|---|
| customer_id | bigint | |
| first_name/last_name | varchar | |
| company | varchar nullable | B2B 公司 |
| address1/address2 | varchar | |
| city / province / postal_code / country_code | varchar | |
| phone | varchar nullable | |
| is_default_shipping / is_default_billing | bool | |

### 6.3 customer_group / customer_group_member
- `customer_group(id, slug, name)` + `customer_group_member(group_id, customer_id)`；用于差异化定价/优惠定向。

---

## 7. 交易域

### 7.1 cart / cart_item
| 表 | 字段 |
|---|---|
| `cart` | id, customer_id nullable, region_id, currency_code, email nullable, status(active/converted/abandoned), expires_at |
| `cart_item` | cart_id, variant_id, quantity, amount(currency快照), metadata jsonb |

### 7.2 checkout
| 字段 | 类型 | 说明 |
|---|---|---|
| cart_id | bigint unique | |
| email | varchar | |
| shipping_address_id / billing_address_id | bigint nullable | |
| shipping_option_id | bigint nullable | |
| payment_session_id | bigint nullable | |
| status | enum | draft/completed/expired |
| totals | jsonb | subtotal/discount/tax/shipping/total 快照 |

### 7.3 order
| 字段 | 类型 | 说明 |
|---|---|---|
| number | varchar unique | 展示单号（如 `EFA-2026-0001`） |
| customer_id | bigint nullable | |
| region_id / currency_code | | |
| email | varchar | |
| status | enum | pending/paid/fulfilled/cancelled/refunded |
| fulfillment_status | enum | not_fulfilled/partial/fulfilled |
| payment_status | enum | awaiting/paid/partial/refunded |
| shipping_address / billing_address | jsonb | 下单时快照（地址可删，订单不可变） |
| totals | jsonb | 金额快照 |
| placed_at / cancelled_at | timestamptz nullable | |

### 7.4 order_item
| 字段 | 类型 | 说明 |
|---|---|---|
| order_id | bigint | |
| variant_id | bigint nullable | 删除商品保留快照 |
| sku / title | varchar | 快照 |
| quantity | int | |
| unit_amount / total_amount | bigint | 快照（含币种） |
| metadata | jsonb | 规格快照 |

### 7.5 fulfillment / fulfillment_item
- `fulfillment(id, order_id, shipping_option_id, tracking_number, status)`；`fulfillment_item(fulfillment_id, order_item_id, quantity)`。

### 7.6 return / swap
- `return(id, order_id, status, reason, totals)`；`return_item(return_id, order_item_id, quantity, reason)`。
- `swap` 结构类似，关联新 order_item。

---

## 8. 支付 / 物流 / 优惠域

### 8.1 payment_provider / payment_session / transaction
| 表 | 字段 |
|---|---|
| `payment_provider` | id, code(stripe/paypal/...), name, is_active, config jsonb |
| `payment_session` | id, checkout_id, provider_code, status(pending/authorized/canceled), amount, currency_code, provider_data jsonb |
| `transaction` | id, payment_session_id, amount, currency_code, type(capture/refund), status, reference |

### 8.2 shipping_profile / shipping_option
| 表 | 字段 |
|---|---|
| `shipping_profile` | id, name, product_type（virtual 不计运） |
| `shipping_option` | id, profile_id, region_id, name, price_amount, price_currency, estimated_days, requirements jsonb |
| `product_shipping_profile` | product_id, profile_id |

### 8.3 discount / coupon / discount_rule
| 表 | 字段 |
|---|---|
| `discount` | id, code(优惠码 nullable), name, is_dynamic, starts_at, ends_at, usage_limit, usage_count, status |
| `discount_rule` | discount_id, type(percentage/fixed/shipping/free_item), value, allocation(all/items), target(jsonb 条件) |
| `discount_condition` | products/collections/categories/customer_groups jsonb 多条件 |

---

## 9. CMS 域

### 9.1 page
| 字段 | 类型 | 说明 |
|---|---|---|
| slug | varchar unique | |
| status | enum | draft/published |
| template | varchar | 模板名 |
| published_at | timestamptz nullable | |
- `page_translation(page_id, locale, title, content(jsonb blocks), seo_title, seo_desc)`。

### 9.2 blog_post
| 字段 | 类型 | 说明 |
|---|---|---|
| slug | varchar unique | |
| author_admin_id | bigint | |
| status | enum | draft/published |
| published_at | timestamptz nullable | |
- `blog_post_translation(blog_post_id, locale, title, excerpt, content, seo_*)`。
- `blog_category` / `blog_tag` 同 category/tag 结构。

### 9.3 block（可复用内容块）
| 字段 | 类型 | 说明 |
|---|---|---|
| code | varchar unique | 标识 |
| type | enum | richtext/image/cta/list/gallery |
| data | jsonb | 块数据 |
- `block_translation(block_id, locale, content)` 用于可翻译块。

### 9.4 menu / menu_item
| 表 | 字段 |
|---|---|
| `menu` | id, slug, location(header/footer/...) |
| `menu_item` | menu_id, parent_id, title(翻译), target_type(page/category/url/post), target_id nullable, url nullable, position |
| `menu_item_translation` | menu_item_id, locale, title |

---

## 10. 询盘 / 通知域

### 10.1 form_definition（询盘表单定义）
| 字段 | 类型 | 说明 |
|---|---|---|
| slug | varchar unique | 如 `contact`/`quote` |
| fields | jsonb | 字段定义（name/type/required/options/validation） |
| notify_emails | jsonb | 提交后通知邮箱 |
| is_active | bool | |
- `form_definition_translation(form_id, locale, title, field_labels jsonb)`。

### 10.2 inquiry（询盘提交记录）
| 字段 | 类型 | 说明 |
|---|---|---|
| form_id | bigint | |
| customer_id | bigint nullable | 已登录客户 |
| email / phone / name / company | varchar | 提交人快照 |
| payload | jsonb | 表单数据 |
| product_id / variant_id | bigint nullable | 关联商品（产品页询盘） |
| status | enum | new/contacted/qualified/converted/closed |
| converted_order_id | bigint nullable | 转单后回写 |
| assigned_admin_id | bigint nullable | 跟进人 |

### 10.3 notification / email_template
| 表 | 字段 |
|---|---|
| `email_template` | id, code(order_placed/inquiry_received/...), subject(翻译), body_html(翻译), variables jsonb |
| `email_template_translation` | template_id, locale, subject, body_html |
| `notification` | id, channel(email/webhook/inapp), recipient, template_code, payload jsonb, status, sent_at |

---

## 11. 索引与约束要点

- 所有 `(slug)` 业务唯一用**部分唯一索引**（`WHERE deleted_at IS NULL`）。
- 翻译表 `(entity_id, locale)` 唯一。
- `variant_price(variant_id, currency_code)` 唯一。
- 高频查询索引：`product(status, published_at)`、`order(customer_id, created_at)`、`inquiry(status, created_at)`。
- 全文检索：`product_translation` 加 `tsvector(title, description)` 生成列 + GIN 索引（MVP）。

---

## 12. 与 ent schema 的映射

- 本文件每个实体对应 `internal/ent/schema/<entity>.go` 一个 schema。
- 翻译表单独 schema（`ProductTranslation` 等），用 ent edges 表达 `1—N`。
- 枚举用 ent `enum` field；jsonb 用 `json` field。
- 公共字段封装为 ent mixin（`BaseMixin`：id/时间戳/软删/审计/版本/tenant_id）。
- 迁移由 Atlas 生成到 `migrations/`（`make migrate-diff name=...`）。

> 字段细节在实现时以本文件为准；如有歧义以本文件为准并记 ADR。
