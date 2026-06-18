# 06 - 鉴权与 RBAC 设计

> 状态：规范基线
> 实体见 `docs/03-data-model.md` §3.4，端点见 `docs/04-api-design.md` §6。

---

## 1. 总则

- **两套隔离体系**：Admin（后台运营）、Customer（前台消费者），独立 JWT 密钥、独立登录端点、独立 RBAC 范围。
- **无状态 access + 可吊销 refresh**：access 短 TTL 不入库；refresh 长 TTL 存 Redis（可主动吊销/登出/过期）。
- **密码安全**：bcrypt（cost ≥12）；禁止明文/MD5/SHA1；禁止打印密钥与密码（`AGENTS.md` §5）。
- **密钥来源**：`JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` 走 `.env`，强随机，禁止弱默认值提交。

---

## 2. Admin 鉴权

### 2.1 登录与令牌
- `POST /api/v1/admin/auth/login`（email + password）→ `{ access_token, refresh_token, expires_in, admin }`。
- 失败统一 `AUTH_INVALID_CREDENTIALS`（401），不区分「用户不存在/密码错」防枚举。
- 登录成功写 `last_login_at`，发布 `admin.logged_in`（审计）。

### 2.2 access token（admin）
- claims：`{ sub: admin_id, typ: "admin", roles: [slug...], perms: ["product:write", ...], iat, exp }`。
- TTL：`JWT_ACCESS_TTL`（默认 15m）。
- 校验：中间件解析 → 注入 `AdminContext{ID, Roles, Perms}` 到 Fiber Locals。
- 不入库，自然过期。

### 2.3 refresh token（admin）
- claims：`{ sub: admin_id, typ: "admin_refresh", jti, iat, exp }`。
- TTL：`JWT_REFRESH_TTL`（默认 30d）。
- 存储：`redis: admin:refresh:{jti} = admin_id`，TTL 同 refresh；登出/吊销删 key。
- 刷新：`POST /admin/auth/refresh` 校验签名 + Redis 存在 → 签新 access + 新 refresh（旋转，旧 jti 失效）。
- 盗用检测：刷新时若 jti 存在但已用于旋转 → 视为被盗，吊销该 admin 全部 refresh（删 `admin:refresh:admin:{id}:*`）。

### 2.4 登出与吊销
- `POST /admin/auth/logout` → 删当前 refresh jti；access 自然过期（短 TTL 可接受）。
- 管理员强制吊销某用户：删该 admin 全部 refresh key；如需即时失效 access，引入 token 版本号（`admin.token_version`，变更后 access 失效，MVP 可选）。

### 2.5 me
- `GET /admin/auth/me` → 当前 admin profile + roles + perms。

---

## 3. Customer 鉴权（Storefront）

### 3.1 注册 / 登录
- `POST /api/v1/store/auth/register`（email, password, 可选 profile）。
- `POST /api/v1/store/auth/login`（email + password）。
- 询盘/下单可 guest（无账号）；注册时同邮箱去重合并。
- 密码同样 bcrypt。

### 3.2 令牌
- 结构同 admin，但 `typ: "customer"` / `"customer_refresh"`，独立密钥，独立 Redis 命名空间 `customer:refresh:{jti}`。
- claims 含 `customer_id`、`default_currency`、`default_locale`，无 RBAC perms。
- 公开读（产品/页面/博文）无需 token；写（下单/询盘/地址）需 token 或 guest 标识。

### 3.3 guest → registered
- guest 下单/询盘留 email；注册后历史订单/询盘按 email 自动归属（异步合并任务）。

---

## 4. RBAC 模型（Admin）

### 4.1 实体关系
```
admin_user N—N role N—N permission
   └ admin_role         └ role_permission
permission = (resource, action)
```
- `permission.resource`：`product`/`variant`/`order`/`customer`/`cms`/`inquiry`/`settings`/`media`/`discount`/`payment`/`shipping`/`region`/`admin`/`webhook`/...
- `permission.action`：`read`/`write`/`delete`/`publish`/`export`/`approve`...

### 4.2 权限编码
- 字符串形如 `<resource>:<action>`：`product:write`、`order:approve`、`inquiry:assign`、`settings:write`。
- 权限表预置全集（见 §6）；角色挂载子集。

### 4.3 内置角色（is_system）
| 角色 | 含义 | 典型权限 |
|---|---|---|
| `owner` | 站点所有者 | 全权限（is_system，不可删） |
| `admin` | 全功能运营 | 除 `admin:*` 与 `settings:dangerous` 外全部 |
| `operator` | 运营 | product/order/customer/cms/inquiry 读写 |
| `content` | 内容编辑 | cms/blog/page/menu 读写，product 读 |
| `support` | 客服 | order/inquiry/customer 读改，无财务删除 |
| `viewer` | 只读 | 全域 read |

> 系统角色由 seed 迁移写入，可改挂载权限但不可删；自定义角色由 owner/admin 创建。

### 4.4 授权执行
- 中间件 `RBAC(required string...)`：路由声明所需权限，如 `router.Post("/products", RBAC("product:write"), handler)`。
- 校验：`AdminContext.Perms ⊇ required` 否则 403 `AUTH_FORBIDDEN`。
- 多权限语义：声明多个为「需全部满足」；「或」用 `RBACAny(...)`。

### 4.5 数据级权限（MVP 简化）
- MVP 不做行级数据隔离（单租户全员见全店数据）；预留 `inquiry.assigned_admin_id` 的「我的询盘」过滤为业务层过滤，非 RBAC。

---

## 5. 密码与安全策略

- bcrypt cost ≥12；登录失败计数 `admin:loginfail:{email}`（Redis，5 次锁 15 分钟）。
- 密码重置：`POST /admin/auth/password-reset/request`（发邮件 token，Redis 30 分钟）→ `POST .../confirm`（token + 新密码）。
- 修改密码后旋转 refresh（吊销旧 refresh）。
- 禁止把 token/密码写日志、写错误响应细节。

---

## 6. 权限全集（seed 清单）

| resource | actions |
|---|---|
| product | read, write, delete, publish, archive |
| variant | read, write |
| category | read, write, delete |
| collection | read, write, delete |
| order | read, write, approve, cancel, fulfill, refund, export |
| customer | read, write, delete |
| inquiry | read, assign, update, convert, export |
| cms | read, write, delete, publish |
| settings | read, write, dangerous |
| media | read, write, delete |
| region | read, write |
| discount | read, write, delete |
| payment | read, write |
| shipping | read, write |
| admin | read, write, delete |
| webhook | read, write, delete |
| notification | read, write |

> `settings:dangerous` 用于删库/重置等高危操作，仅 owner。

---

## 7. 中间件链（鉴权相关，顺序）

```
Recover → RequestID → Logger → CORS → RateLimit
  → JWTAuth（解析 access；公开路由跳过）
  → RBAC（校验所需权限；公开/已登录即可的路由跳过）
  → handler
```
- `JWTAuth` 失败 401 `AUTH_TOKEN_EXPIRED` / `AUTH_TOKEN_INVALID` / `AUTH_TOKEN_MISSING`。
- `RBAC` 失败 403 `AUTH_FORBIDDEN`。

---

## 8. 与 ent / 迁移

- `admin_user`、`role`、`permission`、`admin_role`、`role_permission` 五表 + seed。
- seed：权限全集 + 系统角色挂载，由 Atlas 迁移或启动时 `bootstrap` 幂等写入。
- `password_hash` 列禁止出现在任何 API 响应/日志/导出。
