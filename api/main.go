package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"onehost/internal/config"
	"onehost/internal/database"
	"onehost/internal/handler"
)

func main() {
	cfg := config.Load("config/config.yaml")

	gin.SetMode(gin.ReleaseMode)
	if cfg.Server.Mode != "" {
		gin.SetMode(cfg.Server.Mode)
	}

	db := database.Init(cfg)
	h := handler.New(db, cfg)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), handler.CORS())
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	handler.RegisterRoutes(r, h, cfg)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("OneHost API 已启动，监听 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
