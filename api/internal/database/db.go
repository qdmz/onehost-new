package database

import (
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"onehost/internal/config"
	"onehost/internal/model"
	"onehost/internal/util"
)

// open 根据配置选择数据库方言：mysql（默认）或 sqlite（本地/测试）。
func open(cfg *config.Config) (*gorm.DB, error) {
	gcfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)}
	if cfg.Database.Driver == "sqlite" {
		path := cfg.Database.Path
		if path == "" {
			path = "onehost.db"
		}
		return gorm.Open(sqlite.Open(path), gcfg)
	}
	return gorm.Open(mysql.Open(cfg.DSN()), gcfg)
}

// Init 连接数据库（带重试）、自动迁移并写入种子数据。
func Init(cfg *config.Config) *gorm.DB {
	var db *gorm.DB
	var err error

	for i := 0; i < 30; i++ {
		db, err = open(cfg)
		if err == nil {
			if sqlDB, e := db.DB(); e == nil {
				err = sqlDB.Ping()
			} else {
				err = e
			}
			if err == nil {
				break
			}
		}
		log.Printf("waiting for database... (%d/30): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("database connect failed: %v", err)
	}

	if err := db.AutoMigrate(
		&model.User{}, &model.BalanceLog{}, &model.Product{}, &model.Order{},
		&model.Instance{}, &model.Node{}, &model.Voucher{}, &model.RechargeOrder{},
		&model.Setting{},
	); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	seed(db, cfg)
	return db
}

func seed(db *gorm.DB, cfg *config.Config) {
	// 管理员账号
	var admins int64
	db.Model(&model.User{}).Where("role = ?", "admin").Count(&admins)
	if admins == 0 {
		hash, _ := util.HashPassword("admin123")
		db.Create(&model.User{
			Email: "admin@onehost.local", Password: hash,
			Nickname: "超级管理员", Role: "admin", Status: 1,
		})
		log.Println("[seed] 已创建管理员账号: admin@onehost.local / admin123")
	}

	// 系统配置默认值
	defaults := map[string]string{
		"site_name":        cfg.App.SiteName,
		"register_enabled": "true",
		"yipay_enabled":    "false",
		"yipay_gateway":    "",
		"yipay_pid":        "",
		"yipay_key":        "",
	}
	for k, v := range defaults {
		var s model.Setting
		if err := db.Where("config_key = ?", k).First(&s).Error; err != nil {
			db.Create(&model.Setting{Key: k, Value: v})
		}
	}

	// 示例节点
	var nodes int64
	db.Model(&model.Node{}).Count(&nodes)
	if nodes == 0 {
		db.Create(&model.Node{Name: "本地 Mock 节点", Type: "mock", Status: 1})
	}

	// 示例产品
	var products int64
	db.Model(&model.Product{}).Count(&products)
	if products == 0 {
		db.Create(&model.Product{
			Name: "1核1G 入门型", Description: "适合个人测试、轻量站点",
			Price: 990, PeriodDays: 30, CPU: 1, MemoryMB: 1024, DiskGB: 10, TrafficGB: 500,
			Image: "images:debian/12", Stock: 100, MaxPerUser: 3, NodeID: 1, Status: 1,
		})
		db.Create(&model.Product{
			Name: "2核2G 标准型", Description: "适合中小型网站与应用",
			Price: 2990, PeriodDays: 30, CPU: 2, MemoryMB: 2048, DiskGB: 20, TrafficGB: 1000,
			Image: "images:debian/12", Stock: 50, MaxPerUser: 3, NodeID: 1, Status: 1,
		})
	}
}
