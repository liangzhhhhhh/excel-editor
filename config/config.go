// Package config
// @description 环境配置管理
// @author      梁志豪
// @datetime    2026/01/21
package config

import "os"

// Config 应用配置
type Config struct {
	BaseURL string // API 基础 URL
	AuthURL string // 认证服务 URL
	OaURL   string // OA 服务 URL
}

var cfg *Config

// GetConfig 获取当前环境配置
func GetConfig() *Config {
	if cfg == nil {
		cfg = loadConfig()
	}
	return cfg
}

// loadConfig 加载配置（优先使用环境变量，否则根据 build tags 或运行模式选择不同环境）
func loadConfig() *Config {
	// 优先使用环境变量
	if baseURL := os.Getenv("API_BASE_URL"); baseURL != "" {
		return &Config{
			BaseURL: baseURL,
			AuthURL: getEnvOrDefault("API_AUTH_URL", "http://wechat.aaagame.com"),
			OaURL:   getEnvOrDefault("API_OA_URL", "https://fatcat-admin-test.54030.com"),
		}
	}

	// 检查是否是开发模式（通过环境变量或 build tags）
	if isDevMode() {
		return getDevConfig()
	}

	// 否则使用生产环境配置
	return getProdConfig()
}

// isDevMode 判断是否是开发模式
func isDevMode() bool {
	// 方式1: 检查环境变量
	if env := os.Getenv("WAILS_ENV"); env == "dev" || env == "development" {
		return true
	}

	// 方式2: 检查 build tags（通过 getDefaultConfig 函数判断）
	// 如果 getDefaultConfig 返回的是开发配置，说明编译时使用了 dev tag
	devCfg := getDevConfig()
	defaultCfg := getDefaultConfig()
	return defaultCfg.BaseURL == devCfg.BaseURL
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDevConfig 开发环境配置
func getDevConfig() *Config {
	return &Config{
		BaseURL: "http://127.0.0.1:3000",
		AuthURL: "http://wechat.aaagame.com",
		OaURL:   "https://fatcat-admin-test.54030.com",
	}
}

// getProdConfig 生产环境配置
func getProdConfig() *Config {
	return &Config{
		BaseURL: "http://192.168.1.6:3001",
		AuthURL: "http://wechat.aaagame.com",
		OaURL:   "https://fatcat-admin-test.54030.com",
	}
}
