# OneHost MVP —— 云计算售卖与管理平台（核心售卖闭环）

对标 [qdmz/onehost](https://github.com/qdmz/onehost) 的核心商业化闭环，用 **Go + Vue3** 重写的一个可运行 MVP。

一句话：把虚拟化资源当成商品卖出去 —— **产品定价 → 用户下单 → 在线支付 → 自动开通 → 自助运维 → 到期续费**。

> 说明：onehost 是 40+ 模块、1700+ 接口、62 张表的完整发行版；本项目是**聚焦售卖闭环的 MVP**，刻意收敛范围，先跑通主干，便于后续按模块扩展。

---

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.22、Gin、GORM、JWT、MySQL |
| 前端 | Vue 3、Vite、Element Plus、Pinia、Vue Router、vue-i18n |
| 虚拟化 | LXD / Incus（REST over TLS，客户端证书）、内置 Mock |
| 支付 | 易支付（第四方网关，MD5 签名 + 异步回调） |
| 部署 | Docker Compose（web + api + mysql 三容器） |

---

## 已实现功能

**用户端**
- 注册 / 登录（JWT）
- 云主机商城：浏览在售产品、实时库存、按数量下单
- 我的订单：余额支付、取消、续费
- 钱包：余额、在线充值（易支付）、余额流水、未接网关时「模拟支付」
- 代金券兑换（事务 + 行锁，防重复兑换）
- 我的实例：开机 / 关机 / 重启 / 销毁

**管理后台**
- 平台总览：用户数、订单数、实例数、累计收入、节点数
- 产品管理：CRUD、定价、库存、每人限购、节点绑定、上下架
- 订单管理：全量订单 + 状态筛选 + 用户邮箱
- 用户管理：启停、调整余额（带流水）
- 节点管理：接入 LXD/Incus/Mock、配置客户端证书、连通性测试
- 代金券：批量生成、作废
- 系统设置：站点名称、注册开关、易支付配置

---

## 目录结构

```
onehost-mvp/
├── docker-compose.yaml
├── .env.example
├── api/                     # Go 后端
│   ├── Dockerfile
│   ├── go.mod
│   ├── main.go
│   ├── config/config.yaml
│   └── internal/
│       ├── config/          # 配置加载（yaml + ONEHOST_* 环境变量）
│       ├── model/           # GORM 模型（User/Product/Order/Instance/Node/Voucher/...）
│       ├── database/        # 连接、AutoMigrate、种子数据
│       ├── util/            # JWT、bcrypt、响应、随机码
│       ├── provider/        # 虚拟化后端：接口 + LXD/Incus + Mock
│       ├── service/         # 易支付签名与验签
│       └── handler/         # 认证/钱包/代金券/订单/实例/管理后台/路由
└── web/                     # Vue3 前端
    ├── Dockerfile
    ├── nginx.conf
    └── src/
        ├── api/  store/  router/  i18n/
        ├── layouts/         # 控制台 / 管理后台布局
        └── views/           # console/* 与 admin/*
```

---

## 快速开始（Docker Compose）

```bash
cp .env.example .env      # 按需修改数据库密码、JWT密钥、PUBLIC_URL
docker compose up -d --build
```

访问 `http://localhost`（默认 80 端口，可用 `.env` 的 `WEB_PORT` 修改）。

**默认管理员账号：`admin@onehost.local` / `admin123`**（首次启动自动创建，请尽快修改）。

### 试跑一遍完整闭环（无需 LXD，默认使用 Mock 节点）
1. 管理员登录 → 确认「节点管理」里已有 `本地 Mock 节点`，并在「产品管理」有一个在售产品；
2. 注册一个普通用户（或用户端登录）；
3. 「钱包」→ 生成充值订单 → 点「模拟支付」到账；或在后台给用户调整余额；
4. 「云主机商城」→ 立即购买 → 「我的订单」→ 余额支付；
5. 系统异步开通实例，状态变为「已开通」；在「我的实例」中开机/关机/重启/销毁。

---

## 本地开发

```bash
# 后端（需要 Go 1.22+ 与本地 MySQL，数据库名 onehost）
cd api && go mod tidy && go run main.go        # http://localhost:8888

# 前端（Node 18+）
cd web && npm install && npm run dev           # http://localhost:8080，已代理 /api -> :8888
```

---

## 接入真实 LXD / Incus 节点

1. 在目标主机启用 LXD/Incus 的 HTTPS API：`lxc config set core.https_address :8443`；
2. 生成客户端证书并信任：
   ```bash
   lxc remote add onehost https://<主机IP>:8443 --accept-certificate --auth-type=tls
   # 或使用 lxc config trust add client.crt
   ```
3. 后台「节点管理」→ 接入节点：类型选 `lxd` 或 `incus`，填 `https://<主机IP>:8443`，
   粘贴**客户端证书**与**私钥**（PEM）。若填写**服务端证书**则启用严格校验，否则跳过校验（仅建议内网/测试）；
4. 点「测试」确认连通，然后在「产品管理」把产品绑定到该节点，镜像别名填例如 `images:debian/12`。

> 开通即调用 LXD/Incus 的 `/1.0/instances` 创建容器并按规格设置 CPU/内存，随后启动、回填 IPv4。

---

## 接入易支付

后台「系统设置」填写：网关地址（如 `https://pay.example.com`）、商户号 `PID`、商户密钥 `KEY`，并开启「启用易支付」。

- 签名规则：参数按键名升序拼成 `a=b&c=d`，末尾拼接 `KEY` 后取**小写 MD5**；
- 异步回调：`{PUBLIC_URL}/api/v1/pay/notify`；同步回跳：`{PUBLIC_URL}/api/v1/pay/return`；
- 回调验签通过且 `trade_status=TRADE_SUCCESS` 时，自动为用户余额入账。

---

## 金额与时间约定

- 金额统一以 **分（int64）** 存储与传输，前端展示时除以 100；后台产品价格、用户余额调整、充值输入框以「元」录入并自动换算。
- 时间使用服务器本地时区（容器内设为 `Asia/Shanghai`），交付到前端为 ISO8601。

---

## 与完整版 onehost 的差距（后续可扩展方向）

本 MVP 未覆盖（已在架构上预留扩展点）：多虚拟化后端（Docker/PVE/KubeVirt 等）、VNC/WebSSH 控制台、
快照与流量/性能监控、工单、KYC 实名、邀请码、OAuth2、兑换码、定时任务与日志审计、国际化全量文案等。

---

## 安全提示

- 生产环境务必修改 `.env` 中的 `MYSQL_PASSWORD` 与 `JWT_SECRET`；
- 默认管理员密码 `admin123` 请立即修改；
- 节点客户端私钥、易支付 KEY 属敏感信息，勿提交版本库（见 `.gitignore`）。
