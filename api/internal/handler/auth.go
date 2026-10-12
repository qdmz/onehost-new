package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"onehost/internal/model"
	"onehost/internal/util"
)

// PublicSettings 暴露给前端的公开配置。
func (h *Handler) PublicSettings(c *gin.Context) {
	util.OK(c, gin.H{
		"site_name":        h.GetSetting("site_name"),
		"register_enabled": h.GetSetting("register_enabled") != "false",
		"yipay_enabled":    h.yipayEnabled(),
	})
}

type regReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

// Register 用户注册。
func (h *Handler) Register(c *gin.Context) {
	if h.GetSetting("register_enabled") == "false" {
		util.Fail(c, "注册功能已关闭")
		return
	}
	var req regReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, "参数错误")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		util.Fail(c, "邮箱格式不正确")
		return
	}
	if len(req.Password) < 6 {
		util.Fail(c, "密码至少 6 位")
		return
	}

	var cnt int64
	h.DB.Model(&model.User{}).Where("email = ?", req.Email).Count(&cnt)
	if cnt > 0 {
		util.Fail(c, "该邮箱已注册")
		return
	}

	hash, err := util.HashPassword(req.Password)
	if err != nil {
		util.Fail(c, "密码处理失败")
		return
	}
	nick := strings.TrimSpace(req.Nickname)
	if nick == "" {
		nick = strings.Split(req.Email, "@")[0]
	}
	u := model.User{Email: req.Email, Password: hash, Nickname: nick, Role: "user", Status: 1}
	if err := h.DB.Create(&u).Error; err != nil {
		util.Fail(c, "注册失败")
		return
	}
	token, _ := util.GenerateToken(h.Cfg.JWT.Secret, u.ID, u.Role, h.Cfg.JWT.ExpireHours)
	util.OK(c, gin.H{"token": token, "user": u})
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login 用户登录。
func (h *Handler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, "参数错误")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	var u model.User
	if err := h.DB.Where("email = ?", req.Email).First(&u).Error; err != nil {
		util.Fail(c, "邮箱或密码错误")
		return
	}
	if !util.CheckPassword(u.Password, req.Password) {
		util.Fail(c, "邮箱或密码错误")
		return
	}
	if u.Status != 1 {
		util.Fail(c, "账号已被禁用")
		return
	}
	token, _ := util.GenerateToken(h.Cfg.JWT.Secret, u.ID, u.Role, h.Cfg.JWT.ExpireHours)
	util.OK(c, gin.H{"token": token, "user": u})
}

// Profile 当前用户信息。
func (h *Handler) Profile(c *gin.Context) {
	var u model.User
	if err := h.DB.First(&u, UID(c)).Error; err != nil {
		util.Fail(c, "用户不存在")
		return
	}
	util.OK(c, u)
}
