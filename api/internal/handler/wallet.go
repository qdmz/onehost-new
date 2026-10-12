package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"onehost/internal/model"
	"onehost/internal/service"
	"onehost/internal/util"
)

func (h *Handler) yipay() *service.Yipay {
	return &service.Yipay{
		Gateway: h.GetSetting("yipay_gateway"),
		PID:     h.GetSetting("yipay_pid"),
		Key:     h.GetSetting("yipay_key"),
	}
}

func (h *Handler) yipayEnabled() bool {
	return h.GetSetting("yipay_enabled") == "true" &&
		h.GetSetting("yipay_pid") != "" &&
		h.GetSetting("yipay_key") != "" &&
		h.GetSetting("yipay_gateway") != ""
}

// applyBalanceTx 在事务内变更余额并写入流水（行锁，避免并发超扣/重复入账）。
func applyBalanceTx(tx *gorm.DB, userID uint, change int64, logType, remark string) (int64, error) {
	var u model.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&u, userID).Error; err != nil {
		return 0, fmt.Errorf("用户不存在")
	}
	newBal := u.Balance + change
	if newBal < 0 {
		return u.Balance, fmt.Errorf("余额不足")
	}
	if err := tx.Model(&u).Update("balance", newBal).Error; err != nil {
		return 0, err
	}
	if err := tx.Create(&model.BalanceLog{
		UserID: userID, Change: change, BalanceAfter: newBal, Type: logType, Remark: remark,
	}).Error; err != nil {
		return 0, err
	}
	return newBal, nil
}

// BalanceLogs 余额流水。
func (h *Handler) BalanceLogs(c *gin.Context) {
	var logs []model.BalanceLog
	h.DB.Where("user_id = ?", UID(c)).Order("id desc").Limit(200).Find(&logs)
	util.OK(c, logs)
}

type rechargeReq struct {
	Amount int64  `json:"amount"` // 单位：分
	Method string `json:"method"`
}

// CreateRecharge 创建充值订单并返回支付地址。
func (h *Handler) CreateRecharge(c *gin.Context) {
	var req rechargeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, "参数错误")
		return
	}
	if req.Amount < 100 || req.Amount > 10000000 {
		util.Fail(c, "充值金额需在 1 元 ~ 10 万元之间")
		return
	}
	if req.Method != "alipay" && req.Method != "wxpay" {
		req.Method = "alipay"
	}

	ro := model.RechargeOrder{
		OrderNo: util.GenOrderNo("R"),
		UserID:  UID(c),
		Amount:  req.Amount,
		Method:  req.Method,
		Status:  "unpaid",
	}
	if err := h.DB.Create(&ro).Error; err != nil {
		util.Fail(c, "创建充值订单失败")
		return
	}

	resp := gin.H{"order": ro, "pay_url": "", "mock": !h.yipayEnabled()}
	if h.yipayEnabled() {
		y := h.yipay()
		notifyURL := h.Cfg.App.PublicURL + "/api/v1/pay/notify"
		returnURL := h.Cfg.App.PublicURL + "/api/v1/pay/return"
		resp["pay_url"] = y.BuildPayURL(ro.OrderNo, "余额充值", util.FormatMoney(ro.Amount), req.Method, notifyURL, returnURL)
	}
	util.OK(c, resp)
}

// RechargeOrders 充值订单列表。
func (h *Handler) RechargeOrders(c *gin.Context) {
	var list []model.RechargeOrder
	h.DB.Where("user_id = ?", UID(c)).Order("id desc").Limit(200).Find(&list)
	util.OK(c, list)
}

// MockPay 未接入真实支付网关时的本地模拟支付（仅用于开发联调）。
func (h *Handler) MockPay(c *gin.Context) {
	if h.yipayEnabled() {
		util.Fail(c, "已启用真实支付网关，禁止模拟支付")
		return
	}
	uid := UID(c)
	var ro model.RechargeOrder
	if err := h.DB.Where("id = ? AND user_id = ?", c.Param("id"), uid).First(&ro).Error; err != nil {
		util.Fail(c, "充值订单不存在")
		return
	}
	if ro.Status == "paid" {
		util.Fail(c, "该订单已支付")
		return
	}
	now := time.Now()
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&ro).Updates(map[string]interface{}{
			"status": "paid", "paid_at": now, "trade_no": "MOCK-" + ro.OrderNo,
		}).Error; err != nil {
			return err
		}
		_, err := applyBalanceTx(tx, uid, ro.Amount, "recharge", "模拟充值 "+util.FormatMoney(ro.Amount)+" 元")
		return err
	})
	if err != nil {
		util.Fail(c, err.Error())
		return
	}
	util.OK(c, gin.H{"order_no": ro.OrderNo})
}

func collectParams(c *gin.Context) map[string]string {
	_ = c.Request.ParseForm()
	m := map[string]string{}
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	for k, v := range c.Request.Form {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	return m
}

// PayNotify 易支付异步回调。
func (h *Handler) PayNotify(c *gin.Context) {
	params := collectParams(c)
	y := h.yipay()
	if !y.VerifyNotify(params) {
		c.String(200, "sign error")
		return
	}
	if ts := params["trade_status"]; ts != "" && ts != "TRADE_SUCCESS" {
		c.String(200, "success")
		return
	}
	var ro model.RechargeOrder
	if err := h.DB.Where("order_no = ?", params["out_trade_no"]).First(&ro).Error; err != nil {
		c.String(200, "order not found")
		return
	}
	if ro.Status == "paid" {
		c.String(200, "success")
		return
	}
	now := time.Now()
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&ro).Updates(map[string]interface{}{
			"status": "paid", "paid_at": now, "trade_no": params["trade_no"],
		}).Error; err != nil {
			return err
		}
		_, err := applyBalanceTx(tx, ro.UserID, ro.Amount, "recharge", "在线充值 "+util.FormatMoney(ro.Amount)+" 元")
		return err
	})
	if err != nil {
		c.String(200, "fail")
		return
	}
	c.String(200, "success")
}

// PayReturn 同步回跳。
func (h *Handler) PayReturn(c *gin.Context) {
	c.Redirect(302, h.Cfg.App.PublicURL+"/#/console/wallet")
}
