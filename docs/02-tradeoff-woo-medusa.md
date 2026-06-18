# 02 - WooCommerce vs MedusaJS 取舍决策矩阵

> 状态：规范基线
> 原则：参考两者特点做**减法与重组**，目标是「易用、可快速建立、可维护」的后台，**不照搬**任何一方。

---

## 1. 取舍总纲

| 维度 | WooCommerce 特征 | MedusaJS 特征 | 我们的取舍 | 理由 |
|---|---|---|---|---|
| 架构形态 | 单体内置前后端 | headless / API-first | **API-first** | 前后端分离，外贸多渠道（站+APP+ERP+导出）复用同一 API |
| 部署形态 | WP 插件挂 WP | 自托管 Node 服务 | **单租户一仓一部署** | 简单可控，「快速建立」靠最小配置 + 一键启动 |
| 目标用户 | 中小站长 | 技术型商家 | **技术型商家 + 运营** | 后台 API 优先，运营前端后补 |

---

## 2. 逐项决策矩阵

### 2.1 产品模型

| 能力 | Woo | Medusa | 取舍 | 理由 |
|---|---|---|---|---|
| 产品类型 | simple/virtual/grouped/variable 四类 | Product + Variant 强统一 | **Woo 四类型 + Medusa Variant 统一表达** | 企业站（simple/virtual）与商城（variable）兼顾；统一用 Variant 承载 SKU/库存/价，simple 退化为单 variant |
| 规格/选项 | attributes + variations | options + variants | **Variant + option_value 关联** | 借鉴 Medusa 的 option→variant 结构，更清晰 |
| 虚拟/可下载 | 内置 downloadable | 需自定义 | **内置 virtual 标记 + media 关联** | 外贸数字样本/图册常用 |

**决策 D-PRODUCT**：四类型语义保留，物理结构统一为 `product 1—N variant`，`product_type` 字段控制行为（是否展示规格、是否计重/计运）。

### 2.2 多区域 / 多币种 / 多语言

| 能力 | Woo | Medusa | 取舍 | 理由 |
|---|---|---|---|---|
| 多区域 | 弱（靠插件/多站） | Region 模型（货币+税+支付+物流捆绑） | **简化 Region** | Medusa Region 过度抽象；外贸通常「站点=语言+币种」组合 |
| 多币种定价 | 货币切换插件 + 汇率 | price_set 按 region | **按币种独立存价** | 外贸多市场需精准差异化定价，汇率换算不满足 B2B 报价 |
| 多语言 | WPML/Polylang 插件 | i18n 插件 + 翻译表 | **主表 + 翻译表** | 查询清晰、可索引、扩展性好 |

**决策 D-REGION**：`region = (locale, currency, tax_config, payment_providers[], shipping_options[])` 扁平多对多；价格存 `variant_price(variant_id, currency_code, amount)`；可翻译字段拆 `{entity}_translation(entity_id, locale)`。

### 2.3 CMS / 博客 / 页面

| 能力 | Woo | Medusa | 取舍 | 理由 |
|---|---|---|---|---|
| 页面/博客 | 强（WP 内核） | 弱（靠第三方插件） | **内置 Page + Blog + Block + Menu** | 企业站刚需，不能依赖插件 |
| 内容块 | Gutenberg block | 无 | **简化 Block（结构化片段）** | 不做 Gutenberg 全套，做够用的结构化 block（富文本/图/CTA/列表） |
| 导航菜单 | 强 | 弱 | **内置 Menu + MenuItem** | 多语言菜单是外贸站标配 |

**决策 D-CMS**：内置轻量 CMS，覆盖企业站 90% 落地页/关于/联系/博客需求；不做 WP 全套主题系统。

### 2.4 交易链路

| 能力 | Woo | Medusa | 取舍 | 理由 |
|---|---|---|---|---|
| 购物车 | 内置 cart | Cart 独立资源 | **Cart 独立资源** | 支持 API 多端操作、Abandoned Cart 召回 |
| 结账 | 内置 checkout | Checkout 独立分步 | **Checkout 独立分步** | 支持多步、询盘转单、地址/物流/支付分步 |
| 订单 | 内置 order | Order + Fulfillment | **Order + Fulfillment + Return/Swap** | 借鉴 Medusa 履约/退换分离 |
| 询盘 | 表单插件 | 无 | **内置 Inquiry 模块** | 外贸 B2B lead 核心通道，可转单 |

**决策 D-TRADE**：Cart/Checkout/Order 分离，Inquiry 可一键转 Cart/Order。

### 2.5 支付 / 物流 / 优惠

