package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Mode string `yaml:"mode"`
	} `yaml:"server"`
	Database struct {
		Driver   string `yaml:"driver"` // mysql（默认）/ sqlite
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Name     string `yaml:"name"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Path     string `yaml:"path"` // sqlite 文件路径
	} `yaml:"database"`
	JWT struct {
		Secret      string `yaml:"secret"`
		ExpireHours int    `yaml:"expire_hours"`
	} `yaml:"jwt"`
	App struct {
		PublicURL string `yaml:"public_url"`
		SiteName  string `yaml:"site_name"`
	} `yaml:"app"`
}

// Load 读取 config.yaml，并用 ONEHOST_* 环境变量覆盖（环境变量优先）。
func Load(path string) *Config {
	cfg := &Config{}
	if b, err := os.ReadFile(path); err == nil {
		_ = yaml.Unmarshal(b, cfg)
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8888
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "release"
	}
	if cfg.Database.Driver == "" {
		cfg.Database.Driver = "mysql"
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 3306
	}
	if cfg.JWT.ExpireHours == 0 {
		cfg.JWT.ExpireHours = 168
	}
	if cfg.App.SiteName == "" {
		cfg.App.SiteName = "OneHost"
	}
	if cfg.App.PublicURL == "" {
		cfg.App.PublicURL = "http://localhost"
	}

	envStr("ONEHOST_DB_DRIVER", &cfg.Database.Driver)
	envStr("ONEHOST_DB_PATH", &cfg.Database.Path)
	envStr("ONEHOST_DB_HOST", &cfg.Database.Host)
	envStr("ONEHOST_DB_NAME", &cfg.Database.Name)
	envStr("ONEHOST_DB_USER", &cfg.Database.User)
	envStr("ONEHOST_DB_PASSWORD", &cfg.Database.Password)
	envStr("ONEHOST_JWT_SECRET", &cfg.JWT.Secret)
	envStr("ONEHOST_PUBLIC_URL", &cfg.App.PublicURL)
	envStr("ONEHOST_SITE_NAME", &cfg.App.SiteName)
	envStr("ONEHOST_MODE", &cfg.Server.Mode)
	envInt("ONEHOST_DB_PORT", &cfg.Database.Port)
	envInt("ONEHOST_PORT", &cfg.Server.Port)

	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "change-me-in-production"
	}
	return cfg
}

// DSN 返回 MySQL 连接串。
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.Name)
}

func envStr(key string, dst *string) {
	if v := os.Getenv(key); v != "" {
		*dst = v
	}
}

func envInt(key string, dst *int) {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			*dst = i
		}
	}
}
