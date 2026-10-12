package model

import "time"

// 金额统一以「分」为单位存储，避免浮点误差。

// User 用户
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"size:128;uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"size:128;not null" json:"-"`
	Nickname  string    `gorm:"size:64" json:"nickname"`
	Role      string    `gorm:"size:16;default:user" json:"role"` // user / admin
	Status    int       `gorm:"default:1" json:"status"`          // 1 正常 0 禁用
	Balance   int64     `gorm:"default:0" json:"balance"`         // 余额（分）
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BalanceLog 余额流水
type BalanceLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index" json:"user_id"`
	Change       int64     `json:"change"` // 正为增加，负为减少
	BalanceAfter int64     `json:"balance_after"`
	Type         string    `gorm:"size:32" json:"type"` // recharge / consume / refund / voucher / admin / renew
	Remark       string    `gorm:"size:255" json:"remark"`
	CreatedAt    time.Time `json:"created_at"`
}

// Product 产品（机型）
type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Price       int64     `gorm:"not null" json:"price"` // 每周期价格（分）
	PeriodDays  int       `gorm:"default:30" json:"period_days"`
	CPU         int       `json:"cpu"`
	MemoryMB    int       `json:"memory_mb"`
	DiskGB      int       `json:"disk_gb"`
	TrafficGB   int       `json:"traffic_gb"`
	Image       string    `gorm:"size:128" json:"image"` // 系统镜像别名，如 images:debian/12
	Stock       int       `gorm:"default:0" json:"stock"`     // 库存，-1 表示不限
	MaxPerUser  int       `gorm:"default:0" json:"max_per_user"` // 每人限购，0 表示不限
	NodeID      uint      `json:"node_id"`
	Status      int       `gorm:"default:1" json:"status"` // 1 上架 0 下架
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Order 订单
type Order struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	OrderNo     string     `gorm:"size:64;uniqueIndex" json:"order_no"`
	UserID      uint       `gorm:"index" json:"user_id"`
	ProductID   uint       `json:"product_id"`
	ProductName string     `gorm:"size:128" json:"product_name"`
	Quantity    int        `gorm:"default:1" json:"quantity"`
	Amount      int64      `json:"amount"` // 应付金额（分）
	PeriodDays  int        `json:"period_days"`
	Spec        string     `gorm:"size:255" json:"spec"`
	Status      string     `gorm:"size:16;default:unpaid" json:"status"` // unpaid/paid/provisioning/active/failed/expired/cancelled
	InstanceID  uint       `json:"instance_id"`
	Remark      string     `gorm:"size:255" json:"remark"`
	CreatedAt   time.Time  `json:"created_at"`
	PaidAt      *time.Time `json:"paid_at"`
	ExpiredAt   *time.Time `json:"expired_at"`
}

// Instance 云主机实例
type Instance struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Name      string     `gorm:"size:64" json:"name"`
	UserID    uint       `gorm:"index" json:"user_id"`
	OrderID   uint       `json:"order_id"`
	Provider  string     `gorm:"size:16" json:"provider"` // lxd / incus / mock
	NodeID    uint       `json:"node_id"`
	RemoteID  string     `gorm:"size:128" json:"remote_id"` // 远端实例名/ID
	IP        string     `gorm:"size:64" json:"ip"`
	Status    string     `gorm:"size:16" json:"status"` // creating/running/stopped/error/deleted
	CPU       int        `json:"cpu"`
	MemoryMB  int        `json:"memory_mb"`
	DiskGB    int        `json:"disk_gb"`
	Image     string     `gorm:"size:128" json:"image"`
	ExpiredAt *time.Time `json:"expired_at"`

	// 控制台接入信息（WebSSH / SFTP / VNC）
	SSHPort     int    `gorm:"default:22" json:"ssh_port"`     // 实例 SSH 端口
	SSHUser     string `gorm:"size:64;default:root" json:"ssh_user"` // SSH 登录用户
	SSHPassword string `gorm:"size:128" json:"-"`              // SSH 登录口令（不下发前端）
	VNCPort     int    `json:"vnc_port"`                       // VNC 端口，0 表示未启用

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Node 虚拟化节点
type Node struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"size:64" json:"name"`
	Type       string    `gorm:"size:16;default:mock" json:"type"` // lxd / incus / mock
	APIURL     string    `gorm:"size:255" json:"api_url"`          // 如 https://10.0.0.10:8443
	Cert       string    `gorm:"type:text" json:"-"`               // 客户端证书 PEM
	Key        string    `gorm:"type:text" json:"-"`               // 客户端私钥 PEM
	ServerCert string    `gorm:"type:text" json:"-"`               // 服务端证书 PEM（用于校验/固定）
	IPPool     string    `gorm:"size:255" json:"ip_pool"`          // 预留 IPv4 池（逗号分隔）
	Status     int       `gorm:"default:1" json:"status"`          // 1 启用 0 停用
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Voucher 代金券
type Voucher struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Code      string     `gorm:"size:32;uniqueIndex" json:"code"`
	Amount    int64      `json:"amount"` // 面额（分）
	Status    int        `gorm:"default:0" json:"status"` // 0 未使用 1 已使用 2 作废
	UsedBy    uint       `json:"used_by"`
	UsedAt    *time.Time `json:"used_at"`
	ExpiredAt *time.Time `json:"expired_at"`
	Batch     string     `gorm:"size:64" json:"batch"`
	CreatedAt time.Time  `json:"created_at"`
}

// RechargeOrder 充值订单
type RechargeOrder struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	OrderNo   string     `gorm:"size:64;uniqueIndex" json:"order_no"`
	UserID    uint       `gorm:"index" json:"user_id"`
	Amount    int64      `json:"amount"` // 充值金额（分）
	Method    string     `gorm:"size:16" json:"method"` // alipay / wxpay
	Status    string     `gorm:"size:16;default:unpaid" json:"status"` // unpaid / paid / failed
	TradeNo   string     `gorm:"size:64" json:"trade_no"` // 支付网关交易号
	CreatedAt time.Time  `json:"created_at"`
	PaidAt    *time.Time `json:"paid_at"`
}

// Setting 系统配置（键值对）
type Setting struct {
	Key   string `gorm:"column:config_key;primaryKey;size:64" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}
