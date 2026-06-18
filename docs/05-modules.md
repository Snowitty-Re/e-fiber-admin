# 05 - 模块详细设计

> 状态：规范基线（模块边界与职责唯一真源）
> 实现约束：禁止跨域直接操作对方 repository（`AGENTS.md` §4.3 / `docs/01` §2）；跨域协作经 `service` 编排或事件。
> 实体字段见 `docs/03-data-model.md`，端点约定见 `docs/04-api-design.md`，不在此重复全字段。

每个模块统一描述：**职责 / 核心实体 / 关键接口 / 主要流程 / 扩展点 / 事件**。

---

## 模块 0 — 基础设施（settings / media / auth / events / jobs）

> 一切模块的依赖底座，最先实现（见 `docs/08-roadmap.md` M0）。

### 职责
- `settings`：站点配置（store 单条 + settings 键值 + feature_flags）读写与校验。
- `media`：对象存储上传/读取/删除，媒体库管理，缩略图/格式（依赖 S3 client）。
- `auth`：admin/customer 两套 JWT 签发校验、刷新、吊销（详见 `docs/06-auth-rbac.md`）。
- `events`：Redis Pub/Sub 事件总线，发布/订阅/路由（详见 `docs/07-events-jobs.md`）。
- `jobs`：asynq 任务注册、调度、重试。

### 关键接口（示意）
```
settings.Service.Get(namespace, key) / Set(namespace, key, value) / GetStore()
media.Service.Upload(file) / Delete(key) / SignedURL(key)
events.Bus.Publish(event) / Subscribe(topic, handler)
jobs.Client.Enqueue(task, opts) / Register(taskType, handler)
```

### 扩展点
- `media` 支持多 provider（local/s3/minio）通过接口切换。
- `events` 可升级 outbox（ADR-0005）。
- `jobs` 注册中心开放给各域注册 handler。

### 事件
- `store.updated`、`media.uploaded`、`media.deleted`。

---

## 模块 1 — 产品域（product）

### 职责
商品目录全生命周期：产品/规格/价格/分类/集合/标签/媒体关联/多语言。

### 核心实体
`product`、`variant`、`variant_price`、`product_option`、`product_option_value`、`variant_option_value`、`category`(+translation)、`collection`(+translation)、`tag`(+translation)、`product_media`。

### 关键接口
```
ProductService:
  Create(in) -> Product
  Update(id, in, version) -> Product
  Get(id, expand) -> Product
  List(filter, page) -> []Product
  Delete(id) / Archive(id)
  Publish(id) / Unpublish(id)
VariantService:
  AddVariant(productID, in)
  UpdatePrice(variantID, currency, amount)   # 按币种独立存价
  UpdateInventory(variantID, qty, op)
CategoryService / CollectionService / TagService: CRUD + 树/排序
```

### 主要流程
1. **创建产品**：先建 product（draft）→ 建 variant(s) → 写 variant_price(每启用币种一条) → 关联 option/媒体 → 发布。
2. **多币种定价**：variant_price 按 `currency_code` 维护，列表/详情按请求 locale+currency 展示对应价格；缺失币种回退 default_currency 并标记。
3. **库存**：variant.inventory 增减通过事件或事务内扣减；`allow_backorder` 控制缺货可下单。

### 扩展点
- 产品属性扩展：`product.meta jsonb` + 后续可加 `product_attribute` 动态属性。
- 价格策略：未来可加 `price_list` 覆盖层（本期不做，见 ADR-0002）。

### 事件
- `product.created` / `product.updated` / `product.published` / `product.archived`
- `variant.inventory.changed` / `variant.price.updated`
- `category.updated`

---

## 模块 2 — 区域域（region）

### 职责
外贸地基：区域/语言/币种/税率管理；驱动 storefront 的语言与价格选择。

### 核心实体
`region`、`locale`、`currency`、`tax_rate`、`region_payment_provider`、`region_shipping_option`。

### 关键接口
```
RegionService: CRUD + Activate(id) + AssignPaymentProviders / AssignShippingOptions
CurrencyService / LocaleService: Enable/Disable(code)
TaxRateService: CRUD（按 region/country）
```

### 主要流程
- 站点初始化：建 default region（locale+currency）→ 启用币种/语言字典 → 配税率。
- storefront 解析：`Accept-Language` + `X-Currency` → 匹配 region → 价格/税/物流/支付可用集。

### 扩展点
- 未来多仓/多渠道库存隔离预留（region 不承担库存）。

