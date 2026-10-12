package handler

import (
	"github.com/gin-gonic/gin"

	"onehost/internal/config"
)

// RegisterRoutes 注册全部 API 路由。
func RegisterRoutes(r *gin.Engine, h *Handler, cfg *config.Config) {
	api := r.Group("/api/v1")

	// 公开
	api.GET("/settings/public", h.PublicSettings)
	api.POST("/auth/register", h.Register)
	api.POST("/auth/login", h.Login)
	api.GET("/products", h.ListProducts)
	api.GET("/products/:id", h.GetProduct)
	api.GET("/pay/notify", h.PayNotify)
	api.POST("/pay/notify", h.PayNotify)
	api.GET("/pay/return", h.PayReturn)

	// 用户（需登录）
	u := api.Group("")
	u.Use(h.AuthRequired())
	u.GET("/user/profile", h.Profile)
	u.GET("/user/balance/logs", h.BalanceLogs)
	u.POST("/user/voucher/redeem", h.RedeemVoucher)
	u.POST("/user/recharge", h.CreateRecharge)
	u.GET("/user/recharge/orders", h.RechargeOrders)
	u.POST("/user/recharge/:id/mockpay", h.MockPay)
	u.GET("/user/orders", h.ListMyOrders)
	u.POST("/user/orders", h.CreateOrder)
	u.POST("/user/orders/:id/pay", h.PayOrder)
	u.POST("/user/orders/:id/cancel", h.CancelOrder)
	u.POST("/user/orders/:id/renew", h.RenewOrder)
	u.GET("/user/instances", h.ListMyInstances)
	u.POST("/user/instances/:id/start", h.StartInstance)
	u.POST("/user/instances/:id/stop", h.StopInstance)
	u.POST("/user/instances/:id/restart", h.RestartInstance)
	u.DELETE("/user/instances/:id", h.DestroyInstance)
	// 控制台：WebSSH / VNC / SFTP
	u.GET("/user/instances/:id/console", h.InstanceConsoleInfo)
	u.GET("/user/instances/:id/ssh", h.InstanceSSH)
	u.GET("/user/instances/:id/vnc", h.InstanceVNC)
	u.GET("/user/instances/:id/files", h.SFTPList)
	u.GET("/user/instances/:id/files/download", h.SFTPDownload)
	u.POST("/user/instances/:id/files/upload", h.SFTPUpload)
	u.POST("/user/instances/:id/files/mkdir", h.SFTPMkdir)
	u.DELETE("/user/instances/:id/files", h.SFTPDelete)

	// 管理后台
	a := api.Group("/admin")
	a.Use(h.AuthRequired(), h.AdminRequired())
	a.GET("/stats", h.AdminStats)
	a.GET("/products", h.AdminListProducts)
	a.POST("/products", h.AdminCreateProduct)
	a.PUT("/products/:id", h.AdminUpdateProduct)
	a.DELETE("/products/:id", h.AdminDeleteProduct)
	a.GET("/orders", h.AdminListOrders)
	a.GET("/recharge/orders", h.AdminListRechargeOrders)
	a.GET("/instances", h.AdminListInstances)
	a.GET("/users", h.AdminListUsers)
	a.PUT("/users/:id/status", h.AdminUpdateUserStatus)
	a.POST("/users/:id/balance", h.AdminAdjustBalance)
	a.GET("/nodes", h.AdminListNodes)
	a.POST("/nodes", h.AdminCreateNode)
	a.PUT("/nodes/:id", h.AdminUpdateNode)
	a.DELETE("/nodes/:id", h.AdminDeleteNode)
	a.POST("/nodes/:id/test", h.AdminTestNode)
	a.GET("/vouchers", h.AdminListVouchers)
	a.POST("/vouchers/generate", h.AdminGenerateVouchers)
	a.POST("/vouchers/:id/void", h.AdminVoidVoucher)
	a.GET("/settings", h.AdminGetSettings)
	a.PUT("/settings", h.AdminUpdateSettings)
}
