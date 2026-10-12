package handler

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"onehost/internal/model"
	"onehost/internal/provider"
	"onehost/internal/util"
)

type createOrderReq struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

// CreateOrder 下单（不扣款，支付时才扣余额与库存）。
func (h *Handler) CreateOrder(c *gin.Context) {
	uid := UID(c)
	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, "参数错误")
		return
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	var p model.Product
	if err := h.DB.First(&p, req.ProductID).Error; err != nil {
		util.Fail(c, "产品不存在")
		return
	}
	if p.Status != 1 {
		util.Fail(c, "产品已下架")
		return
	}
	if p.Stock >= 0 && p.Stock < req.Quantity {
		util.Fail(c, "库存不足")
		return
	}
	if p.MaxPerUser > 0 {
		var cnt int64
		h.DB.Model(&model.Order{}).
			Where("user_id = ? AND product_id = ? AND status IN ?", uid, p.ID,
				[]string{"paid", "provisioning", "active"}).
			Count(&cnt)
		if int(cnt)+req.Quantity > p.MaxPerUser {
			util.Fail(c, "超出每人限购数量")
			return
		}
	}

	order := model.Order{
		OrderNo:     util.GenOrderNo("P"),
		UserID:      uid,
		ProductID:   p.ID,
		ProductName: p.Name,
		Quantity:    req.Quantity,
		Amount:      p.Price * int64(req.Quantity),
		PeriodDays:  p.PeriodDays,
		Spec:        specText(p, req.Quantity),
		Status:      "unpaid",
	}
	if err := h.DB.Create(&order).Error; err != nil {
		util.Fail(c, "下单失败")
		return
	}
	util.OK(c, order)
}

func specText(p model.Product, qty int) string {
	return fmt.Sprintf("%d × [%d核 %dG内存 %dG硬盘]", qty, p.CPU, p.MemoryMB/1024, p.DiskGB)
}

// ListMyOrders 我的订单。
func (h *Handler) ListMyOrders(c *gin.Context) {
	var list []model.Order
	h.DB.Where("user_id = ?", UID(c)).Order("id desc").Limit(200).Find(&list)
	util.OK(c, list)
}

// PayOrder 余额支付并触发异步开通。
func (h *Handler) PayOrder(c *gin.Context) {
	uid := UID(c)
	var order model.Order
	if err := h.DB.Where("id = ? AND user_id = ?", c.Param("id"), uid).First(&order).Error; err != nil {
		util.Fail(c, "订单不存在")
		return
	}
	if order.Status != "unpaid" {
		util.Fail(c, "订单状态不可支付")
		return
	}
	var p model.Product
	if err := h.DB.First(&p, order.ProductID).Error; err != nil {
		util.Fail(c, "产品不存在")
		return
	}

	now := time.Now()
	expiredAt := now.AddDate(0, 0, order.PeriodDays)

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if p.Stock >= 0 {
			res := tx.Model(&model.Product{}).
				Where("id = ? AND stock >= ?", p.ID, order.Quantity).
				UpdateColumn("stock", gorm.Expr("stock - ?", order.Quantity))
			if res.RowsAffected == 0 {
				return fmt.Errorf("库存不足")
			}
		}
		if _, err := applyBalanceTx(tx, uid, -order.Amount, "consume", "购买 "+order.ProductName); err != nil {
			return err
		}
		return tx.Model(&order).Updates(map[string]interface{}{
			"status": "provisioning", "paid_at": now, "expired_at": expiredAt,
		}).Error
	})
	if err != nil {
		util.Fail(c, err.Error())
		return
	}

	go h.provision(order.ID)
	util.OK(c, gin.H{"order_no": order.OrderNo})
}

// CancelOrder 取消未支付订单。
func (h *Handler) CancelOrder(c *gin.Context) {
	uid := UID(c)
	var order model.Order
	if err := h.DB.Where("id = ? AND user_id = ?", c.Param("id"), uid).First(&order).Error; err != nil {
		util.Fail(c, "订单不存在")
		return
	}
	if order.Status != "unpaid" {
		util.Fail(c, "仅未支付订单可取消")
		return
	}
	h.DB.Model(&order).Update("status", "cancelled")
	util.OK(c, nil)
}