### 事件
- `region.created` / `region.updated` / `currency.enabled` / `tax_rate.updated`

---

## 模块 3 — 客户域（customer）

### 职责
前台客户账号、地址簿、客户分组；与询盘/订单关联。

### 核心实体
`customer`、`customer_address`、`customer_group`、`customer_group_member`。

### 关键接口
```
CustomerService: Register / Login / Get / Update / Disable
AddressService: CRUD / SetDefaultShipping / SetDefaultBilling
CustomerGroupService: CRUD / AddMembers / RemoveMembers
```

### 主要流程
- 询盘客户可无密码（guest）；下单/注册时升级为正式 customer（邮箱合并去重）。
- 地址簿用于 checkout 预填，下单时快照进 order（地址可删，订单不变）。

### 扩展点
- 客户分组用于差异化定价/优惠定向（discount_condition 引用 group）。

### 事件
- `customer.registered` / `customer.updated` / `customer.address.added` / `customer.group.changed`

---

## 模块 4 — 交易域（cart / checkout / order / fulfillment / return）

### 职责
商城主链路：购物车 → 结账 → 订单 → 履约 → 退换。

### 核心实体
`cart`、`cart_item`、`checkout`、`order`、`order_item`、`fulfillment`、`fulfillment_item`、`return`、`return_item`、`swap`。

### 关键接口
```
CartService: Create / AddItem / UpdateItem / RemoveItem / ApplyDiscount / Get
CheckoutService: Init(cartID) / SetShipping / SetBilling / SetShippingOption /
                 SetPaymentSession / Complete() -> Order
OrderService: Get / List / Cancel / Fulfill / AddFulfillment / CreateReturn / CreateSwap
```

### 主要流程
1. **下单**：cart 装载 → checkout 分步采集（地址→物流→支付）→ Complete 校验库存/优惠/税 → 创建 order（快照地址/价格/优惠）→ 发布 `order.placed` → cart 标记 converted。
2. **履约**：order 创建 fulfillment + fulfillment_item（按 item 分批）→ tracking_number → `order.fulfilled`（部分/全）。
3. **退换**：CreateReturn 校验可退数量 → 退款 transaction → `order.returned`；swap 生成新 order_item。
4. **Abandoned Cart**：cart 长时未转化 → `cart.abandoned` 事件 → asynq 延时召回任务。

### 扩展点
- 履约可多仓/多物流商（预留 `fulfillment.provider` 字段）。
- 退换流程可扩展审批节点（MVP 直接退，后续加审批状态机）。

### 事件
- `cart.created` / `cart.item.added` / `cart.abandoned`
- `checkout.completed`
- `order.placed` / `order.paid` / `order.fulfilled` / `order.cancelled` / `order.returned`
- `fulfillment.created`

### 状态机（order.status）
```
pending --paid--> paid --fulfilled--> fulfilled
pending --cancelled--> cancelled
paid/fulfilled --return--> refunded（partial/全）
```

---

## 模块 5 — 支付 / 物流 / 优惠（payment / shipping / discount）

### 职责
交易支撑：支付网关抽象、运费配置、优惠与券。

### 核心实体
`payment_provider`、`payment_session`、`transaction`；`shipping_profile`、`shipping_option`、`product_shipping_profile`；`discount`、`discount_rule`、`discount_condition`。

### 关键接口
```
PaymentProviderService: CRUD / Activate(code) / Authorize / Capture / Refund (via Provider interface)
ShippingService: Profile CRUD / Option CRUD / Quote(cart, address) -> []Option
DiscountService: CRUD / Validate(code, cart) / Apply / UsageInc
```

### Provider 接口（关键扩展点）
```go
type PaymentProvider interface {
    Code() string
    Authorize(ctx, session) (ProviderData, error)
    Capture(ctx, session, amount) (Tx, error)
    Refund(ctx, session, amount) (Tx, error)
    WebhookHandler(ctx, payload) error
}
```
- MVP 内置占位 + Stripe 占位实现；其余按需实现，注册到 `payment_provider` 表驱动。

### 主要流程
- **支付**：checkout 选 provider → 创建 payment_session → provider.Authorize → Complete 时 Capture → transaction 记录 → `order.paid`。
- **物流报价**：根据 cart items 的 shipping_profile + 收货国 → 匹配 region_shipping_option → 返回可用 `ShippingOption[]` + 估价。
- **优惠**：checkout 输入 code → 校验有效期/用量/条件(products/collections/groups) → 计算 `DiscountRule`（percentage/fixed/shipping/free_item）→ 应用到 totals。

