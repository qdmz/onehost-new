package handler

import (
	"github.com/gin-gonic/gin"

	"onehost/internal/model"
	"onehost/internal/util"
)

// ListProducts 商城产品列表（仅上架）。
func (h *Handler) ListProducts(c *gin.Context) {
	var list []model.Product
	h.DB.Where("status = ?", 1).Order("price asc").Find(&list)
	util.OK(c, list)
}

// GetProduct 产品详情。
func (h *Handler) GetProduct(c *gin.Context) {
	var p model.Product
	if err := h.DB.First(&p, c.Param("id")).Error; err != nil {
		util.Fail(c, "产品不存在")
		return
	}
	util.OK(c, p)
}