// RenewOrder 续费。
func (h *Handler) RenewOrder(c *gin.Context) {
	uid := UID(c)
	var order model.Order
	if err := h.DB.Where("id = ? AND user_id = ?", c.Param("id"), uid).First(&order).Error; err != nil {
		util.Fail(c, "订单不存在")
		return
	}
	if order.Status != "active" {
		util.Fail(c, "仅已开通订单可续费")
		return
	}
	var p model.Product
	if err := h.DB.First(&p, order.ProductID).Error; err != nil {
		util.Fail(c, "产品不存在")
		return
	}
	var inst model.Instance
	if err := h.DB.First(&inst, order.InstanceID).Error; err != nil {
		util.Fail(c, "实例不存在")
		return
	}

	now := time.Now()
	base := now
	if inst.ExpiredAt != nil && inst.ExpiredAt.After(now) {
		base = *inst.ExpiredAt
	}
	newExp := base.AddDate(0, 0, p.PeriodDays)

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := applyBalanceTx(tx, uid, -p.Price, "renew", "续费 "+p.Name); err != nil {
			return err
		}
		if err := tx.Model(&inst).Update("expired_at", newExp).Error; err != nil {
			return err
		}
		return tx.Model(&order).Update("expired_at", newExp).Error
	})
	if err != nil {
		util.Fail(c, err.Error())
		return
	}
	util.OK(c, gin.H{"expired_at": newExp})
}

// ---------------- 实例操作 ----------------

func (h *Handler) instanceForUser(c *gin.Context) (*model.Instance, bool) {
	var inst model.Instance
	if err := h.DB.Where("id = ? AND user_id = ?", c.Param("id"), UID(c)).First(&inst).Error; err != nil {
		util.Fail(c, "实例不存在")
		return nil, false
	}
	return &inst, true
}

func (h *Handler) providerForInstance(inst *model.Instance) (provider.Provider, error) {
	var node model.Node
	if err := h.DB.First(&node, inst.NodeID).Error; err != nil {
		return nil, fmt.Errorf("节点不存在")
	}
	return provider.ForNode(&node)
}

// ListMyInstances 我的实例。
func (h *Handler) ListMyInstances(c *gin.Context) {
	var list []model.Instance
	h.DB.Where("user_id = ? AND status <> ?", UID(c), "deleted").Order("id desc").Find(&list)
	util.OK(c, list)
}

func (h *Handler) StartInstance(c *gin.Context) {
	inst, ok := h.instanceForUser(c)
	if !ok {
		return
	}
	prov, err := h.providerForInstance(inst)
	if err != nil {
		util.Fail(c, err.Error())
		return
	}
	if err := prov.Start(c.Request.Context(), inst.RemoteID); err != nil {
		util.Fail(c, "开机失败: "+err.Error())
		return
	}
	h.DB.Model(inst).Update("status", "running")
	util.OK(c, nil)
}

func (h *Handler) StopInstance(c *gin.Context) {
	inst, ok := h.instanceForUser(c)
	if !ok {
		return
	}
	prov, err := h.providerForInstance(inst)
	if err != nil {
		util.Fail(c, err.Error())
		return
	}
	if err := prov.Stop(c.Request.Context(), inst.RemoteID); err != nil {
		util.Fail(c, "关机失败: "+err.Error())
		return
	}
	h.DB.Model(inst).Update("status", "stopped")
	util.OK(c, nil)
}

func (h *Handler) RestartInstance(c *gin.Context) {
	inst, ok := h.instanceForUser(c)
	if !ok {
		return
	}
	prov, err := h.providerForInstance(inst)
	if err != nil {
		util.Fail(c, err.Error())
		return
	}
	if err := prov.Restart(c.Request.Context(), inst.RemoteID); err != nil {
		util.Fail(c, "重启失败: "+err.Error())
		return
	}
	h.DB.Model(inst).Update("status", "running")
	util.OK(c, nil)
}

