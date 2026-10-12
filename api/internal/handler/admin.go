package handler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"onehost/internal/model"
	"onehost/internal/provider"
	"onehost/internal/util"
)

// ---------------- 仪表盘 ----------------

func (h *Handler) AdminStats(c *gin.Context) {
	var users, orders, instances, nodes int64
	var revenue int64
	h.DB.Model(&model.User{}).Where("role = ?", "user").Count(&users)
	h.DB.Model(&model.Order{}).Count(&orders)
	h.DB.Model(&model.Instance{}).Where("status <> ?", "deleted").Count(&instances)
	h.DB.Model(&model.Node{}).Count(&nodes)
	h.DB.Model(&model.Order{}).
		Where("status IN ?", []string{"paid", "provisioning", "active"}).
		Select("COALESCE(SUM(amount),0)").Scan(&revenue)

	util.OK(c, gin.H{
		"users":     users,
		"orders":    orders,
		"instances": instances,
		"nodes":     nodes,
		"revenue":   revenue,
	})
}

// ---------------- 产品管理 ----------------

func (h *Handler) AdminListProducts(c *gin.Context) {
	var list []model.Product
	h.DB.Order("id desc").Find(&list)
	util.OK(c, list)
}

func (h *Handler) AdminCreateProduct(c *gin.Context) {
	var p model.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		util.Fail(c, "参数错误")
		return
	}
	p.ID = 0
	if err := h.DB.Create(&p).Error; err != nil {
		util.Fail(c, "创建失败")
		return
	}
	util.OK(c, p)
}

func (h *Handler) AdminUpdateProduct(c *gin.Context) {
	var p model.Product
	if err := h.DB.First(&p, c.Param("id")).Error; err != nil {
		util.Fail(c, "产品不存在")
		return
	}
	var req model.Product
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, "参数错误")
		return
	}
	req.ID = p.ID
	if err := h.DB.Model(&p).Select("*").Omit("id", "created_at", "updated_at").Updates(&req).Error; err != nil {
		util.Fail(c, "更新失败")
		return
	}
	util.OK(c, req)
}

func (h *Handler) AdminDeleteProduct(c *gin.Context) {
	h.DB.Delete(&model.Product{}, c.Param("id"))
	util.OK(c, nil)
}

// ---------------- 订单管理 ----------------

func (h *Handler) AdminListOrders(c *gin.Context) {
	var orders []model.Order
	q := h.DB.Model(&model.Order{})
	if s := c.Query("status"); s != "" {
		q = q.Where("status = ?", s)
	}
	if u := c.Query("user_id"); u != "" {
		q = q.Where("user_id = ?", u)
	}
	q.Order("id desc").Limit(500).Find(&orders)

	emailMap := map[uint]string{}
	var users []model.User
	h.DB.Find(&users)
	for _, u := range users {
		emailMap[u.ID] = u.Email
	}
	rows := make([]gin.H, 0, len(orders))
	for _, o := range orders {
		rows = append(rows, gin.H{"order": o, "user_email": emailMap[o.UserID]})
	}
	util.OK(c, rows)
}

func (h *Handler) AdminListRechargeOrders(c *gin.Context) {
	var list []model.RechargeOrder
	h.DB.Order("id desc").Limit(500).Find(&list)
	util.OK(c, list)
}

func (h *Handler) AdminListInstances(c *gin.Context) {
	var list []model.Instance
	h.DB.Where("status <> ?", "deleted").Order("id desc").Limit(500).Find(&list)
	util.OK(c, list)
}

// ---------------- 用户管理 ----------------

func (h *Handler) AdminListUsers(c *gin.Context) {
	var list []model.User
	h.DB.Order("id desc").Limit(500).Find(&list)
	util.OK(c, list)
}

func (h *Handler) AdminUpdateUserStatus(c *gin.Context) {
	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, "参数错误")
		return
	}
	h.DB.Model(&model.User{}).Where("id = ?", c.Param("id")).Update("status", req.Status)
	util.OK(c, nil)
}

func (h *Handler) AdminAdjustBalance(c *gin.Context) {
	var req struct {
		Change int64  `json:"change"` // 分，可正可负
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Change == 0 {
		util.Fail(c, "参数错误")
		return
	}
	id64, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	uid := uint(id64)
	if req.Remark == "" {
		req.Remark = "管理员调整"
	}
	var newBal int64
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		bal, err := applyBalanceTx(tx, uid, req.Change, "admin", req.Remark)
		newBal = bal
		return err
	})
	if err != nil {
		util.Fail(c, err.Error())
		return
	}
	util.OK(c, gin.H{"balance": newBal})
}

// ---------------- 节点管理 ----------------

func (h *Handler) AdminListNodes(c *gin.Context) {
	var list []model.Node
	h.DB.Order("id desc").Find(&list)
	rows := make([]gin.H, 0, len(list))
	for _, n := range list {
		rows = append(rows, gin.H{
			"id": n.ID, "name": n.Name, "type": n.Type, "api_url": n.APIURL,
			"ip_pool": n.IPPool, "status": n.Status,
			"has_cert": n.Cert != "", "created_at": n.CreatedAt,
		})
	}
	util.OK(c, rows)
}

