package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"onehost/internal/util"
)

// AuthRequired 校验 JWT，注入 uid / role。
func (h *Handler) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ""
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		}
		if token == "" {
			token = c.Query("token")
		}
		if token == "" {
			util.FailAbort(c, 401, "未登录")
			return
		}
		claims, err := util.ParseToken(h.Cfg.JWT.Secret, token)
		if err != nil {
			util.FailAbort(c, 401, "登录已过期，请重新登录")
			return
		}
		c.Set("uid", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// AdminRequired 要求管理员角色。
func (h *Handler) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			util.FailAbort(c, 403, "无管理员权限")
			return
		}
		c.Next()
	}
}

// CORS 跨域中间件。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// UID 从上下文取当前用户 ID。
func UID(c *gin.Context) uint {
	v, _ := c.Get("uid")
	if id, ok := v.(uint); ok {
		return id
	}
	return 0
}
