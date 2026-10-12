package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"onehost/internal/model"
	"onehost/internal/util"
)

type redeemReq struct {
	Code string `json:"code"`
}

// RedeemVoucher 用户兑换代金券到余额（事务 + 行锁，杜绝并发重复兑换）。
func (h *Handler) RedeemVoucher(c *gin.Context) {
	var req redeemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, "参数错误")
		return
	}
	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	if req.Code == "" {
		util.Fail(c, "请输入券码")
		return
	}

	uid := UID(c)
	var amount int64
	var newBal int64

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var v model.Voucher
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("code = ?", req.Code).First(&v).Error; err != nil {
			return fmt.Errorf("代金券不存在")
		}
		if v.Status != 0 {
			return fmt.Errorf("代金券已被使用或作废")
		}
		if v.ExpiredAt != nil && time.Now().After(*v.ExpiredAt) {
			return fmt.Errorf("代金券已过期")
		}
		now := time.Now()
		if err := tx.Model(&v).Updates(map[string]interface{}{
			"status": 1, "used_by": uid, "used_at": now,
		}).Error; err != nil {
			return err
		}
		bal, err := applyBalanceTx(tx, uid, v.Amount, "voucher", "代金券兑换 "+v.Code)
		if err != nil {
			return err
		}
		amount = v.Amount
		newBal = bal
		return nil
	})
	if err != nil {
		util.Fail(c, err.Error())
		return
	}
	util.OK(c, gin.H{"amount": amount, "balance": newBal})
}