func (h *Handler) DestroyInstance(c *gin.Context) {
	inst, ok := h.instanceForUser(c)
	if !ok {
		return
	}
	prov, err := h.providerForInstance(inst)
	if err != nil {
		util.Fail(c, err.Error())
		return
	}
	if err := prov.Delete(c.Request.Context(), inst.RemoteID); err != nil {
		util.Fail(c, "销毁失败: "+err.Error())
		return
	}
	h.DB.Model(inst).Update("status", "deleted")
	util.OK(c, nil)
}

// ---------------- 开通流程 ----------------

func (h *Handler) provision(orderID uint) {
	var order model.Order
	if err := h.DB.First(&order, orderID).Error; err != nil {
		return
	}
	var p model.Product
	if err := h.DB.First(&p, order.ProductID).Error; err != nil {
		h.failProvision(&order, nil, nil, fmt.Errorf("产品不存在"))
		return
	}
	var node model.Node
	if err := h.DB.First(&node, p.NodeID).Error; err != nil {
		h.failProvision(&order, nil, &p, fmt.Errorf("节点不存在"))
		return
	}
	prov, err := provider.ForNode(&node)
	if err != nil {
		h.failProvision(&order, nil, &p, err)
		return
	}

	name := fmt.Sprintf("oh-%d-%s", order.UserID, strings.ToLower(util.RandCode(6)))
	isMock := node.Type == "mock" || node.Type == ""
	inst := model.Instance{
		Name: name, UserID: order.UserID, OrderID: order.ID,
		Provider: node.Type, NodeID: node.ID, Status: "creating",
		CPU: p.CPU, MemoryMB: p.MemoryMB, DiskGB: p.DiskGB, Image: p.Image,
		ExpiredAt: order.ExpiredAt,
	}
	// 控制台接入参数：mock 节点直连内置的真实 SSH/VNC 靶机；真实 lxd/incus 走实例 IP。
	if isMock {
		inst.SSHPort = envInt("ONEHOST_MOCK_SSH_PORT", 2222)
		inst.SSHUser = envStr("ONEHOST_MOCK_SSH_USER", "root")
		inst.SSHPassword = envStr("ONEHOST_MOCK_SSH_PASS", "onehost")
		inst.VNCPort = envInt("ONEHOST_MOCK_VNC_PORT", 5901)
	} else {
		inst.SSHPort = 22
		inst.SSHUser = "root"
		inst.SSHPassword = util.RandCode(16)
	}
	if err := h.DB.Create(&inst).Error; err != nil {
		h.failProvision(&order, nil, &p, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	remoteID, err := prov.Create(ctx, provider.Spec{
		Name: name, Image: p.Image, CPU: p.CPU, MemoryMB: p.MemoryMB, DiskGB: p.DiskGB,
	})
	if err != nil {
		h.failProvision(&order, &inst, &p, err)
		return
	}
	if err := prov.Start(ctx, remoteID); err != nil {
		h.failProvision(&order, &inst, &p, err)
		return
	}
	ip, _ := prov.IP(ctx, remoteID)

	h.DB.Model(&inst).Updates(map[string]interface{}{
		"remote_id": remoteID, "ip": ip, "status": "running",
	})
	h.DB.Model(&order).Updates(map[string]interface{}{
		"status": "active", "instance_id": inst.ID,
	})
}

func (h *Handler) failProvision(order *model.Order, inst *model.Instance, p *model.Product, cause error) {
	if inst != nil && inst.ID > 0 {
		h.DB.Model(&model.Instance{}).Where("id = ?", inst.ID).Update("status", "error")
	}
	h.DB.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
		"status": "failed", "remark": cause.Error(),
	})
	// 退款
	_ = h.DB.Transaction(func(tx *gorm.DB) error {
		_, e := applyBalanceTx(tx, order.UserID, order.Amount, "refund", "开通失败退款 "+order.OrderNo)
		return e
	})
	// 回补库存（仅有限库存产品）
	if p != nil && p.Stock >= 0 {
		h.DB.Model(&model.Product{}).Where("id = ?", p.ID).
			UpdateColumn("stock", gorm.Expr("stock + ?", order.Quantity))
	}
}

// envStr / envInt：读取环境变量，缺省回退。
func envStr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
