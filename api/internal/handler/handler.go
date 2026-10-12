package handler

import (
	"gorm.io/gorm"

	"onehost/internal/config"
	"onehost/internal/model"
)

type Handler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func New(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{DB: db, Cfg: cfg}
}

// GetSetting 读取系统配置。
func (h *Handler) GetSetting(key string) string {
	var s model.Setting
	if err := h.DB.Where("config_key = ?", key).First(&s).Error; err != nil {
		return ""
	}
	return s.Value
}

// SetSetting 写入（存在则更新）系统配置。
func (h *Handler) SetSetting(key, value string) error {
	var s model.Setting
	err := h.DB.Where("config_key = ?", key).First(&s).Error
	if err == gorm.ErrRecordNotFound {
		return h.DB.Create(&model.Setting{Key: key, Value: value}).Error
	}
	if err != nil {
		return err
	}
	return h.DB.Model(&s).Update("value", value).Error
}