type nodeReq struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	APIURL     string `json:"api_url"`
	Cert       string `json:"cert"`
	Key        string `json:"key"`
	ServerCert string `json:"server_cert"`
	IPPool     string `json:"ip_pool"`
	Status     int    `json:"status"`
}

func (h *Handler) AdminCreateNode(c *gin.Context) {
	var req nodeReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		util.Fail(c, "参数错误")
		return
	}
	if req.Type == "" {
		req.Type = "mock"
	}
	n := model.Node{
		Name: req.Name, Type: req.Type, APIURL: req.APIURL,
		Cert: req.Cert, Key: req.Key, ServerCert: req.ServerCert,
		IPPool: req.IPPool, Status: req.Status,
	}
	if err := h.DB.Create(&n).Error; err != nil {
		util.Fail(c, "创建失败")
		return
	}
	util.OK(c, n)
}

func (h *Handler) AdminUpdateNode(c *gin.Context) {
	var n model.Node
	if err := h.DB.First(&n, c.Param("id")).Error; err != nil {
		util.Fail(c, "节点不存在")
		return
	}
	var req nodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, "参数错误")
		return
	}
	n.Name = req.Name
	if req.Type != "" {
		n.Type = req.Type
	}
	n.APIURL = req.APIURL
	n.IPPool = req.IPPool
	n.Status = req.Status
	if req.Cert != "" {
		n.Cert = req.Cert
	}
	if req.Key != "" {
		n.Key = req.Key
	}
	if req.ServerCert != "" {
		n.ServerCert = req.ServerCert
	}
	if err := h.DB.Save(&n).Error; err != nil {
		util.Fail(c, "更新失败")
		return
	}
	util.OK(c, nil)
}

func (h *Handler) AdminDeleteNode(c *gin.Context) {
	h.DB.Delete(&model.Node{}, c.Param("id"))
	util.OK(c, nil)
}

func (h *Handler) AdminTestNode(c *gin.Context) {
	var n model.Node
	if err := h.DB.First(&n, c.Param("id")).Error; err != nil {
		util.Fail(c, "节点不存在")
		return
	}
	prov, err := provider.ForNode(&n)
	if err != nil {
		util.Fail(c, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := prov.Ping(ctx); err != nil {
		util.Fail(c, "连接失败: "+err.Error())
		return
	}
	util.OK(c, gin.H{"message": fmt.Sprintf("节点 %s 连接正常", n.Name)})
}

// ---------------- 代金券管理 ----------------

func (h *Handler) AdminListVouchers(c *gin.Context) {
	var list []model.Voucher
	q := h.DB.Model(&model.Voucher{})
	if s := c.Query("status"); s != "" {
		q = q.Where("status = ?", s)
	}
	q.Order("id desc").Limit(1000).Find(&list)
	util.OK(c, list)
}

func (h *Handler) AdminGenerateVouchers(c *gin.Context) {
	var req struct {
		Amount    int64  `json:"amount"`
		Count     int    `json:"count"`
		ExpiredAt string `json:"expired_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 || req.Count <= 0 || req.Count > 1000 {
		util.Fail(c, "参数错误（面额>0，数量 1~1000）")
		return
	}
	var expired *time.Time
	if req.ExpiredAt != "" {
		if t, err := time.ParseInLocation("2006-01-02", req.ExpiredAt, time.Local); err == nil {
			expired = &t
		} else if t, err := time.Parse(time.RFC3339, req.ExpiredAt); err == nil {
			expired = &t
		}
	}
	batch := util.RandCode(8)
	created := make([]model.Voucher, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		code := "OH" + util.RandCode(14)
		v := model.Voucher{
			Code: code, Amount: req.Amount, Status: 0,
			ExpiredAt: expired, Batch: batch,
		}
		if err := h.DB.Create(&v).Error; err == nil {
			created = append(created, v)
		}
	}
	util.OK(c, gin.H{"batch": batch, "count": len(created), "items": created})
}

func (h *Handler) AdminVoidVoucher(c *gin.Context) {
	var v model.Voucher
	if err := h.DB.First(&v, c.Param("id")).Error; err != nil {
		util.Fail(c, "代金券不存在")
		return
	}
	if v.Status == 1 {
		util.Fail(c, "已使用的代金券无法作废")
		return
	}
	h.DB.Model(&v).Update("status", 2)
	util.OK(c, nil)
}

// ---------------- 系统设置 ----------------

var settingWhitelist = map[string]bool{
	"site_name":        true,
	"register_enabled": true,
	"yipay_enabled":    true,
	"yipay_gateway":    true,
	"yipay_pid":        true,
	"yipay_key":        true,
}

func (h *Handler) AdminGetSettings(c *gin.Context) {
	var list []model.Setting
	h.DB.Find(&list)
	m := map[string]string{}
	for _, s := range list {
		m[s.Key] = s.Value
	}
	util.OK(c, m)
}

func (h *Handler) AdminUpdateSettings(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, "参数错误")
		return
	}
	for k, v := range req {
		if !settingWhitelist[k] {
			continue
		}
		if err := h.SetSetting(k, v); err != nil {
			util.Fail(c, "保存失败: "+k)
			return
		}
	}
	util.OK(c, nil)
}