| 能力 | Woo | Medusa | 取舍 | 理由 |
|---|---|---|---|---|
| 支付 | 大量网关插件 | Provider 接口 + 内置 stripe | **Provider 接口 + 预留实现** | MVP 内置占位接口，按需接 Stripe/PayPal/本地支付 |
| 物流 | 运费表 + 插件 | Shipping Profile + Option | **Shipping Profile + Option** | 借鉴 Medusa，按 profile 分组运费 |
| 优惠/券 | 复杂 coupon | Discount + Rule | **简化 Discount + Rule** | 不做 Woo 全套营销，做满减/百分比/包邮/买X送Y 基础规则 |

**决策 D-PAY-SHIP**：Provider 接口化，MVP 内置占位 + 1 个示例（Stripe 占位）；优惠做基础规则集，复杂玩法留事件钩子。

### 2.6 事件 / 扩展

| 能力 | Woo | Medusa | 取舍 | 理由 |
|---|---|---|---|---|
| 事件机制 | hooks（同步） | Event Bus + Subscriber | **Redis 事件总线 + Subscriber** | 解耦邮件/库存/统计，Medusa 风格，异步可扩 |
| 插件 | 庞大市场 | 模块化 | **内置模块 + 事件钩子 + 预留插件接口** | 不做市场，但留扩展点（事件 + Webhook + Provider 接口） |
| Webhook | 插件 | 内置 | **内置 Webhook** | 让外部系统订阅关键事件（订单/询盘） |

**决策 D-EXT**：Redis Pub/Sub 事件总线 + asynq 异步任务 + Webhook 出口；不做插件市场。

### 2.7 鉴权 / 权限

| 能力 | Woo | Medusa | 取舍 | 理由 |
|---|---|---|---|---|
| 管理员鉴权 | WP role/cap | JWT + RBAC | **JWT(access+refresh) + RBAC** | headless 需无状态 token；refresh 存 Redis 可吊销 |
| 客户鉴权 | WP 用户 | Customer + JWT | **Customer 独立体系 + JWT** | 后台管理员与前台客户两套隔离 |
| 权限粒度 | capability | permission | **资源 + 操作 permission** | 细到 (resource, action)，运营/管理员分级 |

**决策 D-AUTH**：两套 JWT 体系（admin / customer），RBAC 细到 (resource, action)，详见 `docs/06-auth-rbac.md`。

### 2.8 后台 UI

| 能力 | Woo | Medusa | 取舍 | 理由 |
|---|---|---|---|---|
| 管理界面 | 内置 WP-Admin | 分离 admin-ui | **本期不做，API-first** | 范围控制，后续独立前端工程对接 |

**决策 D-UI**：本期仅后端 API；前端后续单独立项。

---

## 3. 「不照搬」的具体减法清单

明确**不做**的 Woo/Medusa 能力，避免范围蔓延：

| 来源 | 不做项 | 原因 |
|---|---|---|
| Woo | 完整主题系统 / Gutenberg 全套 block | 与 headless 定位冲突 |
| Woo | 庞大插件市场 | 范围/维护成本 |
| Woo | 复杂营销（拼团/分销/积分商城） | 预留事件钩子，不内置 |
| Medusa | Region 过度抽象（currency+tax+payment+shipping 强捆绑的复杂继承） | 外贸实际用扁平组合更易理解 |
| Medusa | price_list 高级规则引擎 | 按币种独立存价已满足核心需求 |
| Medusa | Sales Channel 多渠道库存隔离 | 单租户单站，MVP 不需要 |
| 两者 | 多租户 SaaS | 本期单租户 |

---

## 4. 「重组」的关键创新点

| 创新点 | 说明 |
|---|---|
| 统一数据模型 + 功能开关 | 一套模型服务企业站/商城，靠 `store.site_type` + `feature_flags` 裁剪 API 暴露面与行为 |
| 询盘 ↔ 订单互通 | Inquiry 可一键转 Cart/Order，打通 B2B lead → 成交 |
| 内置外贸 i18n/多币种 | 翻译表 + 按币种存价，外贸刚需开箱即用 |
| 事件总线 + Webhook 双出口 | 内部 Subscriber + 外部 Webhook，兼顾自研副作用与第三方集成 |

---

## 5. 取舍决策与后续文档映射

| 决策 | 落地文档 |
|---|---|
| D-PRODUCT | `docs/03-data-model.md` §产品域、`docs/05-modules.md` §产品 |
| D-REGION | `docs/03` §区域/价格/翻译、`docs/05` §区域 |
| D-CMS | `docs/03` §CMS、`docs/05` §CMS |
| D-TRADE | `docs/03` §交易域、`docs/05` §交易 |
| D-PAY-SHIP | `docs/03` §支付/物流/优惠、`docs/05` §支付物流 |
| D-EXT | `docs/07-events-jobs.md`、`docs/05` §扩展点 |
| D-AUTH | `docs/06-auth-rbac.md` |
| D-UI | `docs/08-roadmap.md`（本期不含 UI） |

关键决策另以 ADR 记录于 `docs/adr/`。