### 扩展点
- 优惠规则可扩展自定义 type（注册到 DiscountRule registry）。
- 物流可接实时运费 API（预留 `shipping_option.provider` + 外部报价接口）。

### 事件
- `payment.authorized` / `payment.captured` / `payment.refunded`
- `shipping_option.quoted`
- `discount.redeemed` / `discount.expired`

---

## 模块 6 — CMS 域（cms）

### 职责
企业站主体 + 商城落地页：页面/博客/内容块/导航菜单/SEO。

### 核心实体
`page`(+translation)、`blog_post`(+translation)、`blog_category`、`blog_tag`、`block`(+translation)、`menu`、`menu_item`(+translation)。

### 关键接口
```
PageService: CRUD / Publish / GetBySlug(slug, locale)
BlogService: Post CRUD / Publish / ListByCategory
BlockService: CRUD / GetByCode(code, locale)
MenuService: CRUD / AddItem / Reorder / GetByLocation(location, locale)
```

### 主要流程
- 页面内容用 `content jsonb` 存结构化 block 引用或内联块；storefront 按 locale 渲染。
- 菜单树形：menu_item 自引用 parent_id；按 location + locale 解析整树。

### 扩展点
- block 类型可扩展（注册 BlockType handler）。
- 模板系统预留 `page.template` 字段（MVP 不做主题）。

### 事件
- `page.published` / `page.updated` / `blog.published` / `menu.updated`

---

## 模块 7 — 询盘 / 通知域（inquiry / notification）

### 职责
外贸 B2B lead 通道：表单定义、询盘提交、跟进、转单；邮件模板与通知。

### 核心实体
`form_definition`(+translation)、`inquiry`、`email_template`(+translation)、`notification`、`webhook`、`event_log`。

### 关键接口
```
FormService: CRUD / Publish
InquiryService: Submit(formID, payload) -> Inquiry / Assign / UpdateStatus / Convert(orderID)
NotificationService: Send(channel, templateCode, recipient, vars)
EmailTemplateService: CRUD
WebhookService: CRUD / Dispatch(event)
```

### 主要流程
1. **询盘提交**（storefront）：`POST /store/inquiries` → 校验 form_definition.fields → 存 inquiry（status=new）→ 发布 `inquiry.received`。
2. **通知**：Subscriber 监听 `inquiry.received` → 用 `inquiry_received` 模板渲染（按 locale）→ 发邮件给客户 + 站点 notify_emails → 记 notification。
3. **跟进**：admin 改 inquiry.status（contacted/qualified/converted）→ `inquiry.updated`。
4. **转单**：`/admin/inquiries/{id}/convert` → 基于询盘商品生成 cart/checkout/order → 回写 `converted_order_id` → `inquiry.converted`。

### 扩展点
- 表单字段类型可扩展（注册 FieldType validator）。
- 通知渠道可扩展（email/webhook/in-app/push，registry 注册）。
- Webhook 让外部 CRM/ERP 订阅询盘与订单事件。

### 事件
- `inquiry.received` / `inquiry.updated` / `inquiry.converted` / `inquiry.assigned`
- `notification.sent` / `notification.failed`
- `webhook.delivered` / `webhook.failed`

---

## 模块依赖图（实现顺序参考）

```
M0 基础设施(settings/media/auth/events/jobs)
   ├── M1 产品域 ──┐
   ├── M2 区域域 ──┤
   ├── M3 客户域 ──┤
   │              ├── M4 交易域 ──┐
   │              │              ├── M5 支付/物流/优惠
   ├── M6 CMS 域 ──┘              │
   └── M7 询盘/通知 ───────────────┘
```
- M1/M2/M6/M7 可并行（均只依赖 M0）。
- M4 依赖 M1/M2/M3；M5 依赖 M4；M7 转单依赖 M4。

---

## 模块边界硬规则（再强调）

- 每个 `domain/<x>` 包内：`entity.go`、`service.go`、`repo.go`(接口)、`event.go`。
- `repository` 包提供各域 repo 的 ent 实现；`service` 通过接口依赖，**不跨域直接 new 对方 repo**。
- 跨域数据需求：①经 service 公共方法；②经事件异步通知；③只读 DTO 由 service 聚合。
- 禁止 `domain/product` import `domain/order`；如需「下单扣库存」，由 `order_service` 调 `product_service.UpdateInventory`（应用层编排），不直接碰 product repo。
